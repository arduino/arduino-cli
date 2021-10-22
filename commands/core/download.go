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

package core

import (
	"context"
	"errors"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

var tr = i18n.Tr

// PlatformDownload FIXMEDOC
func PlatformDownload(ctx context.Context, req *rpc.PlatformDownloadRequest, downloadCB commands.DownloadProgressCB) (*rpc.PlatformDownloadResponse, error) {
	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	version, err := commands.ParseVersion(req)
	if err != nil {
		return nil, &arduino.InvalidVersionError{Cause: err}
	}

	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version,
	}
	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		return nil, &arduino.PlatformNotFound{Platform: ref.String(), Cause: err}
	}

	if err := downloadPlatform(pm, platform, downloadCB); err != nil {
		return nil, err
	}

	for _, tool := range tools {
		if err := downloadTool(pm, tool, downloadCB); err != nil {
			return nil, err
		}
	}

	return &rpc.PlatformDownloadResponse{}, nil
}

func downloadPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, downloadCB commands.DownloadProgressCB) error {
	// Download platform
	config, err := commands.GetDownloaderConfig()
	if err != nil {
		return &arduino.FailedDownloadError{Message: tr("Error downloading platform %s", platformRelease), Cause: err}
	}
	resp, err := pm.DownloadPlatformRelease(platformRelease, config)
	if err != nil {
		return &arduino.FailedDownloadError{Message: tr("Error downloading platform %s", platformRelease), Cause: err}
	}
	return commands.Download(resp, platformRelease.String(), downloadCB)
}

func downloadTool(pm *packagemanager.PackageManager, tool *cores.ToolRelease, downloadCB commands.DownloadProgressCB) error {
	// Check if tool has a flavor available for the current OS
	if tool.GetCompatibleFlavour() == nil {
		return &arduino.FailedDownloadError{
			Message: tr("Error downloading tool %s", tool),
			Cause:   errors.New(tr("no versions available for the current OS", tool))}
	}

	if err := commands.DownloadToolRelease(pm, tool, downloadCB); err != nil {
		return &arduino.FailedDownloadError{Message: tr("Error downloading tool %s", tool), Cause: err}
	}

	return nil
}
