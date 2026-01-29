// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package librariesmanager

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/arduino/arduino-cli/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
)

type alreadyInstalledError struct{}

func (e *alreadyInstalledError) Error() string {
	return tr("library already installed")
}

var (
	// ErrAlreadyInstalled is returned when a library is already installed and task
	// cannot proceed.
	ErrAlreadyInstalled = &alreadyInstalledError{}
)

// InstallPrerequisiteCheck performs prequisite checks to install a library. It returns the
// install path, where the library should be installed and the possible library that is already
// installed on the same folder and it's going to be replaced by the new one.
func (lm *LibrariesManager) InstallPrerequisiteCheck(indexLibrary *librariesindex.Release, installLocation libraries.LibraryLocation) (*paths.Path, *libraries.Library, error) {
	installDir := lm.getLibrariesDir(installLocation)
	if installDir == nil {
		if installLocation == libraries.User {
			return nil, nil, fmt.Errorf("user directory not set")
		}
		return nil, nil, fmt.Errorf("builtin libraries directory not set")
	}

	name := indexLibrary.Library.Name
	libs := lm.FindByReference(&librariesindex.Reference{Name: name}, installLocation)
	for _, lib := range libs {
		if lib.Version != nil && lib.Version.Equal(indexLibrary.Version) {
			return lib.InstallDir, nil, ErrAlreadyInstalled
		}
	}

	if len(libs) > 1 {
		libsDir := paths.NewPathList()
		for _, lib := range libs {
			libsDir.Add(lib.InstallDir)
		}
		return nil, nil, &arduino.MultipleLibraryInstallDetected{
			LibName: name,
			LibsDir: libsDir,
			Message: tr("Automatic library install can't be performed in this case, please manually remove all duplicates and retry."),
		}
	}

	var replaced *libraries.Library
	if len(libs) == 1 {
		replaced = libs[0]
	}

	libPath := installDir.Join(utils.SanitizeName(indexLibrary.Library.Name))
	if replaced != nil && replaced.InstallDir.EquivalentTo(libPath) {
		return libPath, replaced, nil
	} else if libPath.IsDir() {
		return nil, nil, fmt.Errorf("destination dir %s already exists, cannot install", libPath)
	}
	return libPath, replaced, nil
}

// Install installs a library on the specified path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release, libPath *paths.Path) error {
	return indexLibrary.Resource.Install(lm.DownloadsDir, libPath.Parent(), libPath)
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if lib == nil || lib.InstallDir == nil {
		return fmt.Errorf("install directory not set")
	}
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing lib directory: %s", err)
	}

	alternatives := lm.Libraries[lib.Name]
	alternatives.Remove(lib)
	lm.Libraries[lib.Name] = alternatives
	return nil
}

// InstallZipLib installs a Zip library on the specified path.
func (lm *LibrariesManager) InstallZipLib(ctx context.Context, archivePath string, overwrite bool) error {
	libsDir := lm.getLibrariesDir(libraries.User)
	if libsDir == nil {
		return fmt.Errorf("user directory not set")
	}

	tmpDir, err := paths.MkTempDir(paths.TempDir().String(), "arduino-cli-lib-")
	if err != nil {
		return err
	}
	// Deletes temp dir used to extract archive when finished
	defer tmpDir.RemoveAll()

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract to a temporary directory so we can check if the zip is structured correctly.
	// We also use the top level folder from the archive to infer the library name.
	if err := extract.Archive(ctx, file, tmpDir.String(), nil); err != nil {
		return fmt.Errorf("extracting archive: %w", err)
	}

	paths, err := tmpDir.ReadDir()
	if err != nil {
		return err
	}

	// Ignores metadata from Mac OS X
	paths.FilterOutPrefix("__MACOSX")

	if len(paths) > 1 {
		return fmt.Errorf("archive is not valid: multiple files found in zip file top level")
	}

	extractionPath := paths[0]
	libraryName := extractionPath.Base()

	if err := validateLibrary(extractionPath); err != nil {
		return err
	}

	installPath := libsDir.Join(libraryName)

	if err := libsDir.MkdirAll(); err != nil {
		return err
	}
	defer func() {
		// Clean up install dir if installation failed
		files, err := installPath.ReadDir()
		if err == nil && len(files) == 0 {
			installPath.RemoveAll()
		}
	}()

	// Delete library folder if already installed
	if installPath.IsDir() {
		if !overwrite {
			return fmt.Errorf("library %s already installed", libraryName)
		}
		logrus.
			WithField("library name", libraryName).
			WithField("install path", installPath).
			Trace("Deleting library")
		installPath.RemoveAll()
	}

	logrus.
		WithField("library name", libraryName).
		WithField("install path", installPath).
		WithField("zip file", archivePath).
		Trace("Installing library")

	// Copy extracted library in the destination directory
	if err := extractionPath.CopyDirTo(installPath); err != nil {
		return fmt.Errorf("moving extracted archive to destination dir: %s", err)
	}

	return nil
}

