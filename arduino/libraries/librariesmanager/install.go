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

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	paths "github.com/arduino/go-paths-helper"
	"github.com/codeclysm/extract/v3"
	"gopkg.in/src-d/go-git.v4"
)

var (
	// ErrAlreadyInstalled is returned when a library is already installed and task
	// cannot proceed.
	ErrAlreadyInstalled = errors.New("library already installed")
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
			if installedLib.Version.Equal(indexLibrary.Version) {
				return installedLib.InstallDir, nil, ErrAlreadyInstalled
			}
			replaced = installedLib
		}
	}

	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return nil, nil, fmt.Errorf("User directory not set")
	}

	libPath := libsDir.Join(saneName)
	if replaced != nil && replaced.InstallDir.EquivalentTo(libPath) {

	} else if libPath.IsDir() {
		return nil, nil, fmt.Errorf("destination dir %s already exists, cannot install", libPath)
	}
	return libPath, replaced, nil
}

// Install installs a library on the specified path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release, libPath *paths.Path) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("User directory not set")
	}
	return indexLibrary.Resource.Install(lm.DownloadsDir, libsDir, libPath)
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if lib == nil || lib.InstallDir == nil {
		return fmt.Errorf("install directory not set")
	}
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf("removing lib directory: %s", err)
	}

	lm.Libraries[lib.Name].Remove(lib)
	return nil
}

//InstallZipLib  installs a Zip library on the specified path.
func (lm *LibrariesManager) InstallZipLib(ctx context.Context, archivePath string) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("User directory not set")
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := extract.Archive(ctx, file, libsDir.String(), nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}
	return nil
}

//InstallGitLib  installs a library hosted on a git repository on the specified path.
func (lm *LibrariesManager) InstallGitLib(gitURL string) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("User directory not set")
	}

	libraryName, err := parseGitURL(gitURL)
	if err != nil {
		return err
	}

	installPath := libsDir.Join(libraryName)

	_, err = git.PlainClone(installPath.String(), false, &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
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
		return "", fmt.Errorf("invalid git url")
	}
	return res, nil
}
