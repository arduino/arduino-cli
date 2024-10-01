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

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/utils"
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
func (lm *LibrariesManager) InstallPrerequisiteCheck(name string, version *semver.Version, installLocation libraries.LibraryLocation) (*LibraryInstallPlan, error) {
	installDir, err := lm.getLibrariesDir(installLocation)
	if err != nil {
		return nil, err
	}

	lm.RescanLibraries()
	libs := lm.FindByReference(&librariesindex.Reference{Name: name}, installLocation)

	if len(libs) > 1 {
		libsDir := paths.NewPathList()
		for _, lib := range libs {
			libsDir.Add(lib.InstallDir)
		}
		return nil, &arduino.MultipleLibraryInstallDetected{
			LibName: name,
			LibsDir: libsDir,
			Message: tr("Automatic library install can't be performed in this case, please manually remove all duplicates and retry."),
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
			return nil, fmt.Errorf(tr("destination dir %s already exists, cannot install"), libPath)
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

// Install installs a library on the specified path.
func (lm *LibrariesManager) Install(indexLibrary *librariesindex.Release, installPath *paths.Path) error {
	return indexLibrary.Resource.Install(lm.DownloadsDir, installPath.Parent(), installPath)
}

// importLibraryFromDirectory installs a library by copying it from the given directory.
func (lm *LibrariesManager) importLibraryFromDirectory(libPath *paths.Path, overwrite bool) error {
	// Check if the library is valid and load metatada
	if err := validateLibrary(libPath); err != nil {
		return err
	}
	library, err := libraries.Load(libPath, libraries.User)
	if err != nil {
		return err
	}

	// Check if the library is already installed and determine install path
	installPlan, err := lm.InstallPrerequisiteCheck(library.Name, library.Version, libraries.User)
	if err != nil {
		return err
	}

	if installPlan.UpToDate {
		if !overwrite {
			return fmt.Errorf(tr("library %s already installed"), installPlan.Name)
		}
	}
	if installPlan.ReplacedLib != nil {
		if !overwrite {
			return fmt.Errorf(tr("Library %[1]s is already installed, but with a different version: %[2]s", installPlan.Name, installPlan.ReplacedLib))
		}
		if err := lm.Uninstall(installPlan.ReplacedLib); err != nil {
			return err
		}
	}
	if installPlan.TargetPath.Exist() {
		return fmt.Errorf("%s: %s", tr("destination directory already exists"), installPlan.TargetPath)
	}
	if err := libPath.CopyDirTo(installPlan.TargetPath); err != nil {
		return fmt.Errorf("%s: %w", tr("copying library to destination directory:"), err)
	}
	return nil
}

// Uninstall removes a Library
func (lm *LibrariesManager) Uninstall(lib *libraries.Library) error {
	if lib == nil || lib.InstallDir == nil {
		return fmt.Errorf(tr("install directory not set"))
	}
	if err := lib.InstallDir.RemoveAll(); err != nil {
		return fmt.Errorf(tr("removing library directory: %s"), err)
	}

	alternatives := lm.Libraries[lib.Name]
	alternatives.Remove(lib)
	lm.Libraries[lib.Name] = alternatives
	return nil
}

// InstallZipLib installs a Zip library on the specified path.
func (lm *LibrariesManager) InstallZipLib(ctx context.Context, archivePath *paths.Path, overwrite bool) error {
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
		return fmt.Errorf(tr("extracting archive: %w"), err)
	}

	libRootFiles, err := tmpDir.ReadDir()
	if err != nil {
		return err
	}
	libRootFiles.FilterOutPrefix("__MACOSX") // Ignores metadata from Mac OS X
	if len(libRootFiles) > 1 {
		return fmt.Errorf(tr("archive is not valid: multiple files found in zip file top level"))
	}
	if len(libRootFiles) == 0 {
		return fmt.Errorf(tr("archive is not valid: no files found in zip file top level"))
	}
	tmpInstallPath := libRootFiles[0]

	// Install extracted library in the destination directory
	if err := lm.importLibraryFromDirectory(tmpInstallPath, overwrite); err != nil {
		return fmt.Errorf(tr("moving extracted archive to destination dir: %s"), err)
	}

	return nil
}

// InstallGitLib installs a library hosted on a git repository on the specified path.
func (lm *LibrariesManager) InstallGitLib(gitURL string, overwrite bool) error {
	gitLibraryName, ref, err := parseGitURL(gitURL)
	if err != nil {
		return err
	}

	// Clone library in a temporary directory
	tmp, err := paths.MkTempDir("", "")
	if err != nil {
		return err
	}
	defer tmp.RemoveAll()
	tmpInstallPath := tmp.Join(gitLibraryName)

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
	if err := lm.importLibraryFromDirectory(tmpInstallPath, overwrite); err != nil {
		return fmt.Errorf(tr("moving extracted archive to destination dir: %s"), err)
	}

	return nil
}

// parseGitURL tries to recover a library name from a git URL.
// Returns an error in case the URL is not a valid git URL.
func parseGitURL(gitURL string) (string, plumbing.Revision, error) {
	var res string
	var rev plumbing.Revision
	if strings.HasPrefix(gitURL, "git@") {
		// We can't parse these as URLs
		i := strings.LastIndex(gitURL, "/")
		res = strings.TrimSuffix(gitURL[i+1:], ".git")
	} else if path := paths.New(gitURL); path != nil && path.Exist() {
		res = path.Base()
	} else if parsed, err := url.Parse(gitURL); parsed.String() != "" && err == nil {
		i := strings.LastIndex(parsed.Path, "/")
		res = strings.TrimSuffix(parsed.Path[i+1:], ".git")
		rev = plumbing.Revision(parsed.Fragment)
	} else {
		return "", "", fmt.Errorf(tr("invalid git url"))
	}
	return res, rev, nil
}

// validateLibrary verifies the dir contains a valid library, meaning it has either
// library.properties file and an header in src/ or an header in its root folder.
// Returns nil if dir contains a valid library, error on all other cases.
func validateLibrary(dir *paths.Path) error {
	if dir.NotExist() {
		return fmt.Errorf(tr("directory doesn't exist: %s", dir))
	}

	searchHeaderFile := func(d *paths.Path) (bool, error) {
		if d.NotExist() {
			// A directory that doesn't exist can't obviously contain any header file
			return false, nil
		}
		dirContent, err := d.ReadDir()
		if err != nil {
			return false, fmt.Errorf(tr("reading directory %s content: %w", dir, err))
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

	return fmt.Errorf(tr("library not valid"))
}
