// This file is part of arduino-cli.
//
// Copyright 2020-2022 ARDUINO SA (http://www.arduino.cc/)
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

package packagemanager

import (
	"context"
	"fmt"
	"net/url"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/resources"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/cli/configuration"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

// LoadHardwareForProfile load the hardware platforms for the given profile.
// If installMissing is true then possibly missing tools and platforms will be downloaded and installed.
func (pmb *Builder) LoadHardwareForProfile(ctx context.Context, p *sketch.Profile, installMissing bool, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB, settings *configuration.Settings) []error {
	pmb.profile = p

	// Load required platforms
	var merr []error
	var platformReleases []*cores.PlatformRelease
	indexURLs := map[string]*url.URL{}
	for _, platformRef := range p.Platforms {
		if platformRelease, err := pmb.loadProfilePlatform(ctx, platformRef, installMissing, downloadCB, taskCB, settings); err != nil {
			merr = append(merr, fmt.Errorf("%s: %w", i18n.Tr("loading required platform %s", platformRef), err))
			logrus.WithField("platform", platformRef).WithError(err).Debugf("Error loading platform for profile")
		} else {
			platformReleases = append(platformReleases, platformRelease)
			indexURLs[platformRelease.Name] = platformRef.PlatformIndexURL
			logrus.WithField("platform", platformRef).Debugf("Loaded platform for profile")
		}
	}

	// Load tools dependencies for the platforms
	for _, platformRelease := range platformReleases {
		// TODO: pm.FindPlatformReleaseDependencies(platformRelease)

		for _, toolDep := range platformRelease.ToolDependencies {
			indexURL := indexURLs[toolDep.ToolPackager]
			if err := pmb.loadProfileTool(ctx, toolDep, indexURL, installMissing, downloadCB, taskCB, settings); err != nil {
				merr = append(merr, fmt.Errorf("%s: %w", i18n.Tr("loading required tool %s", toolDep), err))
				logrus.WithField("tool", toolDep).WithField("index_url", indexURL).WithError(err).Debugf("Error loading tool for profile")
			} else {
				logrus.WithField("tool", toolDep).WithField("index_url", indexURL).Debugf("Loaded tool for profile")
			}
		}
	}

	return merr
}

func (pmb *Builder) loadProfilePlatform(ctx context.Context, platformRef *sketch.ProfilePlatformReference, installMissing bool, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB, settings *configuration.Settings) (*cores.PlatformRelease, error) {
	targetPackage := pmb.packages.GetOrCreatePackage(platformRef.Packager)
	platform := targetPackage.GetOrCreatePlatform(platformRef.Architecture)
	release := platform.GetOrCreateRelease(platformRef.Version)

	uid := platformRef.InternalUniqueIdentifier()
	destDir := settings.ProfilesCacheDir().Join(uid)
	if !destDir.IsDir() && installMissing {
		// Try installing the missing platform
		if err := pmb.installMissingProfilePlatform(ctx, platformRef, destDir, downloadCB, taskCB); err != nil {
			return nil, err
		}
	}
	return release, pmb.loadPlatformRelease(release, destDir)
}

func (pmb *Builder) installMissingProfilePlatform(ctx context.Context, platformRef *sketch.ProfilePlatformReference, destDir *paths.Path, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	// Instantiate a temporary package manager only for platform installation
	_ = pmb.tempDir.MkdirAll()
	tmp, err := paths.MkTempDir(pmb.tempDir.String(), "")
	if err != nil {
		return fmt.Errorf("installing missing platform: could not create temp dir %s", err)
	}
	tmpPmb := NewBuilder(tmp, tmp, nil, pmb.DownloadDir, tmp, pmb.userAgent, pmb.downloaderConfig)
	defer tmp.RemoveAll()

	// Download the main index and parse it
	taskCB(&rpc.TaskProgress{Name: i18n.Tr("Downloading platform %s", platformRef)})
	defaultIndexURL, _ := url.Parse(globals.DefaultIndexURL)
	indexesToDownload := []*url.URL{defaultIndexURL}
	if platformRef.PlatformIndexURL != nil {
		indexesToDownload = append(indexesToDownload, platformRef.PlatformIndexURL)
	}
	for _, indexURL := range indexesToDownload {
		indexResource := resources.IndexResource{URL: indexURL}
		if err := indexResource.Download(ctx, tmpPmb.IndexDir, downloadCB, pmb.downloaderConfig); err != nil {
			taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error downloading %s", indexURL)})
			return &cmderrors.FailedDownloadError{Message: i18n.Tr("Error downloading %s", indexURL), Cause: err}
		}
		if err := tmpPmb.LoadPackageIndex(indexURL); err != nil {
			taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error loading index %s", indexURL)})
			return &cmderrors.FailedInstallError{Message: i18n.Tr("Error loading index %s", indexURL), Cause: err}
		}
	}

	// Download the platform
	tmpTargetPackage := tmpPmb.packages.GetOrCreatePackage(platformRef.Packager)
	tmpPlatform := tmpTargetPackage.GetOrCreatePlatform(platformRef.Architecture)
	tmpPlatformRelease := tmpPlatform.GetOrCreateRelease(platformRef.Version)
	tmpPm := tmpPmb.Build()
	tmpPme, tmpRelease := tmpPm.NewExplorer()
	defer tmpRelease()

	if err := tmpPme.DownloadPlatformRelease(ctx, tmpPlatformRelease, downloadCB); err != nil {
		taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error downloading platform %s", tmpPlatformRelease)})
		return &cmderrors.FailedInstallError{Message: i18n.Tr("Error downloading platform %s", tmpPlatformRelease), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	// Perform install
	taskCB(&rpc.TaskProgress{Name: i18n.Tr("Installing platform %s", tmpPlatformRelease)})
	if err := tmpPme.InstallPlatformInDirectory(tmpPlatformRelease, destDir); err != nil {
		taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error installing platform %s", tmpPlatformRelease)})
		return &cmderrors.FailedInstallError{Message: i18n.Tr("Error installing platform %s", tmpPlatformRelease), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	return nil
}

func (pmb *Builder) loadProfileTool(ctx context.Context, toolRef *cores.ToolDependency, indexURL *url.URL, installMissing bool, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB, settings *configuration.Settings) error {
	targetPackage := pmb.packages.GetOrCreatePackage(toolRef.ToolPackager)
	tool := targetPackage.GetOrCreateTool(toolRef.ToolName)

	uid := toolRef.InternalUniqueIdentifier(indexURL)
	destDir := settings.ProfilesCacheDir().Join(uid)

	if !destDir.IsDir() && installMissing {
		// Try installing the missing tool
		toolRelease := tool.GetOrCreateRelease(toolRef.ToolVersion)
		if toolRelease == nil {
			return &cmderrors.InvalidVersionError{Cause: fmt.Errorf(i18n.Tr("version %s not found", toolRef.ToolVersion))}
		}
		if err := pmb.installMissingProfileTool(ctx, toolRelease, destDir, downloadCB, taskCB); err != nil {
			return err
		}
	}

	return pmb.loadToolReleaseFromDirectory(tool, toolRef.ToolVersion, destDir)
}

func (pmb *Builder) installMissingProfileTool(ctx context.Context, toolRelease *cores.ToolRelease, destDir *paths.Path, downloadCB rpc.DownloadProgressCB, taskCB rpc.TaskProgressCB) error {
	// Instantiate a temporary package manager only for platform installation
	tmp, err := paths.MkTempDir(destDir.Parent().String(), "")
	if err != nil {
		return fmt.Errorf("installing missing platform: could not create temp dir %s", err)
	}
	defer tmp.RemoveAll()

	// Download the tool
	toolResource := toolRelease.GetCompatibleFlavour()
	if toolResource == nil {
		return &cmderrors.InvalidVersionError{Cause: fmt.Errorf(i18n.Tr("version %s not available for this operating system", toolRelease))}
	}
	taskCB(&rpc.TaskProgress{Name: i18n.Tr("Downloading tool %s", toolRelease)})
	if err := toolResource.Download(ctx, pmb.DownloadDir, pmb.downloaderConfig, toolRelease.String(), downloadCB, ""); err != nil {
		taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error downloading tool %s", toolRelease)})
		return &cmderrors.FailedInstallError{Message: i18n.Tr("Error installing tool %s", toolRelease), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})

	// Install tool
	taskCB(&rpc.TaskProgress{Name: i18n.Tr("Installing tool %s", toolRelease)})
	if err := toolResource.Install(pmb.DownloadDir, tmp, destDir); err != nil {
		taskCB(&rpc.TaskProgress{Name: i18n.Tr("Error installing tool %s", toolRelease)})
		return &cmderrors.FailedInstallError{Message: i18n.Tr("Error installing tool %s", toolRelease), Cause: err}
	}
	taskCB(&rpc.TaskProgress{Completed: true})
	return nil
}
