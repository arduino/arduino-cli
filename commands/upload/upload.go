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

package upload

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

// Upload FIXMEDOC
func Upload(ctx context.Context, req *rpc.UploadReq, outStream io.Writer, errStream io.Writer) (*rpc.UploadResp, error) {
	logrus.Tracef("Upload %s on %s started", req.GetSketchPath(), req.GetFqbn())

	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	sketchPath := paths.New(req.GetSketchPath())
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil && req.GetImportDir() == "" && req.GetImportFile() == "" {
		return nil, fmt.Errorf("opening sketch: %s", err)
	}

	pm := commands.GetPackageManager(req.GetInstance().GetId())

	err = runProgramAction(
		pm,
		sketch,
		req.GetImportFile(),
		req.GetImportDir(),
		req.GetFqbn(),
		req.GetPort(),
		req.GetProgrammer(),
		req.GetVerbose(),
		req.GetVerify(),
		false, // burnBootloader
		outStream,
		errStream,
	)
	if err != nil {
		return nil, err
	}
	return &rpc.UploadResp{}, nil
}

// UsingProgrammer FIXMEDOC
func UsingProgrammer(ctx context.Context, req *rpc.UploadUsingProgrammerReq, outStream io.Writer, errStream io.Writer) (*rpc.UploadUsingProgrammerResp, error) {
	logrus.Tracef("Upload using programmer %s on %s started", req.GetSketchPath(), req.GetFqbn())

	if req.GetProgrammer() == "" {
		return nil, errors.New("programmer not specified")
	}
	_, err := Upload(ctx, &rpc.UploadReq{
		Instance:   req.GetInstance(),
		SketchPath: req.GetSketchPath(),
		ImportFile: req.GetImportFile(),
		ImportDir:  req.GetImportDir(),
		Fqbn:       req.GetFqbn(),
		Port:       req.GetPort(),
		Programmer: req.GetProgrammer(),
		Verbose:    req.GetVerbose(),
		Verify:     req.GetVerify(),
	}, outStream, errStream)
	return &rpc.UploadUsingProgrammerResp{}, err
}

