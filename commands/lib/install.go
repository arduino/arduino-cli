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

package lib

import (
	"context"
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/libraries"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// LibraryInstall resolves the library dependencies, then downloads and installs the libraries into the install location.
func LibraryInstall(ctx context.Context, req *rpc.LibraryInstallRequest, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req.GetInstance())
	if lm == nil {
		return &arduino.InvalidInstanceError{}
	}

	toInstall := map[string]*rpc.LibraryDependencyStatus{}
	installLocation := libraries.FromRPCLibraryInstallLocation(req.GetInstallLocation())
	if req.NoDeps {
		toInstall[req.Name] = &rpc.LibraryDependencyStatus{
			Name:            req.Name,
			VersionRequired: req.Version,
		}
	} else {
		res, err := LibraryResolveDependencies(ctx, &rpc.LibraryResolveDependenciesRequest{
			Instance: req.Instance,
			Name:     req.Name,
			Version:  req.Version,
		})
		if err != nil {
			return err
		}

		for _, dep := range res.Dependencies {
			if existingDep, has := toInstall[dep.Name]; has {
				if existingDep.VersionRequired != dep.VersionRequired {
					err := errors.New(
						tr("two different versions of the library %[1]s are required: %[2]s and %[3]s",
							dep.Name, dep.VersionRequired, existingDep.VersionRequired))
					return &arduino.LibraryDependenciesResolutionFailedError{Cause: err}
				}
			}
			toInstall[dep.Name] = dep
		}
	}

	// Find the libReleasesToInstall to install
	libReleasesToInstall := map[*librariesindex.Release]*librariesmanager.LibraryInstallPlan{}
	for _, lib := range toInstall {
		libRelease, err := findLibraryIndexRelease(lm, &rpc.LibraryInstallRequest{
			Name:    lib.Name,
			Version: lib.VersionRequired,
		})
		if err != nil {
			return err
		}

		installTask, err := lm.InstallPrerequisiteCheck(libRelease.Library.Name, libRelease.Version, installLocation)
		if err != nil {
			return err
		}
		if installTask.UpToDate {
			taskCB(&rpc.TaskProgress{Message: tr("Already installed %s", libRelease), Completed: true})
			continue
		}

		if req.GetNoOverwrite() {
			if installTask.ReplacedLib != nil {
				return fmt.Errorf(tr("Library %[1]s is already installed, but with a different version: %[2]s", libRelease, installTask.ReplacedLib))
			}
		}
		libReleasesToInstall[libRelease] = installTask
	}

	for libRelease, installTask := range libReleasesToInstall {
		// Checks if libRelease is the requested library and not a dependency
		downloadReason := "depends"
		if libRelease.GetName() == req.Name {
			downloadReason = "install"
			if installTask.ReplacedLib != nil {
				downloadReason = "upgrade"
			}
			if installLocation == libraries.IDEBuiltIn {
				downloadReason += "-builtin"
			}
		}
		if err := downloadLibrary(lm, libRelease, downloadCB, taskCB, downloadReason); err != nil {
			return err
		}
		if err := installLibrary(lm, libRelease, installTask, taskCB); err != nil {
			return err
		}
	}

	if err := commands.Init(&rpc.InitRequest{Instance: req.Instance}, nil); err != nil {
		return err
	}

	return nil
}

func installLibrary(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release, installTask *librariesmanager.LibraryInstallPlan, taskCB rpc.TaskProgressCB) error {
	taskCB(&rpc.TaskProgress{Name: tr("Installing %s", libRelease)})
	logrus.WithField("library", libRelease).Info("Installing library")

	if libReplaced := installTask.ReplacedLib; libReplaced != nil {
		taskCB(&rpc.TaskProgress{Message: tr("Replacing %[1]s with %[2]s", libReplaced, libRelease)})
		if err := lm.Uninstall(libReplaced); err != nil {
			return &arduino.FailedLibraryInstallError{
				Cause: fmt.Errorf("%s: %s", tr("could not remove old library"), err)}
		}
	}
	if err := lm.Install(libRelease, installTask.TargetPath); err != nil {
		return &arduino.FailedLibraryInstallError{Cause: err}
	}

	taskCB(&rpc.TaskProgress{Message: tr("Installed %s", libRelease), Completed: true})
	return nil
}

// ZipLibraryInstall FIXMEDOC
func ZipLibraryInstall(ctx context.Context, req *rpc.ZipLibraryInstallRequest, taskCB rpc.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req.GetInstance())
	if err := lm.InstallZipLib(ctx, paths.New(req.Path), req.Overwrite); err != nil {
		return &arduino.FailedLibraryInstallError{Cause: err}
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	return nil
}

// GitLibraryInstall FIXMEDOC
func GitLibraryInstall(ctx context.Context, req *rpc.GitLibraryInstallRequest, taskCB rpc.TaskProgressCB) error {
	lm := commands.GetLibraryManager(req.GetInstance())
	if err := lm.InstallGitLib(req.Url, req.Overwrite); err != nil {
		return &arduino.FailedLibraryInstallError{Cause: err}
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	return nil
}
