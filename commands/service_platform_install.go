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
	"slices"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/libraries"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesindex"
	"github.com/arduino/arduino-cli/internal/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
)

// UpdateIndexStreamResponseToCallbackFunction returns a gRPC stream to be used in PlatformInstall that sends
// all responses to the callback function.
func PlatformInstallStreamResponseToCallbackFunction(ctx context.Context, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) rpc.ArduinoCoreService_PlatformInstallServer {
	return streamResponseToCallback(ctx, func(r *rpc.PlatformInstallResponse) error {
		if r.GetProgress() != nil {
			downloadCB(r.GetProgress())
		}
		if r.GetTaskProgress() != nil {
			taskCB(r.GetTaskProgress())
		}
		return nil
	})
}

// PlatformInstall installs a platform package
func (s *arduinoCoreServerImpl) PlatformInstall(req *rpc.PlatformInstallRequest, stream rpc.ArduinoCoreService_PlatformInstallServer) error {
	ctx := stream.Context()
	syncSend := NewSynchronizedSend(stream.Send)
	taskCB := func(p *rpc.TaskProgress) {
		syncSend.Send(&rpc.PlatformInstallResponse{
			Message: &rpc.PlatformInstallResponse_TaskProgress{
				TaskProgress: p,
			},
		})
	}
	downloadCB := func(p *rpc.DownloadProgress) {
		syncSend.Send(&rpc.PlatformInstallResponse{
			Message: &rpc.PlatformInstallResponse_Progress{
				Progress: p,
			},
		})
	}

	version, err := parseVersion(req.GetVersion())
	if err != nil {
		return &cmderrors.InvalidVersionError{Cause: err}
	}

	install := func() error {
		pme, releasePme, err := instances.GetPackageManagerExplorer(req.GetInstance())
		if err != nil {
			return err
		}
		defer releasePme()

		ref := &packagemanager.PlatformReference{
			Package:              req.GetPlatformPackage(),
			PlatformArchitecture: req.GetArchitecture(),
			PlatformVersion:      version,
		}
		platformRelease, tools, libs, err := pme.FindPlatformReleaseDependencies(ref)
		if err != nil {
			if errors.Is(err, packagemanager.ErrPlatformNotAvailableForOS) {
				return &cmderrors.PlatformNotAvailableForOSError{Platform: ref.String()}
			}
			return &cmderrors.PlatformNotFoundError{Platform: ref.String(), Cause: err}
		}

		// Prerequisite checks before install
		if platformRelease.IsInstalled() {
			taskCB(&rpc.TaskProgress{Name: i18n.Tr("Platform %s already installed", platformRelease), Completed: true})
			return nil
		}

		li, err := instances.GetLibrariesIndex(req.GetInstance())
		if err != nil {
			return err
		}

		lmi, releaseLmi, err := instances.GetLibraryManagerInstaller(req.GetInstance())
		if err != nil {
			return err
		}
		defer releaseLmi()

		if err := s.installLibraries(ctx, li, lmi, libs, pme.DownloadDir, downloadCB, taskCB); err != nil {
			return err
		}

		if req.GetNoOverwrite() {
			if installed := pme.GetInstalledPlatformRelease(platformRelease.Platform); installed != nil {
				return fmt.Errorf("%s: %s",
					i18n.Tr("Platform %s already installed", installed),
					i18n.Tr("could not overwrite"))
			}
		}

		checks := resources.IntegrityCheckFull
		if s.settings.BoardManagerEnableUnsafeInstall() {
			checks = resources.IntegrityCheckNone
		}
		if err := pme.DownloadAndInstallPlatformAndTools(ctx, platformRelease, tools, downloadCB, taskCB, req.GetSkipPostInstall(), req.GetSkipPreUninstall(), checks); err != nil {
			return err
		}

		return nil
	}

	if err := install(); err != nil {
		return err
	}

	if err := s.Init(&rpc.InitRequest{Instance: req.GetInstance()}, InitStreamResponseToCallbackFunction(ctx, nil)); err != nil {
		return err
	}

	return syncSend.Send(&rpc.PlatformInstallResponse{
		Message: &rpc.PlatformInstallResponse_Result_{
			Result: &rpc.PlatformInstallResponse_Result{},
		},
	})
}

// Downloads and installs all libraries in the given list of dependencies, if not already installed.
// If a library is already installed, but with an older version, it will be updated to the required version.
func (s *arduinoCoreServerImpl) installLibraries(
	ctx context.Context, li *librariesindex.Index, lmi *librariesmanager.Installer,
	requiredLibraries cores.LibrariesDependencies,
	downloadsDir *paths.Path,
	downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB,
) error {
	installedLibs := listLibraries(lmi.Explorer, li, false, false)
	for _, libDep := range requiredLibraries {
		matcher := func(lib *installedLib) bool {
			return lib.Library.Name == libDep.Name
		}
		if idx := slices.IndexFunc(installedLibs, matcher); idx != -1 {
			installedVersion := installedLibs[idx].Library.Version
			if installedVersion.Equal(libDep.Version) {
				taskCB(&rpc.TaskProgress{Name: i18n.Tr("Library %s already installed", libDep.Name), Completed: true})
				continue
			}
			if installedVersion.GreaterThanOrEqual(libDep.Version) {
				taskCB(&rpc.TaskProgress{
					Name: i18n.Tr("Skipping installation of library %[1]s, because %[2]s is already installed", libDep, installedVersion), Completed: true})
				continue
			}
		}

		if err := s.downloadAndInstallLibrary(ctx, li, lmi, libDep.Name, libDep.Version.String(), libraries.User, false, false, downloadsDir, taskCB, downloadCB); err != nil {
			return err
		}
	}
	return nil
}