func runProgramAction(pm *packagemanager.PackageManager,
	sketch *sketches.Sketch,
	importFile, importDir, fqbnIn, port string,
	programmerID string,
	verbose, verify, burnBootloader bool,
	outStream, errStream io.Writer) error {

	if burnBootloader && programmerID == "" {
		return fmt.Errorf("no programmer specified for burning bootloader")
	}

	// FIXME: make a specification on how a port is specified via command line
	if port == "" && sketch != nil && sketch.Metadata != nil {
		deviceURI, err := url.Parse(sketch.Metadata.CPU.Port)
		if err != nil {
			return fmt.Errorf("invalid Device URL format: %s", err)
		}
		if deviceURI.Scheme == "serial" {
			port = deviceURI.Host + deviceURI.Path
		}
	}
	logrus.WithField("port", port).Tracef("Upload port")

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
	logrus.WithField("fqbn", fqbn).Tracef("Detected FQBN")

	// Find target board and board properties
	_, boardPlatform, board, boardProperties, buildPlatform, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return fmt.Errorf("incorrect FQBN: %s", err)
	}
	logrus.
		WithField("boardPlatform", boardPlatform).
		WithField("board", board).
		WithField("buildPlatform", buildPlatform).
		Tracef("Upload data")

	// Extract programmer properties (when specified)
	var programmer *cores.Programmer
	if programmerID != "" {
		programmer = boardPlatform.Programmers[programmerID]
		if programmer == nil {
			// Try to find the programmer in the referenced build platform
			programmer = buildPlatform.Programmers[programmerID]
		}
		if programmer == nil {
			return fmt.Errorf("programmer '%s' not available", programmerID)
		}
	}

	// Determine upload tool
	var uploadToolID string
	{
		toolProperty := "upload.tool"
		if burnBootloader {
			toolProperty = "bootloader.tool"
		} else if programmer != nil {
			toolProperty = "program.tool"
		}

		// create a temporary configuration only for the selection of upload tool
		props := properties.NewMap()
		props.Merge(boardPlatform.Properties)
		props.Merge(boardPlatform.RuntimeProperties())
		props.Merge(boardProperties)
		if programmer != nil {
			props.Merge(programmer.Properties)
		}
		if t, ok := props.GetOk(toolProperty); ok {
			uploadToolID = t
		} else {
			return fmt.Errorf("cannot get programmer tool: undefined '%s' property", toolProperty)
		}
	}

	var uploadToolPlatform *cores.PlatformRelease
	if programmer != nil {
		uploadToolPlatform = programmer.PlatformRelease
	} else {
		uploadToolPlatform = boardPlatform
	}
	logrus.
		WithField("uploadToolID", uploadToolID).
		WithField("uploadToolPlatform", uploadToolPlatform).
		Trace("Upload tool")

	if split := strings.Split(uploadToolID, ":"); len(split) > 2 {
		return fmt.Errorf("invalid 'upload.tool' property: %s", uploadToolID)
	} else if len(split) == 2 {
		uploadToolID = split[1]
		uploadToolPlatform = pm.GetInstalledPlatformRelease(
			pm.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: boardPlatform.Platform.Architecture,
			}),
		)
	}

	// Build configuration for upload
	uploadProperties := properties.NewMap()
	if uploadToolPlatform != nil {
		uploadProperties.Merge(uploadToolPlatform.Properties)
	}
	uploadProperties.Merge(boardPlatform.Properties)
	uploadProperties.Merge(boardPlatform.RuntimeProperties())
	uploadProperties.Merge(boardProperties)
	uploadProperties.Merge(uploadProperties.SubTree("tools." + uploadToolID))
	if programmer != nil {
		uploadProperties.Merge(programmer.Properties)
	}

	for _, tool := range pm.GetAllInstalledToolsReleases() {
		uploadProperties.Merge(tool.RuntimeProperties())
	}
	if requiredTools, err := pm.FindToolsRequiredForBoard(board); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for upload")
			if requiredTool.IsInstalled() {
				uploadProperties.Merge(requiredTool.RuntimeProperties())
			} else {
				errStream.Write([]byte(fmt.Sprintf("Warning: tool '%s' is not installed. It might not be available for your OS.", requiredTool)))
			}
		}
	}

	if !uploadProperties.ContainsKey("upload.protocol") && programmer == nil {
		return fmt.Errorf("a programmer is required to upload for this board")
	}

	// Set properties for verbose upload
	if verbose {
		if v, ok := uploadProperties.GetOk("upload.params.verbose"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("program.params.verbose"); ok {
			uploadProperties.Set("program.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("erase.params.verbose"); ok {
			uploadProperties.Set("erase.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("bootloader.params.verbose"); ok {
			uploadProperties.Set("bootloader.verbose", v)
		}
	} else {
		if v, ok := uploadProperties.GetOk("upload.params.quiet"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("program.params.quiet"); ok {
			uploadProperties.Set("program.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("erase.params.quiet"); ok {
			uploadProperties.Set("erase.verbose", v)
		}
		if v, ok := uploadProperties.GetOk("bootloader.params.quiet"); ok {
			uploadProperties.Set("bootloader.verbose", v)
		}
	}

	// Set properties for verify
	if verify {
		uploadProperties.Set("upload.verify", uploadProperties.Get("upload.params.verify"))
		uploadProperties.Set("program.verify", uploadProperties.Get("program.params.verify"))
		uploadProperties.Set("erase.verify", uploadProperties.Get("erase.params.verify"))
		uploadProperties.Set("bootloader.verify", uploadProperties.Get("bootloader.params.verify"))
	} else {
		uploadProperties.Set("upload.verify", uploadProperties.Get("upload.params.noverify"))
		uploadProperties.Set("program.verify", uploadProperties.Get("program.params.noverify"))
		uploadProperties.Set("erase.verify", uploadProperties.Get("erase.params.noverify"))
		uploadProperties.Set("bootloader.verify", uploadProperties.Get("bootloader.params.noverify"))
	}

	if !burnBootloader {
		importPath, sketchName, err := determineBuildPathAndSketchName(importFile, importDir, sketch, fqbn)
		if err != nil {
			return errors.Errorf("retrieving build artifacts: %s", err)
		}
		if !importPath.Exist() {
			return fmt.Errorf("compiled sketch not found in %s", importPath)
		}
		if !importPath.IsDir() {
			return fmt.Errorf("expected compiled sketch in directory %s, but is a file instead", importPath)
		}
		uploadProperties.SetPath("build.path", importPath)
		uploadProperties.Set("build.project_name", sketchName)
	}

	// If not using programmer perform some action required
	// to set the board in bootloader mode
	actualPort := port
	if programmer == nil && !burnBootloader {
		// Perform reset via 1200bps touch if requested
		if uploadProperties.GetBoolean("upload.use_1200bps_touch") {
			if port == "" {
				return fmt.Errorf("no upload port provided")
			}

			ports, err := serial.GetPortsList()
			if err != nil {
				return fmt.Errorf("cannot get serial port list: %s", err)
			}
			for _, p := range ports {
				if p == port {
					if verbose {
						outStream.Write([]byte(fmt.Sprintf("Performing 1200-bps touch reset on serial port %s", p)))
						outStream.Write([]byte(fmt.Sprintln()))
					}
					logrus.Infof("Touching port %s at 1200bps", port)
					if err := serialutils.TouchSerialPortAt1200bps(p); err != nil {
						outStream.Write([]byte(fmt.Sprintf("Cannot perform port reset: %s", err)))
						outStream.Write([]byte(fmt.Sprintln()))
					}
					break
				}
			}
		}

		// Wait for upload port if requested
		if uploadProperties.GetBoolean("upload.wait_for_upload_port") {
			if verbose {
				outStream.Write([]byte(fmt.Sprintln("Waiting for upload port...")))
			}

			actualPort, err = serialutils.WaitForNewSerialPortOrDefaultTo(actualPort)
			if err != nil {
				return errors.WithMessage(err, "detecting serial port")
			}
		}
	}

	if port != "" {
		// Set serial port property
		uploadProperties.Set("serial.port", actualPort)
		if strings.HasPrefix(actualPort, "/dev/") {
			uploadProperties.Set("serial.port.file", actualPort[5:])
		} else {
			uploadProperties.Set("serial.port.file", actualPort)
		}
	}

	// Run recipes for upload
	if burnBootloader {
		if err := runTool("erase.pattern", uploadProperties, outStream, errStream, verbose); err != nil {
			return fmt.Errorf("chip erase error: %s", err)
		}
		if err := runTool("bootloader.pattern", uploadProperties, outStream, errStream, verbose); err != nil {
			return fmt.Errorf("burn bootloader error: %s", err)
		}
	} else if programmer != nil {
		if err := runTool("program.pattern", uploadProperties, outStream, errStream, verbose); err != nil {
			return fmt.Errorf("programming error: %s", err)
		}
	} else {
		if err := runTool("upload.pattern", uploadProperties, outStream, errStream, verbose); err != nil {
			return fmt.Errorf("uploading error: %s", err)
		}
	}

	logrus.Tracef("Upload successful")
	return nil
}

func runTool(recipeID string, props *properties.Map, outStream, errStream io.Writer, verbose bool) error {
	recipe, ok := props.GetOk(recipeID)
	if !ok {
		return fmt.Errorf("recipe not found '%s'", recipeID)
	}
	if strings.TrimSpace(recipe) == "" {
		return nil // Nothing to run
	}
	if props.IsPropertyMissingInExpandPropsInString("serial.port", recipe) {
		return fmt.Errorf("no upload port provided")
	}
	cmdLine := props.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return fmt.Errorf("invalid recipe '%s': %s", recipe, err)
	}

	// Run Tool
	if verbose {
		outStream.Write([]byte(fmt.Sprintln(cmdLine)))
	}
	cmd, err := executils.NewProcess(cmdArgs...)
	if err != nil {
		return fmt.Errorf("cannot execute upload tool: %s", err)
	}

	cmd.RedirectStdoutTo(outStream)
	cmd.RedirectStderrTo(errStream)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cannot execute upload tool: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("uploading error: %s", err)
	}

	return nil
}

func determineBuildPathAndSketchName(importFile, importDir string, sketch *sketches.Sketch, fqbn *cores.FQBN) (*paths.Path, string, error) {
	// In general, compiling a sketch will produce a set of files that are
	// named as the sketch but have different extensions, for example Sketch.ino
	// may produce: Sketch.ino.bin; Sketch.ino.hex; Sketch.ino.zip; etc...
	// These files are created together in the build directory and anyone of
	// them may be required for upload.

	// The upload recipes are already written using the 'build.project_name' property
	// concatenated with an explicit extension. To perform the upload we should now
	// determine the project name (e.g. 'sketch.ino) and set the 'build.project_name'
	// property accordingly, together with the 'build.path' property to point to the
	// directory containing the build artifacts.

	// Case 1: importFile flag has been specified
	if importFile != "" {
		if importDir != "" {
			return nil, "", fmt.Errorf("importFile and importDir cannot be used together")
		}

		// We have a path like "path/to/my/build/SketchName.ino.bin". We are going to
		// ignore the extension and set:
		// - "build.path" as "path/to/my/build"
		// - "build.project_name" as "SketchName.ino"

		importFilePath := paths.New(importFile)
		if !importFilePath.Exist() {
			return nil, "", fmt.Errorf("binary file not found in %s", importFilePath)
		}
		return importFilePath.Parent(), strings.TrimSuffix(importFilePath.Base(), importFilePath.Ext()), nil
	}

	if importDir != "" {
		// Case 2: importDir flag has been specified

		// In this case we have a build path but ignore the sketch name, we'll
		// try to determine the sketch name by applying some euristics to the build folder.
		// - "build.path" as importDir
		// - "build.project_name" after trying to autodetect it from the build folder.
		buildPath := paths.New(importDir)
		sketchName, err := detectSketchNameFromBuildPath(buildPath)
		if err != nil {
			return nil, "", errors.Errorf("autodetect build artifact: %s", err)
		}
		return buildPath, sketchName, nil
	}

	// Case 3: nothing given...
	if sketch == nil {
		return nil, "", fmt.Errorf("no sketch or build directory/file specified")
	}

	// Case 4: only sketch specified. In this case we use the generated build path
	// and the given sketch name.
	return bldr.GenBuildPath(sketch.FullPath), sketch.Name + sketch.MainFileExtension, nil
}

func detectSketchNameFromBuildPath(buildPath *paths.Path) (string, error) {
	files, err := buildPath.ReadDir()
	if err != nil {
		return "", err
	}

	if absBuildPath, err := buildPath.Abs(); err == nil {
		for ext := range globals.MainFileValidExtensions {
			candidateName := absBuildPath.Base() + ext
			f := files.Clone()
			f.FilterPrefix(candidateName + ".")
			if f.Len() > 0 {
				return candidateName, nil
			}
		}
	}

	candidateName := ""
	var candidateFile *paths.Path
	for _, file := range files {
		// Build artifacts are usually names as "Blink.ino.hex" or "Blink.ino.bin".
		// Extract the "Blink.ino" part
		name := strings.TrimSuffix(file.Base(), file.Ext())

		// Sometimes we may have particular files like:
		// Blink.ino.with_bootloader.bin
		if _, ok := globals.MainFileValidExtensions[filepath.Ext(name)]; !ok {
			// just ignore those files
			continue
		}

		if candidateName == "" {
			candidateName = name
			candidateFile = file
		}

		if candidateName != name {
			return "", errors.Errorf("multiple build artifacts found: '%s' and '%s'", candidateFile, file)
		}
	}

	if candidateName == "" {
		return "", errors.New("could not find a valid build artifact")
	}
	return candidateName, nil
}
