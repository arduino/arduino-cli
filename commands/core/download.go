/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/rpc"
	"go.bug.st/downloader"
	semver "go.bug.st/relaxed-semver"
)

func PlatformDownload(ctx context.Context, req *rpc.PlatformDownloadReq, downloadCB commands.DownloadProgressCB) (*rpc.PlatformDownloadResp, error) {
	var version *semver.Version
	if v, err := semver.Parse(req.Version); err == nil {
		version = v
	} else {
		return nil, fmt.Errorf("invalid version: %s", err)
	}

	pm := commands.GetPackageManager(req)
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	platform, tools, err := pm.FindPlatformReleaseDependencies(&packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version,
	})
	if err != nil {
		return nil, fmt.Errorf("find platform dependencies: %s", err)
	}

	err = downloadPlatform(pm, platform, downloadCB)
	if err != nil {
		return nil, err
	}

	for _, tool := range tools {
		err := downloadTool(pm, tool, downloadCB)
		if err != nil {
			return nil, fmt.Errorf("downloading tool %s: %s", tool, err)
		}
	}

	return &rpc.PlatformDownloadResp{}, nil
}

func downloadPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, downloadCB commands.DownloadProgressCB) error {
	// Download platform
	resp, err := pm.DownloadPlatformRelease(platformRelease)
	if err != nil {
		return err
	}
	return download(resp, platformRelease.String(), downloadCB)
}

func downloadTool(pm *packagemanager.PackageManager, tool *cores.ToolRelease, downloadCB commands.DownloadProgressCB) error {
	// Check if tool has a flavor available for the current OS
	if tool.GetCompatibleFlavour() == nil {
		return fmt.Errorf("tool %s not available for the current OS", tool)
	}

	return DownloadToolRelease(pm, tool, downloadCB)
}

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, downloadCB commands.DownloadProgressCB) error {
	resp, err := pm.DownloadToolRelease(toolRelease)
	if err != nil {
		return err
	}
	return download(resp, toolRelease.String(), downloadCB)
}

func download(d *downloader.Downloader, label string, downloadCB commands.DownloadProgressCB) error {
	if d == nil {
		// This signal means that the file is already downloaded
		downloadCB(&rpc.DownloadProgress{
			File:      label,
			Completed: true,
		})
		return nil
	}
	downloadCB(&rpc.DownloadProgress{
		File:      label,
		Url:       d.URL,
		TotalSize: d.Size(),
	})
	d.RunAndPoll(func(downloaded int64) {
		downloadCB(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if d.Error() != nil {
		return d.Error()
	}
	downloadCB(&rpc.DownloadProgress{Completed: true})
	return nil
}
