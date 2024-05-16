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

package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// LibraryInstallStreamResponseToCallbackFunction returns a gRPC stream to be used in LibraryInstall that sends
// all responses to the callback function.
func LibraryInstallStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_LibraryInstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.LibraryInstallResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// LibraryInstall resolves the library dependencies, then downloads and installs the libraries into the install location.
func (s *arduinoCoreServerImpl) LibraryInstall(req *rpc.LibraryInstallRequest, stream rpc.ArduinoCoreService_LibraryInstallServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.LibraryInstallResponse{
			Message: &rpc.LibraryInstallResponse_Progress{Progress: p},
		})
	}
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.LibraryInstallResponse{
			Message: &rpc.LibraryInstallResponse_TaskProgress{TaskProgress: p},
		})
	}

	// Obtain the library index from the manager
	li, err := instances.GetLibrariesIndex(req.GetInstance())
	if err != nil {
		return err
	}

	toInstall := map[string]*rpc.LibraryDependencyStatus{}
	if req.GetNoDeps() {
		toInstall[req.GetName()] = &rpc.LibraryDependencyStatus{
			Name:            req.GetName(),
			VersionRequired: req.GetVersion(),
		}
	} else {
		// Obtain the library explorer from the instance
		lme, releaseLme, err := instances.GetLibraryManagerExplorer(req.GetInstance())
		if err != nil {
			return err
		}

		res, err := libraryResolveDependencies(lme, li, req.GetName(), req.GetVersion(), req.GetNoOverwrite())
		releaseLme()
		if err != nil {
			return err
		}

		for _, dep := range res.GetDependencies() {
			if existingDep, has := toInstall[dep.GetName()]; has {
				if existingDep.GetVersionRequired() != dep.GetVersionRequired() {
					err := errors.New(
						tr("two different versions of the library %[1]s are required: %[2]s and %[3]s",
							dep.GetName(), dep.GetVersionRequired(), existingDep.GetVersionRequired()))
					return &cmderrors.LibraryDependenciesResolutionFailedError{Cause: err}
				}
			}
			toInstall[dep.GetName()] = dep
		}
	}

	// Obtain the download directory
	var downloadsDir *paths.Path
	if pme, releasePme, err := instances.GetPackageManagerExplorer(req.GetInstance()); err != nil {
		return err
	} else {
		downloadsDir = pme.DownloadDir
		releasePme()
	}

	// Obtain the library installer from the manager
	lmi, releaseLmi, err := instances.GetLibraryManagerInstaller(req.GetInstance())
	if err != nil {
		return err
	}
	defer releaseLmi()

	// Find the libReleasesToInstall to install
	libReleasesToInstall := map[*librariesindex.Release]*librariesmanager.LibraryInstallPlan{}
	installLocation := libraries.FromRPCLibraryInstallLocation(req.GetInstallLocation())
	for _, lib := range toInstall {
		version, err := parseVersion(lib.GetVersionRequired())
		if err != nil {
			return err
		}
		libRelease, err := li.FindRelease(lib.GetName(), version)
		if err != nil {
			return err
		}

		installTask, err := lmi.InstallPrerequisiteCheck(libRelease.Library.Name, libRelease.Version, installLocation)
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
		if libRelease.GetName() == req.GetName() {
			downloadReason = "install"
			if installTask.ReplacedLib != nil {
				downloadReason = "upgrade"
			}
			if installLocation == libraries.IDEBuiltIn {
				downloadReason += "-builtin"
			}
		}
		if err := downloadLibrary(ctx, downloadsDir, libRelease, downloadCB, taskCB, downloadReason, s.settings); err != nil {
			return err
		}
		if err := installLibrary(lmi, downloadsDir, libRelease, installTask, taskCB); err != nil {
			return err
		}
	}

	err = s.Init(
		&rpc.InitRequest{Instance: req.GetInstance()},
		InitStreamResponseToCallbackFunction(ctx, nil))
	if err != nil {
		return err
	}

	syncSend.Send(&rpc.LibraryInstallResponse{
		Message: &rpc.LibraryInstallResponse_Result_{
			Result: &rpc.LibraryInstallResponse_Result{},
		},
	})
	return nil
}

