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
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/arduino/serialutils"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var tr = i18n.Tr

// SupportedUserFields returns a SupportedUserFieldsResponse containing all the UserFields supported
// by the upload tools needed by the board using the protocol specified in SupportedUserFieldsRequest.
func SupportedUserFields(ctx context.Context, req *rpc.SupportedUserFieldsRequest) (*rpc.SupportedUserFieldsResponse, error) {
	if req.Protocol == "" {
		return nil, &arduino.MissingPortProtocolError{}
	}

	pme, release := commands.GetPackageManagerExplorer(req)
	defer release()

	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}

	fqbn, err := cores.ParseFQBN(req.GetFqbn())
	if err != nil {
		return nil, &arduino.InvalidFQBNError{Cause: err}
	}

	_, platformRelease, _, boardProperties, _, err := pme.ResolveFQBN(fqbn)
	if platformRelease == nil {
		return nil, &arduino.PlatformNotFoundError{
			Platform: fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch),
			Cause:    err,
		}
	} else if err != nil {
		return nil, &arduino.UnknownFQBNError{Cause: err}
	}

	toolID, err := getToolID(boardProperties, "upload", req.Protocol)
	if err != nil {
		return nil, err
	}

	return &rpc.SupportedUserFieldsResponse{
		UserFields: getUserFields(toolID, platformRelease),
	}, nil
}

// getToolID returns the ID of the tool that supports the action and protocol combination by searching in props.
// Returns error if tool cannot be found.
func getToolID(props *properties.Map, action, protocol string) (string, error) {
	toolProperty := fmt.Sprintf("%s.tool.%s", action, protocol)
	defaultToolProperty := fmt.Sprintf("%s.tool.default", action)

	if t, ok := props.GetOk(toolProperty); ok {
		return t, nil
	}

	if t, ok := props.GetOk(defaultToolProperty); ok {
		// Fallback for platform that don't support the specified protocol for specified action:
		// https://arduino.github.io/arduino-cli/latest/platform-specification/#sketch-upload-configuration
		return t, nil
	}

	return "", &arduino.MissingPlatformPropertyError{Property: toolProperty}
}

// getUserFields return all user fields supported by the tools specified.
// Returns error only in case the secret property is not a valid boolean.
func getUserFields(toolID string, platformRelease *cores.PlatformRelease) []*rpc.UserField {
	userFields := []*rpc.UserField{}
	fields := platformRelease.Properties.SubTree(fmt.Sprintf("tools.%s.upload.field", toolID))
	keys := fields.FirstLevelKeys()

	for _, key := range keys {
		value := fields.Get(key)
		if len(value) > 50 {
			value = fmt.Sprintf("%s…", value[:49])
		}
		isSecret := fields.GetBoolean(fmt.Sprintf("%s.secret", key))
		userFields = append(userFields, &rpc.UserField{
			ToolId: toolID,
			Name:   key,
			Label:  value,
			Secret: isSecret,
		})
	}

	return userFields
}

// Upload FIXMEDOC
func Upload(ctx context.Context, req *rpc.UploadRequest, outStream io.Writer, errStream io.Writer) (*rpc.UploadResponse, error) {
	logrus.Tracef("Upload %s on %s started", req.GetSketchPath(), req.GetFqbn())

	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil && req.GetImportDir() == "" && req.GetImportFile() == "" {
		return nil, &arduino.CantOpenSketchError{Cause: err}
	}

	pme, release := commands.GetPackageManagerExplorer(req)
	if pme == nil {
		return nil, &arduino.InvalidInstanceError{}
	}
	defer release()

	if err := runProgramAction(
		pme,
		sk,
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
		req.GetDryRun(),
		req.GetUserFields(),
	); err != nil {
		return nil, err
	}

	return &rpc.UploadResponse{}, nil
}

