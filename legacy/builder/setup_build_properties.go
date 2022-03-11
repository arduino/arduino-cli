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

package builder

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	properties "github.com/arduino/go-properties-orderedmap"
	timeutils "github.com/arduino/go-timeutils"
	"github.com/pkg/errors"
)

type SetupBuildProperties struct{}

func (s *SetupBuildProperties) Run(ctx *types.Context) error {
	packages := ctx.Hardware

	targetPlatform := ctx.TargetPlatform
	actualPlatform := ctx.ActualPlatform
	targetBoard := ctx.TargetBoard

	buildProperties := properties.NewMap()
	buildProperties.Merge(actualPlatform.Properties)
	buildProperties.Merge(targetPlatform.Properties)
	buildProperties.Merge(targetBoard.Properties)

	if ctx.BuildPath != nil {
		buildProperties.SetPath("build.path", ctx.BuildPath)
	}
	if ctx.Sketch != nil {
		buildProperties.Set("build.project_name", ctx.Sketch.MainFile.Base())
	}
	buildProperties.Set("build.arch", strings.ToUpper(targetPlatform.Platform.Architecture))

	// get base folder and use it to populate BUILD_PROPERTIES_RUNTIME_IDE_PATH (arduino and arduino-builder live in the same dir)
	ex, err := os.Executable()
	exPath := ""
	if err == nil {
		exPath = filepath.Dir(ex)
	}

	buildProperties.Set("build.core", ctx.BuildCore)
	buildProperties.SetPath("build.core.path", actualPlatform.InstallDir.Join("cores", buildProperties.Get("build.core")))
	buildProperties.Set("build.system.path", actualPlatform.InstallDir.Join("system").String())
	buildProperties.Set("runtime.platform.path", targetPlatform.InstallDir.String())
	buildProperties.Set("runtime.hardware.path", targetPlatform.InstallDir.Join("..").String())
	buildProperties.Set("runtime.ide.version", ctx.ArduinoAPIVersion)
	buildProperties.Set("runtime.ide.path", exPath)
	buildProperties.Set("build.fqbn", ctx.FQBN.String())
	buildProperties.Set("ide_version", ctx.ArduinoAPIVersion)
	buildProperties.Set("runtime.os", properties.GetOSSuffix())
	buildProperties.Set("build.library_discovery_phase", "0")

	if ctx.OptimizeForDebug {
		if buildProperties.ContainsKey("compiler.optimization_flags.debug") {
			buildProperties.Set("compiler.optimization_flags", buildProperties.Get("compiler.optimization_flags.debug"))
		}
	} else {
		if buildProperties.ContainsKey("compiler.optimization_flags.release") {
			buildProperties.Set("compiler.optimization_flags", buildProperties.Get("compiler.optimization_flags.release"))
		}
	}
	ctx.OptimizationFlags = buildProperties.Get("compiler.optimization_flags")

	variant := buildProperties.Get("build.variant")
	if variant == "" {
		buildProperties.Set("build.variant.path", "")
	} else {
		var variantPlatformRelease *cores.PlatformRelease
		variantParts := strings.Split(variant, ":")
		if len(variantParts) > 1 {
			variantPlatform := packages[variantParts[0]].Platforms[targetPlatform.Platform.Architecture]
			variantPlatformRelease = ctx.PackageManager.GetInstalledPlatformRelease(variantPlatform)
			variant = variantParts[1]
		} else {
			variantPlatformRelease = targetPlatform
		}
		buildProperties.SetPath("build.variant.path", variantPlatformRelease.InstallDir.Join("variants", variant))
	}

	for _, tool := range ctx.AllTools {
		buildProperties.SetPath("runtime.tools."+tool.Tool.Name+".path", tool.InstallDir)
		buildProperties.SetPath("runtime.tools."+tool.Tool.Name+"-"+tool.Version.String()+".path", tool.InstallDir)
	}
	for _, tool := range ctx.RequiredTools {
		buildProperties.SetPath("runtime.tools."+tool.Tool.Name+".path", tool.InstallDir)
		buildProperties.SetPath("runtime.tools."+tool.Tool.Name+"-"+tool.Version.String()+".path", tool.InstallDir)
	}

	if !buildProperties.ContainsKey("software") {
		buildProperties.Set("software", DEFAULT_SOFTWARE)
	}

	if ctx.SketchLocation != nil {
		sourcePath, err := ctx.SketchLocation.Abs()
		if err != nil {
			return err
		}
		sourcePath = sourcePath.Parent()
		buildProperties.SetPath("build.source.path", sourcePath)
	}

	now := time.Now()
	buildProperties.Set("extra.time.utc", strconv.FormatInt(now.Unix(), 10))
	buildProperties.Set("extra.time.local", strconv.FormatInt(timeutils.LocalUnix(now), 10))
	buildProperties.Set("extra.time.zone", strconv.Itoa(timeutils.TimezoneOffsetNoDST(now)))
	buildProperties.Set("extra.time.dst", strconv.Itoa(timeutils.DaylightSavingsOffset(now)))

	buildProperties.Merge(ctx.PackageManager.CustomGlobalProperties)

	// we check if the properties referring to secure boot have been set correctly.
	if buildProperties.ContainsKey("build.keys.type") {
		if buildProperties.Get("build.keys.type") == "public_keys" {
			if !buildProperties.ContainsKey("build.keys.keychain") || !buildProperties.ContainsKey("build.keys.sign_key") || !buildProperties.ContainsKey("build.keys.encrypt_key") {
				return errors.Errorf("%s core does not specify correctly default sign and encryption keys", ctx.BuildCore)
			}
		} else {
			return errors.New("\"build.keys.type\" key only supports \"public_keys\" value for now")
		}
	}

	ctx.BuildProperties = buildProperties

	return nil
}
