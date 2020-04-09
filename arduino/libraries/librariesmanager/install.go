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
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
	paths "github.com/arduino/go-paths-helper"
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
func (lm *LibrariesManager) InstallZipLib(libPath string) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("User directory not set")
	}
	err := Unzip(libPath, libsDir.String())
	if err != nil {
		return err
	}
	return nil
}

//Unzip takes the ZipLibPath and Extracts it to LibsDir
func Unzip(src string, dest string) error {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

//InstallGitLib  installs a library hosted on a git repository on the specified path.
func (lm *LibrariesManager) InstallGitLib(Name string, gitURL string) error {
	libsDir := lm.getUserLibrariesDir()
	if libsDir == nil {
		return fmt.Errorf("User directory not set")
	}
	err := Fetch(Name, gitURL, libsDir.String())
	if err != nil {
		return err
	}
	return nil
}

//Fetch Clones the repository to LibsDir
func Fetch(name string, url string, dest string) error {
	directory := dest + "/" + name
	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return err
	}
	return nil
}
