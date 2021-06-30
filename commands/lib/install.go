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
	"fmt"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LibraryInstall FIXMEDOC
func LibraryInstall(ctx context.Context, req *rpc.LibraryInstallRequest,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB) *status.Status {

	lm := commands.GetLibraryManager(req.GetInstance().GetId())

	toInstall := map[string]*rpc.LibraryDependencyStatus{}
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
			return status.Newf(err.Code(), tr("Error resolving dependencies for %[1]s@%[2]s: %[3]s", req.Name, req.Version, err.Message()))
		}

		for _, dep := range res.Dependencies {
			if existingDep, has := toInstall[dep.Name]; has {
				if existingDep.VersionRequired != dep.VersionRequired {
					return status.Newf(codes.FailedPrecondition, tr("Two different versions of the library %[1]s are required: %[2]s and %[3]s"),
						dep.Name, dep.VersionRequired, existingDep.VersionRequired)
				}
			}
			toInstall[dep.Name] = dep
		}
	}

	for _, lib := range toInstall {
		libRelease, err := findLibraryIndexRelease(lm, &rpc.LibraryInstallRequest{
			Name:    lib.Name,
			Version: lib.VersionRequired,
		})
		if err != nil {
			return status.Newf(codes.InvalidArgument, tr("Error looking for library: %s", err))
		}

		if err := downloadLibrary(lm, libRelease, downloadCB, taskCB); err != nil {
			return status.Newf(codes.Unknown, tr("Error downloading library: %s", err))
		}

		if err := installLibrary(lm, libRelease, taskCB); err != nil {
			return status.New(codes.PermissionDenied, err.Error())
		}
	}

	stat := commands.Init(&rpc.InitRequest{Instance: req.Instance}, nil)
	if stat != nil {
		return status.Newf(stat.Code(), tr("Error rescanning libraries: %s", stat.Err()))
	}
	return nil
}

func installLibrary(lm *librariesmanager.LibrariesManager, libRelease *librariesindex.Release, taskCB commands.TaskProgressCB) error {
	taskCB(&rpc.TaskProgress{Name: tr("Installing %s", libRelease)})
	logrus.WithField("library", libRelease).Info("Installing library")
	libPath, libReplaced, err := lm.InstallPrerequisiteCheck(libRelease)
	if err == librariesmanager.ErrAlreadyInstalled {
		taskCB(&rpc.TaskProgress{Message: tr("Already installed %s", libRelease), Completed: true})
		return nil
	}

	if err != nil {
		return fmt.Errorf(tr("checking lib install prerequisites: %s"), err)
	}

	if libReplaced != nil {
		taskCB(&rpc.TaskProgress{Message: tr("Replacing %[1]s with %[2]s", libReplaced, libRelease)})
	}

	if err := lm.Install(libRelease, libPath); err != nil {
		return err
	}

	taskCB(&rpc.TaskProgress{Message: tr("Installed %s", libRelease), Completed: true})
	return nil
}

//ZipLibraryInstall FIXMEDOC
func ZipLibraryInstall(ctx context.Context, req *rpc.ZipLibraryInstallRequest, taskCB commands.TaskProgressCB) *status.Status {
	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	if err := lm.InstallZipLib(ctx, req.Path, req.Overwrite); err != nil {
		return status.New(codes.InvalidArgument, err.Error())
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	return nil
}

//GitLibraryInstall FIXMEDOC
func GitLibraryInstall(ctx context.Context, req *rpc.GitLibraryInstallRequest, taskCB commands.TaskProgressCB) *status.Status {
	lm := commands.GetLibraryManager(req.GetInstance().GetId())
	if err := lm.InstallGitLib(req.Url, req.Overwrite); err != nil {
		return status.New(codes.InvalidArgument, err.Error())
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	return nil
}