func installLibrary(lmi *librariesmanager.Installer, downloadsDir *paths.Path, libRelease *librariesindex.Release, installTask *librariesmanager.LibraryInstallPlan, taskCB rpc.TaskProgressCB) error {
	taskCB(&rpc.TaskProgress{Name: tr("Installing %s", libRelease)})
	logrus.WithField("library", libRelease).Info("Installing library")

	if libReplaced := installTask.ReplacedLib; libReplaced != nil {
		taskCB(&rpc.TaskProgress{Message: tr("Replacing %[1]s with %[2]s", libReplaced, libRelease)})
		if err := lmi.Uninstall(libReplaced); err != nil {
			return &cmderrors.FailedLibraryInstallError{
				Cause: fmt.Errorf("%s: %s", tr("could not remove old library"), err)}
		}
	}

	installPath := installTask.TargetPath
	tmpDirPath := installPath.Parent()
	if err := libRelease.Resource.Install(downloadsDir, tmpDirPath, installPath); err != nil {
		return &cmderrors.FailedLibraryInstallError{Cause: err}
	}

	taskCB(&rpc.TaskProgress{Message: tr("Installed %s", libRelease), Completed: true})
	return nil
}

// ZipLibraryInstallStreamResponseToCallbackFunction returns a gRPC stream to be used in ZipLibraryInstall that sends
// all responses to the callback function.
func ZipLibraryInstallStreamResponseToCallbackFunction(ctx context.Context, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_ZipLibraryInstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.ZipLibraryInstallResponse) error {
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// ZipLibraryInstall FIXMEDOC
func (s *arduinoCoreServerImpl) ZipLibraryInstall(req *rpc.ZipLibraryInstallRequest, stream rpc.ArduinoCoreService_ZipLibraryInstallServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.ZipLibraryInstallResponse{
			Message: &rpc.ZipLibraryInstallResponse_TaskProgress{TaskProgress: p},
		})
	}

	lm, err := instances.GetLibraryManager(req.GetInstance())
	if err != nil {
		return err
	}
	lmi, release := lm.NewInstaller()
	defer release()
	if err := lmi.InstallZipLib(ctx, paths.New(req.GetPath()), req.GetOverwrite()); err != nil {
		return &cmderrors.FailedLibraryInstallError{Cause: err}
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	syncSend.Send(&rpc.ZipLibraryInstallResponse{
		Message: &rpc.ZipLibraryInstallResponse_Result_{
			Result: &rpc.ZipLibraryInstallResponse_Result{},
		},
	})
	return nil
}

// GitLibraryInstallStreamResponseToCallbackFunction returns a gRPC stream to be used in GitLibraryInstall that sends
// all responses to the callback function.
func GitLibraryInstallStreamResponseToCallbackFunction(ctx context.Context, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_GitLibraryInstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.GitLibraryInstallResponse) error {
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// GitLibraryInstall FIXMEDOC
func (s *arduinoCoreServerImpl) GitLibraryInstall(req *rpc.GitLibraryInstallRequest, stream rpc.ArduinoCoreService_GitLibraryInstallServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.GitLibraryInstallResponse{
			Message: &rpc.GitLibraryInstallResponse_TaskProgress{TaskProgress: p},
		})
	}
	lm, err := instances.GetLibraryManager(req.GetInstance())
	if err != nil {
		return err
	}
	lmi, release := lm.NewInstaller()
	defer release()

	// TODO: pass context
	// ctx := stream.Context()
	if err := lmi.InstallGitLib(req.GetUrl(), req.GetOverwrite()); err != nil {
		return &cmderrors.FailedLibraryInstallError{Cause: err}
	}
	taskCB(&rpc.TaskProgress{Message: tr("Library installed"), Completed: true})
	syncSend.Send(&rpc.GitLibraryInstallResponse{
		Message: &rpc.GitLibraryInstallResponse_Result_{
			Result: &rpc.GitLibraryInstallResponse_Result{},
		},
	})
	return nil
}
