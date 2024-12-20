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

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/utils"
	"github.com/arduino/arduino-cli/internal/i18n"
	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v4"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	semver "go.bug.st/relaxed-semver"
)

// LibraryInstallPlan contains the main information required to perform a library
// install, like the path where the library should be installed and the library
// that is going to be replaced by the new one.
// This is the result of a call to InstallPrerequisiteCheck.
type LibraryInstallPlan struct {
	// Name of the library to install
	Name string

	// Version of the library to install
	Version *semver.Version

	// TargetPath is the path where the library should be installed.
	TargetPath *paths.Path

	// ReplacedLib is the library that is going to be replaced by the new one.
	ReplacedLib *libraries.Library

	// UpToDate is true if the library to install has the same version of the library we are going to replace.
	UpToDate bool
}

// InstallPrerequisiteCheck performs prequisite checks to install a library. It returns the
// install path, where the library should be installed and the possible library that is already
// installed on the same folder and it's going to be replaced by the new one.
func (lmi *Installer) InstallPrerequisiteCheck(name string, version *semver.Version, installLocation libraries.LibraryLocation) (*LibraryInstallPlan, error) {
	installDir, err := lmi.getLibrariesDir(installLocation)
	if err != nil {
		return nil, err
	}

	lmi.RescanLibraries()

	libs := lmi.FindByReference(name, nil, installLocation)
	if len(libs) > 1 {
		libsDir := paths.NewPathList()
		for _, lib := range libs {
			libsDir.Add(lib.InstallDir)
		}
		return nil, &cmderrors.MultipleLibraryInstallDetected{
			LibName: name,
			LibsDir: libsDir,
			Message: i18n.Tr("Automatic library install can't be performed in this case, please manually remove all duplicates and retry."),
		}
	}

	var replaced *libraries.Library
	var upToDate bool
	if len(libs) == 1 {
		lib := libs[0]
		replaced = lib
		upToDate = lib.Version != nil && lib.Version.Equal(version)
	}

	libPath := installDir.Join(utils.SanitizeName(name))
	if libPath.IsDir() {
		if replaced == nil || !replaced.InstallDir.EquivalentTo(libPath) {
			return nil, errors.New(i18n.Tr("destination dir %s already exists, cannot install", libPath))
		}
	}

	return &LibraryInstallPlan{
		Name:        name,
		Version:     version,
		TargetPath:  libPath,
		ReplacedLib: replaced,
		UpToDate:    upToDate,
	}, nil
}

// importLibraryFromDirectory installs a library by copying it from the given directory.
func (lmi *Installer) importLibraryFromDirectory(libPath *paths.Path, overwrite bool) error {
	// Check if the library is valid and load metatada
	if err := validateLibrary(libPath); err != nil {
		return err
	}
	library, err := libraries.Load(libPath, libraries.User)
	if err != nil {
		return err
	}

	// Check if the library is already installed and determine install path
	installPlan, err := lmi.InstallPrerequisiteCheck(library.Name, library.Version, libraries.User)
	if err != nil {
		return err
	}

	if installPlan.UpToDate {
		if !overwrite {
			return errors.New(i18n.Tr("library %s already installed", installPlan.Name))
		}
	}
	if installPlan.ReplacedLib != nil {
		if !overwrite {
			return errors.New(i18n.Tr("Library %[1]s is already installed, but with a different version: %[2]s", installPlan.Name, installPlan.ReplacedLib))
		}
		if err := lmi.Uninstall(installPlan.ReplacedLib); err != nil {
			return err
		}
	}
	if installPlan.TargetPath.Exist() {
		return fmt.Errorf("%s: %s", i18n.Tr("destination directory already exists"), installPlan.TargetPath)
	}
	if err := libPath.CopyDirTo(installPlan.TargetPath); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("copying library to destination directory:"), err)
	}
	return nil
}

// Uninstall removes a Library
func (lmi *Installer) Uninstall(lib *libraries.Library) error {
	if lib == nil || lib.InstallDir == nil {
		return errors.New(i18n.Tr("install directory not set"))
	}
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return errors.New(i18n.Tr("removing library directory: %s", err))
	}

	alternatives := lmi.libraries[lib.Name]
	alternatives.Remove(lib)
	lmi.libraries[lib.Name] = alternatives
	return nil
}

