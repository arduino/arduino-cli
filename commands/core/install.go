package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/rpc"
	semver "go.bug.st/relaxed-semver"
)

func PlatformInstall(ctx context.Context, req *rpc.PlatformInstallReq,
	progress commands.ProgressCB, taskCB commands.TaskProgressCB) (*rpc.PlatformInstallResp, error) {
	var version *semver.Version
	if req.Version != "" {
		if v, err := semver.Parse(req.Version); err == nil {
			version = v
		} else {
			return nil, fmt.Errorf("invalid version: %s", err)
		}
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
		return nil, fmt.Errorf("finding platform dependencies: %s", err)
	}

	err = installPlatform(pm, platform, tools, progress, taskCB)
	if err != nil {
		return nil, err
	}

	_, err = commands.Rescan(ctx, &rpc.RescanReq{Instance: req.Instance})
	if err != nil {
		return nil, err
	}

	return &rpc.PlatformInstallResp{}, nil
}

func installPlatform(pm *packagemanager.PackageManager,
	platformRelease *cores.PlatformRelease, requiredTools []*cores.ToolRelease,
	progress commands.ProgressCB, taskCB commands.TaskProgressCB) error {
	log := pm.Log.WithField("platform", platformRelease)

	// Prerequisite checks before install
	if platformRelease.IsInstalled() {
		log.Warn("Platform already installed")
		taskCB(&rpc.TaskProgress{Name: "Platform " + platformRelease.String() + " already installed", Completed: true})
		return nil
	}
	toolsToInstall := []*cores.ToolRelease{}
	for _, tool := range requiredTools {
		if tool.IsInstalled() {
			log.WithField("tool", tool).Warn("Tool already installed")
			taskCB(&rpc.TaskProgress{Name: "Tool " + tool.String() + " already installed", Completed: true})
		} else {
			toolsToInstall = append(toolsToInstall, tool)
		}
	}

	// Package download
	taskCB(&rpc.TaskProgress{Name: "Downloading packages"})
	for _, tool := range toolsToInstall {
		downloadTool(pm, tool, progress)
	}
	downloadPlatform(pm, platformRelease, progress)
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tools first
	for _, tool := range toolsToInstall {
		err := InstallToolRelease(pm, tool, taskCB)
		if err != nil {
			// TODO: handle error
		}
	}

	// Are we installing or upgrading?
	platform := platformRelease.Platform
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		log.Info("Installing platform")
		taskCB(&rpc.TaskProgress{Name: "Installing " + platformRelease.String()})
	} else {
		log.Info("Updating platform " + installed.String())
		taskCB(&rpc.TaskProgress{Name: "Updating " + installed.String() + " with " + platformRelease.String()})
	}

	// Install
	err := pm.InstallPlatform(platformRelease)
	if err != nil {
		log.WithError(err).Error("Cannot install platform")
		return err
	}

	// If upgrading remove previous release
	if installed != nil {
		errUn := pm.UninstallPlatform(installed)

		// In case of error try to rollback
		if errUn != nil {
			log.WithError(errUn).Error("Error updating platform.")
			taskCB(&rpc.TaskProgress{Message: "Error updating platform: " + err.Error()})

			// Rollback
			if err := pm.UninstallPlatform(platformRelease); err != nil {
				log.WithError(err).Error("Error rolling-back changes.")
				taskCB(&rpc.TaskProgress{Message: "Error rolling-back changes: " + err.Error()})
				//return fmt.Errorf("rolling-back changes: %s", err)
			}

			return fmt.Errorf("updating platform: %s", errUn)
		}
	}

	log.Info("Platform installed")
	taskCB(&rpc.TaskProgress{Message: platformRelease.String() + " installed", Completed: true})
	return nil
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, taskCB commands.TaskProgressCB) error {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		taskCB(&rpc.TaskProgress{Name: "Tool " + toolRelease.String() + " already installed", Completed: true})
		return nil
	}

	log.Info("Installing tool")
	taskCB(&rpc.TaskProgress{Name: "Installing " + toolRelease.String()})
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		formatter.PrintError(err, "Cannot install tool: "+toolRelease.String())
		return fmt.Errorf("Cannot install tool: "+toolRelease.String(), err)
	}
	log.Info("Tool installed")
	taskCB(&rpc.TaskProgress{Message: toolRelease.String() + " installed", Completed: true})

	return nil
}