// UsingProgrammer FIXMEDOC
func UsingProgrammer(ctx context.Context, req *rpc.UploadUsingProgrammerRequest, outStream io.Writer, errStream io.Writer) (*rpc.UploadUsingProgrammerResponse, error) {
	logrus.Tracef("Upload using programmer %s on %s started", req.GetSketchPath(), req.GetFqbn())

	if req.GetProgrammer() == "" {
		return nil, &arduino.MissingProgrammerError{}
	}
	_, err := Upload(ctx, &rpc.UploadRequest{
		Instance:   req.GetInstance(),
		SketchPath: req.GetSketchPath(),
		ImportFile: req.GetImportFile(),
		ImportDir:  req.GetImportDir(),
		Fqbn:       req.GetFqbn(),
		Port:       req.GetPort(),
		Programmer: req.GetProgrammer(),
		Verbose:    req.GetVerbose(),
		Verify:     req.GetVerify(),
		UserFields: req.GetUserFields(),
	}, outStream, errStream)
	return &rpc.UploadUsingProgrammerResponse{}, err
}

func runProgramAction(pme *packagemanager.Explorer,
	sk *sketch.Sketch,
	importFile, importDir, fqbnIn string, port *rpc.Port,
	programmerID string,
	verbose, verify, burnBootloader bool,
	outStream, errStream io.Writer,
	dryRun bool, userFields map[string]string) error {

	if burnBootloader && programmerID == "" {
		return &arduino.MissingProgrammerError{}
	}

	logrus.WithField("port", port).Tracef("Upload port")

	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return &arduino.InvalidFQBNError{Cause: err}
	}
	logrus.WithField("fqbn", fqbn).Tracef("Detected FQBN")

	// Find target board and board properties
	_, boardPlatform, board, boardProperties, buildPlatform, err := pme.ResolveFQBN(fqbn)
	if boardPlatform == nil {
		return &arduino.PlatformNotFoundError{
			Platform: fmt.Sprintf("%s:%s", fqbn.Package, fqbn.PlatformArch),
			Cause:    err,
		}
	} else if err != nil {
		return &arduino.UnknownFQBNError{Cause: err}
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
			return &arduino.ProgrammerNotFoundError{Programmer: programmerID}
		}
	}

	// Determine upload tool
	// create a temporary configuration only for the selection of upload tool
	props := properties.NewMap()
	props.Merge(boardPlatform.Properties)
	props.Merge(boardPlatform.RuntimeProperties())
	props.Merge(boardProperties)
	if programmer != nil {
		props.Merge(programmer.Properties)
	}
	action := "upload"
	if burnBootloader {
		action = "bootloader"
	} else if programmer != nil {
		action = "program"
	}
	uploadToolID, err := getToolID(props, action, port.Protocol)
	if err != nil {
		return err
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
		return &arduino.InvalidPlatformPropertyError{
			Property: fmt.Sprintf("%s.tool.%s", action, port.Protocol), // TODO: Can be done better, maybe inline getToolID(...)
			Value:    uploadToolID}
	} else if len(split) == 2 {
		uploadToolID = split[1]
		uploadToolPlatform = pme.GetInstalledPlatformRelease(
			pme.FindPlatform(&packagemanager.PlatformReference{
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
	uploadProperties.Set("runtime.os", properties.GetOSSuffix())
	uploadProperties.Merge(boardPlatform.Properties)
	uploadProperties.Merge(boardPlatform.RuntimeProperties())
	uploadProperties.Merge(overrideProtocolProperties(action, port.Protocol, boardProperties))
	uploadProperties.Merge(uploadProperties.SubTree("tools." + uploadToolID))
	if programmer != nil {
		uploadProperties.Merge(programmer.Properties)
	}

	for _, tool := range pme.GetAllInstalledToolsReleases() {
		uploadProperties.Merge(tool.RuntimeProperties())
	}
	if requiredTools, err := pme.FindToolsRequiredForBoard(board); err == nil {
		for _, requiredTool := range requiredTools {
			logrus.WithField("tool", requiredTool).Info("Tool required for upload")
			if requiredTool.IsInstalled() {
				uploadProperties.Merge(requiredTool.RuntimeProperties())
			} else {
				errStream.Write([]byte(tr("Warning: tool '%s' is not installed. It might not be available for your OS.", requiredTool)))
			}
		}
	}

	// Certain tools require the user to provide custom fields at run time,
	// if they've been provided set them
	// For more info:
	// https://arduino.github.io/arduino-cli/latest/platform-specification/#user-provided-fields
	for name, value := range userFields {
		uploadProperties.Set(fmt.Sprintf("%s.field.%s", action, name), value)
	}

	if !uploadProperties.ContainsKey("upload.protocol") && programmer == nil {
		return &arduino.ProgrammerRequiredForUploadError{}
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
		importPath, sketchName, err := determineBuildPathAndSketchName(importFile, importDir, sk, fqbn)
		if err != nil {
			return &arduino.NotFoundError{Message: tr("Error finding build artifacts"), Cause: err}
		}
		if !importPath.Exist() {
			return &arduino.NotFoundError{Message: tr("Compiled sketch not found in %s", importPath)}
		}
		if !importPath.IsDir() {
			return &arduino.NotFoundError{Message: tr("Expected compiled sketch in directory %s, but is a file instead", importPath)}
		}
		uploadProperties.SetPath("build.path", importPath)
		uploadProperties.Set("build.project_name", sketchName)
	}

	// If not using programmer perform some action required
	// to set the board in bootloader mode
	actualPort := port
	if programmer == nil && !burnBootloader && port.Protocol == "serial" {

		// Perform reset via 1200bps touch if requested and wait for upload port also if requested.
		touch := uploadProperties.GetBoolean("upload.use_1200bps_touch")
		wait := false
		portToTouch := ""
		if touch {
			portToTouch = port.Address
			// Waits for upload port only if a 1200bps touch is done
			wait = uploadProperties.GetBoolean("upload.wait_for_upload_port")
		}

		// if touch is requested but port is not specified, print a warning
		if touch && portToTouch == "" {
			outStream.Write([]byte(fmt.Sprintln(tr("Skipping 1200-bps touch reset: no serial port selected!"))))
		}

		cb := &serialutils.ResetProgressCallbacks{
			TouchingPort: func(portAddress string) {
				logrus.WithField("phase", "board reset").Infof("Performing 1200-bps touch reset on serial port %s", portAddress)
				if verbose {
					outStream.Write([]byte(fmt.Sprintln(tr("Performing 1200-bps touch reset on serial port %s", portAddress))))
				}
			},
			WaitingForNewSerial: func() {
				logrus.WithField("phase", "board reset").Info("Waiting for upload port...")
				if verbose {
					outStream.Write([]byte(fmt.Sprintln(tr("Waiting for upload port..."))))
				}
			},
			BootloaderPortFound: func(portAddress string) {
				if portAddress != "" {
					logrus.WithField("phase", "board reset").Infof("Upload port found on %s", portAddress)
				} else {
					logrus.WithField("phase", "board reset").Infof("No upload port found, using %s as fallback", actualPort.Address)
				}
				if verbose {
					if portAddress != "" {
						outStream.Write([]byte(fmt.Sprintln(tr("Upload port found on %s", portAddress))))
					} else {
						outStream.Write([]byte(fmt.Sprintln(tr("No upload port found, using %s as fallback", actualPort.Address))))
					}
				}
			},
			Debug: func(msg string) {
				logrus.WithField("phase", "board reset").Debug(msg)
			},
		}

		if newPortAddress, err := serialutils.Reset(portToTouch, wait, cb, dryRun); err != nil {
			outStream.Write([]byte(fmt.Sprintln(tr("Cannot perform port reset: %s", err))))
		} else {
			if newPortAddress != "" {
				actualPort.Address = newPortAddress
			}
		}
	}

	if actualPort.Address != "" {
		// Set serial port property
		uploadProperties.Set("serial.port", actualPort.Address)
		if actualPort.Protocol == "serial" {
			// This must be done only for serial ports
			portFile := strings.TrimPrefix(actualPort.Address, "/dev/")
			uploadProperties.Set("serial.port.file", portFile)
		}
	}

	// Get Port properties gathered using pluggable discovery
	uploadProperties.Set("upload.port.address", port.Address)
	uploadProperties.Set("upload.port.label", port.Label)
	uploadProperties.Set("upload.port.protocol", port.Protocol)
	uploadProperties.Set("upload.port.protocolLabel", port.ProtocolLabel)
	for prop, value := range actualPort.Properties {
		uploadProperties.Set(fmt.Sprintf("upload.port.properties.%s", prop), value)
	}

	// Run recipes for upload
	toolEnv := pme.GetEnvVarsForSpawnedProcess()
	if burnBootloader {
		if err := runTool("erase.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return &arduino.FailedUploadError{Message: tr("Failed chip erase"), Cause: err}
		}
		if err := runTool("bootloader.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return &arduino.FailedUploadError{Message: tr("Failed to burn bootloader"), Cause: err}
		}
	} else if programmer != nil {
		if err := runTool("program.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return &arduino.FailedUploadError{Message: tr("Failed programming"), Cause: err}
		}
	} else {
		if err := runTool("upload.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return &arduino.FailedUploadError{Message: tr("Failed uploading"), Cause: err}
		}
	}

	logrus.Tracef("Upload successful")
	return nil
}

func runTool(recipeID string, props *properties.Map, outStream, errStream io.Writer, verbose bool, dryRun bool, toolEnv []string) error {
	recipe, ok := props.GetOk(recipeID)
	if !ok {
		return fmt.Errorf(tr("recipe not found '%s'"), recipeID)
	}
	if strings.TrimSpace(recipe) == "" {
		return nil // Nothing to run
	}
	if props.IsPropertyMissingInExpandPropsInString("serial.port", recipe) || props.IsPropertyMissingInExpandPropsInString("serial.port.file", recipe) {
		return fmt.Errorf(tr("no upload port provided"))
	}
	cmdLine := props.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return fmt.Errorf(tr("invalid recipe '%[1]s': %[2]s"), recipe, err)
	}

	// Run Tool
	logrus.WithField("phase", "upload").Tracef("Executing upload tool: %s", cmdLine)
	if verbose {
		outStream.Write([]byte(fmt.Sprintln(cmdLine)))
	}
	if dryRun {
		return nil
	}
	cmd, err := executils.NewProcess(toolEnv, cmdArgs...)
	if err != nil {
		return fmt.Errorf(tr("cannot execute upload tool: %s"), err)
	}

	cmd.RedirectStdoutTo(outStream)
	cmd.RedirectStderrTo(errStream)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf(tr("cannot execute upload tool: %s"), err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf(tr("uploading error: %s"), err)
	}

	return nil
}

func determineBuildPathAndSketchName(importFile, importDir string, sk *sketch.Sketch, fqbn *cores.FQBN) (*paths.Path, string, error) {
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
			return nil, "", fmt.Errorf(tr("%s and %s cannot be used together", "importFile", "importDir"))
		}

		// We have a path like "path/to/my/build/SketchName.ino.bin". We are going to
		// ignore the extension and set:
		// - "build.path" as "path/to/my/build"
		// - "build.project_name" as "SketchName.ino"

		importFilePath := paths.New(importFile)
		if !importFilePath.Exist() {
			return nil, "", fmt.Errorf(tr("binary file not found in %s"), importFilePath)
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
			return nil, "", errors.Errorf(tr("autodetect build artifact: %s"), err)
		}
		return buildPath, sketchName, nil
	}

	// Case 3: nothing given...
	if sk == nil {
		return nil, "", fmt.Errorf(tr("no sketch or build directory/file specified"))
	}

	// Case 4: only sketch specified. In this case we use the generated build path
	// and the given sketch name.
	return sk.BuildPath, sk.Name + sk.MainFile.Ext(), nil
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
			return "", errors.Errorf(tr("multiple build artifacts found: '%[1]s' and '%[2]s'"), candidateFile, file)
		}
	}

	if candidateName == "" {
		return "", errors.New(tr("could not find a valid build artifact"))
	}
	return candidateName, nil
}

// overrideProtocolProperties returns a copy of props overriding action properties with
// specified protocol properties.
//
// For example passing the below properties and "upload" as action and "serial" as protocol:
//	upload.speed=256
//	upload.serial.speed=57600
//	upload.network.speed=19200
//
// will return:
//	upload.speed=57600
//	upload.serial.speed=57600
//	upload.network.speed=19200
func overrideProtocolProperties(action, protocol string, props *properties.Map) *properties.Map {
	res := props.Clone()
	subtree := props.SubTree(fmt.Sprintf("%s.%s", action, protocol))
	for k, v := range subtree.AsMap() {
		res.Set(fmt.Sprintf("%s.%s", action, k), v)
	}
	return res
}
