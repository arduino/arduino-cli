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
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
)

type AlreadyInstalledError struct{}

func (e *AlreadyInstalledError) Error() string {
	return tr("library already installed")
}

var (
	// ErrAlreadyInstalled is returned when a library is already installed and task
	// cannot proceed.
	ErrAlreadyInstalled = &AlreadyInstalledError{}
)

// InstallPrerequisiteCheck performs prequisite checks to install a library. It returns the
// install path, where the library should be installed and the possible library that is already
// installed on the same folder and it's going to be replaced by the new one.
func (lm *LibrariesManager) InstallPrerequisiteCheck(indexLibrary *librariesindex.Release) (*paths.Path, *libraries.Library, error) {
	saneName := utils.SanitizeName(indexLibrary.Library.Name)

	var replaced *libraries.Library
	if installedLibs, have := lm.Libraries[saneName]; have {
		for _, installedLib := range installedLibs.Alternatives {
			if installedLib.Location != libraries.User {
				continue
			}
			if installedLib.Version != nil && installedLib.Version.Equal(indexLibrary.Version) {
				return installedLib.InstallDir, nil, ErrAlreadyInstalled
			}
			replaced = installedLib
		}
	}

	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return nil, nil, fmt.Errorf(tr("User directory not set"))
	}

	libPath := libsDir.Join(saneName)
	if replaced != nil && replaced.InstallDir.EquivalentTo(libPath) {

	} else if libPath.IsDir() {
		return nil, nil, fmt.Errorf(tr("destination dir %s already exists, cannot install"), libPath)
	}
	return libPath, replaced, nil
}

// Install installs a library on the specified path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release, libPath *paths.Path) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf(tr("User directory not set"))
	}
	return indexLibrary.Resource.Install(lm.DownloadsDir, libsDir, libPath)
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if lib == nil || lib.InstallDir == nil {
		return fmt.Errorf(tr("install directory not set"))
	}
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf(tr("removing lib directory: %s"), err)
	}

	lm.Libraries[lib.Name].Remove(lib)
	return nil
}

//InstallZipLib  installs a Zip library on the specified path.
func (lm *LibrariesManager) InstallZipLib(ctx context.Context, archivePath string, overwrite bool) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf(tr("User directory not set"))
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
		return fmt.Errorf(tr("extracting archive: %w"), err)
	}

	paths, err := tmpDir.ReadDir()
	if err != nil {
		return err
	}

	// Ignores metadata from Mac OS X
	paths.FilterOutPrefix("__MACOSX")

	if len(paths) > 1 {
		return fmt.Errorf(tr("archive is not valid: multiple files found in zip file top level"))
	}

	extractionPath := paths[0]
	libraryName := extractionPath.Base()

	if err := validateLibrary(libraryName, extractionPath); err != nil {
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
			return fmt.Errorf(tr("library %s already installed"), libraryName)
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
		return fmt.Errorf(tr("moving extracted archive to destination dir: %s"), err)
	}

	return nil
}

//InstallGitLib  installs a library hosted on a git repository on the specified path.
func (lm *LibrariesManager) InstallGitLib(gitURL string, overwrite bool) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf(tr("User directory not set"))
	}

	libraryName, err := parseGitURL(gitURL)
	if err != nil {
		logrus.
			WithError(err).
			Warn("Parsing git URL")
		return err
	}

	installPath := libsDir.Join(libraryName)

	// Deletes libraries folder if already installed
	if _, ok := lm.Libraries[libraryName]; ok {
		if !overwrite {
			return fmt.Errorf(tr("library %s already installed"), libraryName)
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
		WithField("git url", gitURL).
		Trace("Installing library")

	_, err = git.PlainClone(installPath.String(), false, &git.CloneOptions{
		URL:      gitURL,
		Depth:    1,
		Progress: os.Stdout,
	})
	if err != nil {
		logrus.
			WithError(err).
			Warn("Cloning git repository")
		return err
	}

	if err := validateLibrary(libraryName, installPath); err != nil {
		// Clean up installation directory since this is not a valid library
		installPath.RemoveAll()
		return err
	}

	// We don't want the installed library to be a git repository thus we delete this folder
	installPath.Join(".git").RemoveAll()
	return nil
}

func parseGitURL(gitURL string) (string, error) {
	var res string
	if strings.HasPrefix(gitURL, "git@") {
		// We can't parse these as URLs
		i := strings.LastIndex(gitURL, "/")
		res = strings.TrimRight(gitURL[i+1:], ".git")
	} else if path := paths.New(gitURL); path.Exist() {
		res = path.Base()
	} else if parsed, err := url.Parse(gitURL); err == nil {
		i := strings.LastIndex(parsed.Path, "/")
		res = strings.TrimRight(parsed.Path[i+1:], ".git")
	} else {
		return "", fmt.Errorf(tr("invalid git url"))
	}
	return res, nil
}

// validateLibrary verifies the dir contains a valid library, meaning it has both
// an header <name>.h, either in src or root folder, and a library.properties file
func validateLibrary(name string, dir *paths.Path) error {
	// Verify library contains library header.
	// Checks also root folder for legacy reasons.
	// For more info see:
	// https://arduino.github.io/arduino-cli/latest/library-specification/#source-code
	libraryHeader := name + ".h"
	if !dir.Join("src", libraryHeader).Exist() && !dir.Join(libraryHeader).Exist() {
		return fmt.Errorf(tr(`library is not valid: missing header file "%s"`), libraryHeader)
	}

	// Verifies library contains library.properties
	if !dir.Join("library.properties").Exist() {
		return fmt.Errorf(tr(`library is not valid: missing file "library.properties"`))
	}

	return nil
}
