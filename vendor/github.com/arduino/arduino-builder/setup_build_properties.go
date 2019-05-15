/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package builder

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-builder/types"
	"github.com/arduino/arduino-builder/utils"
	"github.com/arduino/arduino-cli/arduino/cores"
	properties "github.com/arduino/go-properties-orderedmap"
	timeutils "github.com/arduino/go-timeutils"
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
		buildProperties.Set("build.project_name", ctx.Sketch.MainFile.Name.Base())
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
	buildProperties.Set("runtime.os", utils.PrettyOSName())

	variant := buildProperties.Get("build.variant")
	if variant == "" {
		buildProperties.Set("build.variant.path", "")
	} else {
		var variantPlatformRelease *cores.PlatformRelease
		variantParts := strings.Split(variant, ":")
		if len(variantParts) > 1 {
			variantPlatform := packages.Packages[variantParts[0]].Platforms[targetPlatform.Platform.Architecture]
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

	ctx.BuildProperties = buildProperties

	return nil
}