// InstallGitLib installs a library hosted on a git repository on the specified path.
func (lm *LibrariesManager) InstallGitLib(gitURL string, overwrite bool) error {
	installDir := lm.getLibrariesDir(libraries.User)
	if installDir == nil {
		return fmt.Errorf("user directory not set")
	}

	libraryName, gitURL, ref, err := parseGitArgURL(gitURL)
	if err != nil {
		logrus.
			WithError(err).
			Warn("Parsing git URL")
		return err
	}

	// Deletes libraries folder if already installed
	installPath := installDir.Join(libraryName)
	if installPath.IsDir() {
		if !overwrite {
			return fmt.Errorf("library %s already installed", libraryName)
		}
		logrus.
			WithField("library name", libraryName).
			WithField("install path", installPath).
			Trace("Deleting library")
		installPath.RemoveAll()
	}
	if installPath.Exist() {
		return fmt.Errorf("could not create directory %s: a file with the same name exists", installPath)
	}

	logrus.
		WithField("library name", libraryName).
		WithField("install path", installPath).
		WithField("git url", gitURL).
		Trace("Installing library")

	_, err = git.PlainClone(installPath.String(), false, &git.CloneOptions{
		URL:           gitURL,
		ReferenceName: plumbing.ReferenceName(ref),
	})
	if err != nil {
		if err.Error() != "reference not found" {
			return err
		}

		// We did not find the requested reference, let's do a PlainClone and use
		// "ResolveRevision" to find and checkout the requested revision
		if repo, err := git.PlainClone(installPath.String(), false, &git.CloneOptions{
			URL: gitURL,
		}); err != nil {
			return err
		} else if h, err := repo.ResolveRevision(plumbing.Revision(ref)); err != nil {
			return err
		} else if w, err := repo.Worktree(); err != nil {
			return err
		} else if err := w.Checkout(&git.CheckoutOptions{
			Force: true, // workaround for: https://github.com/go-git/go-git/issues/1411
			Hash:  plumbing.NewHash(h.String())}); err != nil {
			return err
		}
	}

	fmt.Println("Validating library...")
	if err := validateLibrary(installPath); err != nil {
		// Clean up installation directory since this is not a valid library
		installPath.RemoveAll()
		return err
	}

	// We don't want the installed library to be a git repository thus we delete this folder
	installPath.Join(".git").RemoveAll()
	return nil
}

// parseGitURL tries to recover a library name from a git URL.
// Returns an error in case the URL is not a valid git URL.
func parseGitArgURL(argURL string) (string, string, string, error) {
	// On Windows handle paths with backslashes in the form C:\Path\to\library
	if path := paths.New(argURL); path != nil && path.Exist() {
		return path.Base(), argURL, "", nil
	}

	// Handle commercial git-specific address in the form "git@xxxxx.com:arduino-libraries/SigFox.git"
	prefixes := map[string]string{
		"git@github.com:":    "https://github.com/",
		"git@gitlab.com:":    "https://gitlab.com/",
		"git@bitbucket.org:": "https://bitbucket.org/",
	}
	for prefix, replacement := range prefixes {
		if strings.HasPrefix(argURL, prefix) {
			// We can't parse these as URLs
			argURL = replacement + strings.TrimPrefix(argURL, prefix)
		}
	}

	parsedURL, err := url.Parse(argURL)
	if err != nil {
		return "", "", "", fmt.Errorf("%s: %w", i18n.Tr("invalid git url"), err)
	}
	if parsedURL.String() == "" {
		return "", "", "", errors.New(i18n.Tr("invalid git url"))
	}

	// Extract lib name from "https://github.com/arduino-libraries/SigFox.git#1.0.3"
	// path == "/arduino-libraries/SigFox.git"
	slash := strings.LastIndex(parsedURL.Path, "/")
	if slash == -1 {
		return "", "", "", errors.New(i18n.Tr("invalid git url"))
	}
	libName := strings.TrimSuffix(parsedURL.Path[slash+1:], ".git")
	if libName == "" {
		return "", "", "", errors.New(i18n.Tr("invalid git url"))
	}
	// fragment == "1.0.3"
	rev := parsedURL.Fragment
	// gitURL == "https://github.com/arduino-libraries/SigFox.git"
	parsedURL.Fragment = ""
	gitURL := parsedURL.String()
	return libName, gitURL, rev, nil
}

// validateLibrary verifies the dir contains a valid library, meaning it has either
// library.properties file and an header in src/ or an header in its root folder.
// Returns nil if dir contains a valid library, error on all other cases.
func validateLibrary(dir *paths.Path) error {
	if dir.NotExist() {
		return fmt.Errorf("directory doesn't exist: %s", dir)
	}

	searchHeaderFile := func(d *paths.Path) (bool, error) {
		if d.NotExist() {
			// A directory that doesn't exist can't obviously contain any header file
			return false, nil
		}
		dirContent, err := d.ReadDir()
		if err != nil {
			return false, fmt.Errorf("reading directory %s content: %w", dir, err)
		}
		dirContent.FilterOutDirs()
		headerExtensions := []string{}
		for k := range globals.HeaderFilesValidExtensions {
			headerExtensions = append(headerExtensions, k)
		}
		dirContent.FilterSuffix(headerExtensions...)
		return len(dirContent) > 0, nil
	}

	// Recursive library layout
	// https://arduino.github.io/arduino-cli/latest/library-specification/#source-code
	if headerFound, err := searchHeaderFile(dir.Join("src")); err != nil {
		return err
	} else if dir.Join("library.properties").Exist() && headerFound {
		return nil
	}

	// Flat library layout
	// https://arduino.github.io/arduino-cli/latest/library-specification/#source-code
	if headerFound, err := searchHeaderFile(dir); err != nil {
		return err
	} else if headerFound {
		return nil
	}

	return fmt.Errorf("library not valid")
}
