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
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/cli/feedback"
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
	if req.GetSketchPath() == "" {
		return nil, fmt.Errorf("missing sketchPath")
	}
	sketchPath := paths.New(req.GetSketchPath())
	sketch, err := sketches.NewSketchFromPath(sketchPath)
	if err != nil {
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

func runProgramAction(pm *packagemanager.PackageManager,
	sketch *sketches.Sketch, importFile string, importDir string, fqbnIn string, port string,
	programmerID string,
	verbose, verify, burnBootloader bool,
	outStream io.Writer, errStream io.Writer) error {

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

	// Load upload tool definitions
	var uploadToolName string
	var uploadToolPlatform *cores.PlatformRelease
	var programmer *cores.Programmer

	if burnBootloader {
		uploadToolName = boardProperties.Get("bootloader.tool")
		uploadToolPlatform = boardPlatform
		if uploadToolName == "" {
			return fmt.Errorf("cannot get programmer tool: undefined 'bootloader.tool' in boards.txt")
		}
		logrus.
			WithField("uploadToolName", uploadToolName).
			WithField("uploadToolPlatform", uploadToolPlatform).
			Trace("Upload tool from 'bootloader.tool' property")
	}

	if programmerID != "" {
		programmer = boardPlatform.Programmers[programmerID]
		if programmer == nil {
			// Try to find the programmer in the referenced build platform
			programmer = buildPlatform.Programmers[programmerID]
		}
		if programmer == nil {
			return fmt.Errorf("programmer '%s' not available", programmerID)
		}
		uploadToolName = programmer.Properties.Get("program.tool")
		uploadToolPlatform = programmer.PlatformRelease
		if uploadToolName == "" {
			return fmt.Errorf("cannot get programmer tool: undefined 'program.tool' property")
		}
		logrus.
			WithField("uploadToolName", uploadToolName).
			WithField("uploadToolPlatform", uploadToolPlatform).
			Trace("Upload tool from --programmer parameter")
	} else {
		uploadToolName = boardProperties.Get("upload.tool")
		uploadToolPlatform = boardPlatform
		if uploadToolName == "" {
			return fmt.Errorf("cannot get upload tool: undefined 'upload.tool' property")
		}
		if split := strings.Split(uploadToolName, ":"); len(split) > 2 {
			return fmt.Errorf("invalid 'upload.tool' property: %s", uploadToolName)
		} else if len(split) == 2 {
			uploadToolName = split[1]
			uploadToolPlatform = pm.GetInstalledPlatformRelease(
				pm.FindPlatform(&packagemanager.PlatformReference{
					Package:              split[0],
					PlatformArchitecture: boardPlatform.Platform.Architecture,
				}),
			)
		}
		logrus.
			WithField("uploadToolName", uploadToolName).
			WithField("uploadToolPlatform", uploadToolPlatform).
			Trace("Upload tool")
	}

	// Build configuration for upload
	uploadProperties := properties.NewMap()
	if uploadToolPlatform != nil {
		uploadProperties.Merge(uploadToolPlatform.Properties)
	}
	uploadProperties.Merge(boardPlatform.Properties)
	uploadProperties.Merge(boardPlatform.RuntimeProperties())
	uploadProperties.Merge(boardProperties)

	uploadToolProperties := uploadProperties.SubTree("tools." + uploadToolName)
	uploadProperties.Merge(uploadToolProperties)
	if programmer != nil {
		uploadProperties.Merge(programmer.Properties)
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
	} else {
		uploadProperties.Set("upload.verify", uploadProperties.Get("upload.params.noverify"))
		uploadProperties.Set("program.verify", uploadProperties.Get("program.params.noverify"))
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
					if err := touchSerialPortAt1200bps(p); err != nil {
						outStream.Write([]byte(fmt.Sprintf("Cannot perform port reset: %s", err)))
						outStream.Write([]byte(fmt.Sprintln()))
					}
					break
				}
			}

			// Scanning for available ports seems to open the port or
			// otherwise assert DTR, which would cancel the WDT reset if
			// it happened within 250 ms. So we wait until the reset should
			// have already occurred before we start scanning.
			time.Sleep(500 * time.Millisecond)
		}

		// Wait for upload port if requested
		if uploadProperties.GetBoolean("upload.wait_for_upload_port") {
			if verbose {
				outStream.Write([]byte(fmt.Sprintln("Waiting for upload port...")))
			}
			if p, err := waitForNewSerialPort(); err != nil {
				return fmt.Errorf("cannot detect serial ports: %s", err)
			} else if p == "" {
				feedback.Print("No new serial port detected.")
			} else {
				actualPort = p
			}

			// on OS X, if the port is opened too quickly after it is detected,
			// a "Resource busy" error occurs, add a delay to workaround.
			// This apply to other platforms as well.
			time.Sleep(500 * time.Millisecond)
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

	// Build recipe for upload
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

func touchSerialPortAt1200bps(port string) error {
	logrus.Infof("Touching port %s at 1200bps", port)

	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return fmt.Errorf("opening port: %s", err)
	}
	defer p.Close()

	if err = p.SetDTR(false); err != nil {
		return fmt.Errorf("cannot set DTR")
	}
	return nil
}

// waitForNewSerialPort is meant to be called just after a reset. It watches the ports connected
// to the machine until a port appears. The new appeared port is returned
func waitForNewSerialPort() (string, error) {
	logrus.Infof("Waiting for upload port...")

	getPortMap := func() (map[string]bool, error) {
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, err
		}
		res := map[string]bool{}
		for _, port := range ports {
			res[port] = true
		}
		return res, nil
	}

	last, err := getPortMap()
	if err != nil {
		return "", fmt.Errorf("scanning serial port: %s", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		now, err := getPortMap()
		if err != nil {
			return "", fmt.Errorf("scanning serial port: %s", err)
		}

		for p := range now {
			if !last[p] {
				return p, nil // Found it!
			}
		}

		last = now
		time.Sleep(250 * time.Millisecond)
	}

	return "", nil
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
		// Case 2: importDir flag with a sketch
		if sketch != nil {
			// In this case we have both the build path and the sketch name given,
			// so we just return them as-is
			return paths.New(importDir), sketch.Name + ".ino", nil
		}

		// Case 3: importDir flag without a sketch

		// In this case we have a build path but the sketch name is not given, we may
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

	// Case 4: nothing given...
	if sketch == nil {
		return nil, "", fmt.Errorf("no sketch or build directory/file specified")
	}

	// Case 5: only sketch specified. In this case we use the default sketch build path
	// and the given sketch name.

	// TODO: Create a function to obtain importPath from sketch
	// Add FQBN (without configs part) to export path
	if fqbn == nil {
		return nil, "", fmt.Errorf("missing FQBN")
	}
	fqbnSuffix := strings.Replace(fqbn.StringWithoutConfig(), ":", ".", -1)
	return sketch.FullPath.Join("build").Join(fqbnSuffix), sketch.Name + ".ino", nil
}

func detectSketchNameFromBuildPath(buildPath *paths.Path) (string, error) {
	files, err := buildPath.ReadDir()
	if err != nil {
		return "", err
	}

	candidateName := ""
	var candidateFile *paths.Path
	for _, file := range files {
		// Build artifacts are usually names as "Blink.ino.hex" or "Blink.ino.bin".
		// Extract the "Blink.ino" part
		name := strings.TrimSuffix(file.Base(), file.Ext())

		// Sometimes we may have particular files like:
		// Blink.ino.with_bootloader.bin
		if filepath.Ext(name) != ".ino" {
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
