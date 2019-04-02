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
	"os"

	"go.bug.st/downloader"
	semver "go.bug.st/relaxed-semver"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
)

func PlatformDownload(ctx context.Context, req *rpc.PlatformDownloadReq) (*rpc.PlatformDownloadResp, error) {
	version, err := semver.Parse(req.Version)
	if err != nil {
		formatter.PrintError(err, "version not readable")
		os.Exit(cli.ErrBadCall)
	}
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version}
	pm, _ := cli.InitPackageAndLibraryManagerWithoutBundles()
	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		os.Exit(cli.ErrBadCall)
	}
	downloadPlatform(pm, platform)
	for _, tool := range tools {
		downloadTool(pm, tool)
	}
	return nil, nil
}

func downloadPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease) {
	// Download platform
	resp, err := pm.DownloadPlatformRelease(platformRelease)
	download(resp, err, platformRelease.String())
}

func downloadTool(pm *packagemanager.PackageManager, tool *cores.ToolRelease) {
	// Check if tool has a flavor available for the current OS
	if tool.GetCompatibleFlavour() == nil {
		formatter.PrintErrorMessage("The tool " + tool.String() + " is not available for the current OS")
		os.Exit(cli.ErrGeneric)
	}

	DownloadToolRelease(pm, tool)
}

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) {
	resp, err := pm.DownloadToolRelease(toolRelease)
	download(resp, err, toolRelease.String())
}

func download(d *downloader.Downloader, err error, label string) {
	if err != nil {
		formatter.PrintError(err, "Error downloading "+label)
		os.Exit(cli.ErrNetwork)
	}
	if d == nil {
		formatter.Print(label + " already downloaded")
		return
	}
	formatter.Print("Downloading " + label + "...")
	formatter.DownloadProgressBar(d, label)
	if d.Error() != nil {
		formatter.PrintError(d.Error(), "Error downloading "+label)
		os.Exit(cli.ErrNetwork)
	}
}
