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
	"fmt"
	"time"

	"go.bug.st/downloader"
	semver "go.bug.st/relaxed-semver"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
)

func PlatformDownload(ctx context.Context, req *rpc.PlatformDownloadReq, progressCallback func(*rpc.DownloadProgress)) (*rpc.PlatformDownloadResp, error) {
	version, err := semver.Parse(req.Version)
	if err != nil {
		formatter.PrintError(err, "Error in version parsing")
		return nil, fmt.Errorf("parse from string error: %s", err)
	}
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version}
	pm := commands.GetPackageManager(req)
	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		return nil, fmt.Errorf("find platform dependencies error: %s", err)
	}
	err = downloadPlatform(pm, platform, progressCallback)
	if err != nil {
		return nil, err
	}
	for _, tool := range tools {
		err := downloadTool(pm, tool, progressCallback)
		if err != nil {
			formatter.PrintError(err, "Could not determine platform dependencies")
			return nil, fmt.Errorf("find platform dependencies error: %s", err)
		}
	}
	return &rpc.PlatformDownloadResp{}, nil
}

func downloadPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, progressCallback func(*rpc.DownloadProgress)) error {
	// Download platform
	resp, err := pm.DownloadPlatformRelease(platformRelease)
	if err != nil {
		formatter.PrintError(err, "Error downloading "+platformRelease.String())
		return err
	} else {
		return download(resp, platformRelease.String(), progressCallback)
	}
}

func downloadTool(pm *packagemanager.PackageManager, tool *cores.ToolRelease, progressCallback func(*rpc.DownloadProgress)) error {
	// Check if tool has a flavor available for the current OS
	if tool.GetCompatibleFlavour() == nil {
		formatter.PrintErrorMessage("The tool " + tool.String() + " is not available for the current OS")
		return fmt.Errorf("The tool " + tool.String() + " is not available")
	}

	return DownloadToolRelease(pm, tool, progressCallback)
}

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, progressCallback func(*rpc.DownloadProgress)) error {
	resp, err := pm.DownloadToolRelease(toolRelease)
	if err != nil {
		formatter.PrintError(err, "Error downloading "+toolRelease.String())
		return err
	} else {
		return download(resp, toolRelease.String(), progressCallback)
	}
}

// TODO: Refactor this into output.*?
func download(d *downloader.Downloader, label string, progressCallback func(*rpc.DownloadProgress)) error {
	if d == nil {
		// TODO: Already downloaded
		progressCallback(&rpc.DownloadProgress{
			File:      label,
			Completed: true,
		})
		return nil
	}
	progressCallback(&rpc.DownloadProgress{
		File:      label,
		Url:       d.URL,
		TotalSize: d.Size(),
	})
	d.RunAndPoll(func(downloaded int64) {
		progressCallback(&rpc.DownloadProgress{Downloaded: downloaded})
	}, 250*time.Millisecond)
	if d.Error() != nil {
		formatter.PrintError(d.Error(), "Error downloading "+label)
		return d.Error()
	}
	progressCallback(&rpc.DownloadProgress{Completed: true})
	return nil
}
