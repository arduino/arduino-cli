package core

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	semver "go.bug.st/relaxed-semver"
)

func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallReq) (*rpc.PlatformInstallResp, error) {
	var version *semver.Version
	if req.Version != "" {
		if v, err := semver.Parse(req.Version); err == nil {
			version = v
		} else {
			return nil, fmt.Errorf("invalid version: %s", err)
		}
	}
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
		PlatformVersion:      version}
	pm := commands.GetPackageManager(req)
	platform, tools, err := pm.FindPlatformReleaseDependencies(ref)
	if err != nil {
		formatter.PrintError(err, "Could not determine platform dependencies")
		return nil, fmt.Errorf("Could not determine platform dependencies", err)
	}

	err = installPlatform(pm, platform, tools)
	if err != nil {
		formatter.PrintError(err, "Error Installing "+platform.String())
		return nil, err
	}

	return &rpc.PlatformInstallResp{}, nil
}

func installPlatform(pm *packagemanager.PackageManager, platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease) error {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		formatter.Print("Platform " + platformRelease.String() + " already installed")
		return fmt.Errorf("Platform " + platformRelease.String() + " already installed")
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
	print := func(curr *rpc.DownloadProgress) {
		fmt.Printf(">> %v\n", curr)
	}
	for _, tool := range toolsToInstall {
		downloadTool(pm, tool, print)
	}
	downloadPlatform(pm, platformRelease, print)

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
		return fmt.Errorf("Cannot install platform", err)
	}

	// If upgrading remove previous release
	if installed != nil {
		errUn := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if errUn != nil {
			log.WithError(errUn).Error("Error updating platform.")
			formatter.PrintError(errUn, "Error updating platform")

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				formatter.PrintError(err, "Error rolling-back changes.")
				return fmt.Errorf("Error rolling-back changes.", err)
			}
			return fmt.Errorf("Error updating platform", errUn)
		}
	}

	log.Info("Platform installed")
	formatter.Print(platformRelease.String() + " installed")
	return nil
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease) error {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		formatter.Print("Tool " + toolRelease.String() + " already installed")
		return fmt.Errorf("Tool " + toolRelease.String() + " already installed")
	}

	log.Info("Installing tool")
	formatter.Print("Installing " + toolRelease.String() + "...")
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		formatter.PrintError(err, "Cannot install tool: "+toolRelease.String())
		return fmt.Errorf("Cannot install tool: "+toolRelease.String(), err)
	}

	log.Info("Tool installed")
	formatter.Print(toolRelease.String() + " installed")
	return nil
}
