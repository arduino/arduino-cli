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

package debug

import (
	"context"
	"fmt"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	dbg "github.com/arduino/arduino-cli/rpc/debug"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Debug FIXMEDOC
func Debug(ctx context.Context, req *dbg.DebugReq, in func(data []bytes), out func(data []bytes)) (*exec.Cmd, error) {
	//logrus.Tracef("Debug %s on %s started", req.GetSketchPath(), req.GetFqbn())

	//// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	if req.GetSketchPath() == "" {
		return fmt.Errorf("missing sketchPath")
	}
	sketchPath := paths.New(req.GetSketchPath())
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil {
		return fmt.Errorf("opening sketch: %s", err)
	}

	// FIXME: make a specification on how a port is specified via command line
	port := req.GetPort()
	if port == "" {
		return fmt.Errorf("no upload port provided")
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sketch != nil && sketch.Metadata != nil {
		fqbnIn = sketch.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		return fmt.Errorf("no Fully Qualified Board Name provided")
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return fmt.Errorf("incorrect FQBN: %s", err)
	}

	pm := commands.GetPackageManager(req.GetInstance().GetId())

	// Find target board and board properties
	_, _, board, boardProperties, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return fmt.Errorf("incorrect FQBN: %s", err)
	}

	// Load programmer tool
	uploadToolPattern, have := boardProperties.GetOk("upload.tool")
	if !have || uploadToolPattern == "" {
		return fmt.Errorf("cannot get programmer tool: undefined 'upload.tool' property")
	}

	var referencedPlatformRelease *cores.PlatformRelease
	if split := strings.Split(uploadToolPattern, ":"); len(split) > 2 {
		return fmt.Errorf("invalid 'upload.tool' property: %s", uploadToolPattern)
	} else if len(split) == 2 {
		referencedPackageName := split[0]
		uploadToolPattern = split[1]
		architecture := board.PlatformRelease.Platform.Architecture

		if referencedPackage := pm.Packages[referencedPackageName]; referencedPackage == nil {
			return fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
		} else if referencedPlatform := referencedPackage.Platforms[architecture]; referencedPlatform == nil {
			return fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
		} else {
			referencedPlatformRelease = pm.GetInstalledPlatformRelease(referencedPlatform)
		}
	}

	// Build configuration for upload
	uploadProperties := properties.NewMap()
	if referencedPlatformRelease != nil {
		uploadProperties.Merge(referencedPlatformRelease.Properties)
	}
	uploadProperties.Merge(board.PlatformRelease.Properties)
	uploadProperties.Merge(board.PlatformRelease.RuntimeProperties())
	uploadProperties.Merge(boardProperties)

	uploadToolProperties := uploadProperties.SubTree("tools." + uploadToolPattern)
	uploadProperties.Merge(uploadToolProperties)

	if requiredTools, err := pm.FindToolsRequiredForBoard(board); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for upload")
			uploadProperties.Merge(requiredTool.RuntimeProperties())
		}
	}

	// Set properties for verbose upload
	Verbose := req.GetVerbose()
	if Verbose {
		if v, ok := uploadProperties.GetOk("upload.params.verbose"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
	} else {
		if v, ok := uploadProperties.GetOk("upload.params.quiet"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
	}

	// Set path to compiled binary
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var importPath *paths.Path
	var importFile string
	if req.GetImportFile() == "" {
		importPath = sketch.FullPath
		importFile = sketch.Name + "." + fqbnSuffix
	} else {
		importPath = paths.New(req.GetImportFile()).Parent()
		importFile = paths.New(req.GetImportFile()).Base()
	}

	outputTmpFile, ok := uploadProperties.GetOk("recipe.output.tmp_file")
	outputTmpFile = uploadProperties.ExpandPropsInString(outputTmpFile)
	if !ok {
		return fmt.Errorf("property 'recipe.output.tmp_file' not defined")
	}
	ext := filepath.Ext(outputTmpFile)
	if strings.HasSuffix(importFile, ext) {
		importFile = importFile[:len(importFile)-len(ext)]
	}

	uploadProperties.SetPath("build.path", importPath)
	uploadProperties.Set("build.project_name", importFile)
	uploadFile := importPath.Join(importFile + ext)
	if _, err := uploadFile.Stat(); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("compiled sketch %s not found", uploadFile.String())
		}
		return fmt.Errorf("cannot open sketch: %s", err)
	}

	// Build recipe for upload
	recipe := uploadProperties.Get("upload.pattern")
	cmdLine := uploadProperties.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return fmt.Errorf("invalid recipe '%s': %s", recipe, err)
	}

	cmdArgs := []string{"gdb"}
	// Run Tool
	cmd, err := executils.Command(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("cannot execute upload tool: %s", err)
	}

	return cmd, nil
}
