package core

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	semver "go.bug.st/relaxed-semver"
)

//"packager:arch@version
func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallReq) (*rpc.PlatformInstallResp, error) {
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

	installPlatform(pm, platform, tools)

	return nil, nil
}

func installPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease) {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		formatter.Print("Platform " + platformRelease.String() + " already installed")
		return
	}
	toolsToInstall := []*cores.ToolRelease{}
	for _, tool := range requiredTools {
		if tool.IsInstalled() {
			log.WithField("tool", tool).Warn("Tool already installed")
			formatter.Print("Tool " + tool.String() + " already installed")
		} else {
			toolsToInstall = append(toolsToInstall, tool)
		}
	}

	// Package download
	for _, tool := range toolsToInstall {
		downloadTool(pm, tool)
	}
	downloadPlatform(pm, platformRelease)

	for _, tool := range toolsToInstall {
		InstallToolRelease(pm, tool)
	}

	// Are we installing or upgrading?
	platform := platformRelease.Platform
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		log.Info("Installing platform")
		formatter.Print("Installing " + platformRelease.String() + "...")
	} else {
		log.Info("Updating platform " + installed.String())
		formatter.Print("Updating " + installed.String() + " with " + platformRelease.String() + "...")
	}

	// Install
	err := pm.InstallPlatform(platformRelease)
	if err != nil {
		log.WithError(err).Error("Cannot install platform")
		formatter.PrintError(err, "Cannot install platform")
		os.Exit(cli.ErrGeneric)
	}

	// If upgrading remove previous release
	if installed != nil {
		err := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if err != nil {
			log.WithError(err).Error("Error updating platform.")
			formatter.PrintError(err, "Error updating platform")

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				formatter.PrintError(err, "Error rolling-back changes.")
			}
			os.Exit(cli.ErrGeneric)
		}
	}

	log.Info("Platform installed")
	formatter.Print(platformRelease.String() + " installed")
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		formatter.Print("Tool " + toolRelease.String() + " already installed")
		return
	}

	log.Info("Installing tool")
	formatter.Print("Installing " + toolRelease.String() + "...")
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		formatter.PrintError(err, "Cannot install tool: "+toolRelease.String())
		os.Exit(cli.ErrGeneric)
	}

	log.Info("Tool installed")
	formatter.Print(toolRelease.String() + " installed")
}