// InstallZipLib installs a Zip library on the specified path.
func (lmi *Installer) InstallZipLib(ctx context.Context, archivePath *paths.Path, overwrite bool) error {
	// Clone library in a temporary directory
	tmpDir, err := paths.MkTempDir("", "")
	if err != nil {
		return err
	}
	defer tmpDir.RemoveAll()

	file, err := archivePath.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Extract to a temporary directory so we can check if the zip is structured correctly.
	// We also use the top level folder from the archive to infer the library name.
	if err := extract.Archive(ctx, file, tmpDir.String(), nil); err != nil {
		return fmt.Errorf("%s: %w", i18n.Tr("extracting archive"), err)
	}

	libRootFiles, err := tmpDir.ReadDir()
	if err != nil {
		return err
	}
	libRootFiles.FilterOutPrefix("__MACOSX") // Ignores metadata from Mac OS X
	if len(libRootFiles) > 1 {
		return errors.New(i18n.Tr("archive is not valid: multiple files found in zip file top level"))
	}
	if len(libRootFiles) == 0 {
		return errors.New(i18n.Tr("archive is not valid: no files found in zip file top level"))
	}
	tmpInstallPath := libRootFiles[0]

	// Install extracted library in the destination directory
	if err := lmi.importLibraryFromDirectory(tmpInstallPath, overwrite); err != nil {
		return errors.New(i18n.Tr("moving extracted archive to destination dir: %s", err))
	}

	return nil
}

// InstallGitLib installs a library hosted on a git repository on the specified path.
func (lmi *Installer) InstallGitLib(argURL string, overwrite bool) error {
	libraryName, gitURL, ref, err := parseGitArgURL(argURL)
	if err != nil {
		return err
	}

	// Clone library in a temporary directory
	tmp, err := paths.MkTempDir("", "")
	if err != nil {
		return err
	}
	defer tmp.RemoveAll()
	tmpInstallPath := tmp.Join(libraryName)

	depth := 1
	if ref != "" {
		depth = 0
	}
	repo, err := git.PlainClone(tmpInstallPath.String(), false, &git.CloneOptions{
		URL:      gitURL,
		Depth:    depth,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	if ref != "" {
		if h, err := repo.ResolveRevision(ref); err != nil {
			return err
		} else if w, err := repo.Worktree(); err != nil {
			return err
		} else if err := w.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(h.String())}); err != nil {
			return err
		}
	}

	// We don't want the installed library to be a git repository thus we delete this folder
	tmpInstallPath.Join(".git").RemoveAll()

	// Install extracted library in the destination directory
	if err := lmi.importLibraryFromDirectory(tmpInstallPath, overwrite); err != nil {
		return errors.New(i18n.Tr("moving extracted archive to destination dir: %s", err))
	}

	return nil
}

// parseGitArgURL tries to recover a library name from a git URL.
// Returns an error in case the URL is not a valid git URL.
func parseGitArgURL(argURL string) (string, string, plumbing.Revision, error) {
	// On Windows handle paths with backslashes in the form C:\Path\to\library
	if path := paths.New(argURL); path != nil && path.Exist() {
		return path.Base(), argURL, "", nil
	}

	// Handle github-specific address in the form "git@github.com:arduino-libraries/SigFox.git"
	if strings.HasPrefix(argURL, "git@github.com:") {
		// We can't parse these as URLs
		argURL = "https://github.com/" + strings.TrimPrefix(argURL, "git@github.com:")
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
	rev := plumbing.Revision(parsedURL.Fragment)
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
		return errors.New(i18n.Tr("directory doesn't exist: %s", dir))
	}

	searchHeaderFile := func(d *paths.Path) (bool, error) {
		if d.NotExist() {
			// A directory that doesn't exist can't obviously contain any header file
			return false, nil
		}
		dirContent, err := d.ReadDir()
		if err != nil {
			return false, fmt.Errorf("%s: %w", i18n.Tr("reading directory %s content", dir), err)
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

	return errors.New(i18n.Tr("library not valid"))
}
