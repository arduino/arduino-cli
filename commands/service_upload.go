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

package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/commands/internal/instances"
	f "github.com/arduino/arduino-cli/internal/algorithms"
	"github.com/arduino/arduino-cli/internal/arduino/cores"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/internal/arduino/globals"
	"github.com/arduino/arduino-cli/internal/arduino/sketch"
	"github.com/arduino/arduino-cli/internal/i18n"
	"github.com/arduino/arduino-cli/pkg/fqbn"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	serialutils "github.com/arduino/go-serial-utils"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/sirupsen/logrus"
)

// SupportedUserFields returns a SupportedUserFieldsResponse containing all the UserFields supported
// by the upload tools needed by the board using the protocol specified in SupportedUserFieldsRequest.
func (s *arduinoCoreServerImpl) SupportedUserFields(ctx context.Context, req *rpc.SupportedUserFieldsRequest) (*rpc.SupportedUserFieldsResponse, error) {
	if req.GetProtocol() == "" {
		return nil, &cmderrors.MissingPortProtocolError{}
	}

	pme, release, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return nil, err
	}
	defer release()

	fqbn, err := fqbn.Parse(req.GetFqbn())
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}

	_, platformRelease, _, boardProperties, _, err := pme.ResolveFQBN(fqbn)
	if platformRelease == nil {
		return nil, &cmderrors.PlatformNotFoundError{
			Platform: fmt.Sprintf("%s:%s", fqbn.Vendor, fqbn.Architecture),
			Cause:    err,
		}
	} else if err != nil {
		return nil, &cmderrors.UnknownFQBNError{Cause: err}
	}

	toolID, err := getToolID(boardProperties, "upload", req.GetProtocol())
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

	return "", &cmderrors.MissingPlatformPropertyError{Property: toolProperty}
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

// UploadToServerStreams return a server stream that forwards the output and error streams to the provided writers.
// It also returns a function that can be used to retrieve the result of the upload.
func UploadToServerStreams(ctx context.Context, outStream io.Writer, errStream io.Writer) (rpc.ArduinoCoreService_UploadServer, func() *rpc.UploadResult) {
	var result *rpc.UploadResult
	stream := streamResponseToCallback(ctx, func(resp *rpc.UploadResponse) error {
		if errData := resp.GetErrStream(); len(errData) > 0 {
			_, err := errStream.Write(errData)
			return err
		}
		if outData := resp.GetOutStream(); len(outData) > 0 {
			_, err := outStream.Write(outData)
			return err
		}
		if res := resp.GetResult(); res != nil {
			result = res
		}
		return nil
	})
	return stream, func() *rpc.UploadResult {
		return result
	}
}

// Upload performs the upload of a sketch to a board.
func (s *arduinoCoreServerImpl) Upload(req *rpc.UploadRequest, stream rpc.ArduinoCoreService_UploadServer) error {
	syncSend := NewSynchronizedSend(stream.Send)

	logrus.Tracef("Upload %s on %s started", req.GetSketchPath(), req.GetFqbn())

	// TODO: make a generic function to extract sketch from request
	// and remove duplication in commands/compile.go
	sketchPath := paths.New(req.GetSketchPath())
	sk, err := sketch.New(sketchPath)
	if err != nil && req.GetImportDir() == "" && req.GetImportFile() == "" {
		return &cmderrors.CantOpenSketchError{Cause: err}
	}

	pme, pmeRelease, err := instances.GetPackageManagerExplorer(req.GetInstance())
	if err != nil {
		return err
	}
	defer pmeRelease()

	fqbn := req.GetFqbn()
	if fqbn == "" && pme.GetProfile() != nil {
		fqbn = pme.GetProfile().FQBN
	}

	programmer := req.GetProgrammer()
	if programmer == "" && pme.GetProfile() != nil {
		programmer = pme.GetProfile().Programmer
	}
	if programmer == "" && sk != nil {
		programmer = sk.GetDefaultProgrammer()
	}

	outStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_OutStream{OutStream: data},
		})
	})
	defer outStream.Close()
	errStream := feedStreamTo(func(data []byte) {
		syncSend.Send(&rpc.UploadResponse{
			Message: &rpc.UploadResponse_ErrStream{ErrStream: data},
		})
	})
	defer errStream.Close()
	updatedPort, err := s.runProgramAction(
		stream.Context(),
		pme,
		sk,
		req.GetImportFile(),
		req.GetImportDir(),
		fqbn,
		req.GetPort(),
		programmer,
		req.GetVerbose(),
		req.GetVerify(),
		false, // burnBootloader
		outStream,
		errStream,
		req.GetDryRun(),
		req.GetUserFields(),
		req.GetUploadProperties(),
	)
	if err != nil {
		return err
	}
	return syncSend.Send(&rpc.UploadResponse{
		Message: &rpc.UploadResponse_Result{
			Result: &rpc.UploadResult{
				UpdatedUploadPort: updatedPort,
			},
		},
	})
}

// UploadUsingProgrammer FIXMEDOC
func (s *arduinoCoreServerImpl) UploadUsingProgrammer(req *rpc.UploadUsingProgrammerRequest, stream rpc.ArduinoCoreService_UploadUsingProgrammerServer) error {
	syncSend := NewSynchronizedSend(stream.Send)
	streamAdapter := streamResponseToCallback(stream.Context(), func(resp *rpc.UploadResponse) error {
		if errData := resp.GetErrStream(); len(errData) > 0 {
			syncSend.Send(&rpc.UploadUsingProgrammerResponse{
				Message: &rpc.UploadUsingProgrammerResponse_ErrStream{
					ErrStream: errData,
				},
			})
		}
		if outData := resp.GetOutStream(); len(outData) > 0 {
			syncSend.Send(&rpc.UploadUsingProgrammerResponse{
				Message: &rpc.UploadUsingProgrammerResponse_OutStream{
					OutStream: outData,
				},
			})
		}
		// resp.GetResult() is ignored
		return nil
	})

	logrus.Tracef("Upload using programmer %s on %s started", req.GetSketchPath(), req.GetFqbn())

	if req.GetProgrammer() == "" {
		return &cmderrors.MissingProgrammerError{}
	}
	return s.Upload(&rpc.UploadRequest{
		Instance:         req.GetInstance(),
		SketchPath:       req.GetSketchPath(),
		ImportFile:       req.GetImportFile(),
		ImportDir:        req.GetImportDir(),
		Fqbn:             req.GetFqbn(),
		Port:             req.GetPort(),
		Programmer:       req.GetProgrammer(),
		Verbose:          req.GetVerbose(),
		Verify:           req.GetVerify(),
		UserFields:       req.GetUserFields(),
		DryRun:           req.GetDryRun(),
		UploadProperties: req.GetUploadProperties(),
	}, streamAdapter)
}

func (s *arduinoCoreServerImpl) runProgramAction(ctx context.Context, pme *packagemanager.Explorer,
	sk *sketch.Sketch,
	importFile, importDir, fqbnIn string, userPort *rpc.Port,
	programmerID string,
	verbose, verify, burnBootloader bool,
	outStream, errStream io.Writer,
	dryRun bool, userFields map[string]string,
	requestUploadProperties []string,
) (*rpc.Port, error) {
	port := rpc.DiscoveryPortFromRPCPort(userPort)
	if port == nil || (port.Address == "" && port.Protocol == "") {
		// For no-port uploads use "default" protocol
		port = &discovery.Port{Protocol: "default"}
	}
	logrus.WithField("port", port).Tracef("Upload port")

	if burnBootloader && programmerID == "" {
		return nil, &cmderrors.MissingProgrammerError{}
	}

	fqbn, err := fqbn.Parse(fqbnIn)
	if err != nil {
		return nil, &cmderrors.InvalidFQBNError{Cause: err}
	}
	logrus.WithField("fqbn", fqbn).Tracef("Detected FQBN")

	// Find target board and board properties
	_, boardPlatform, board, boardProperties, buildPlatform, err := pme.ResolveFQBN(fqbn)
	if boardPlatform == nil {
		return nil, &cmderrors.PlatformNotFoundError{
			Platform: fmt.Sprintf("%s:%s", fqbn.Vendor, fqbn.Architecture),
			Cause:    err,
		}
	} else if err != nil {
		return nil, &cmderrors.UnknownFQBNError{Cause: err}
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
			return nil, &cmderrors.ProgrammerNotFoundError{Programmer: programmerID}
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
		return nil, err
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
		return nil, &cmderrors.InvalidPlatformPropertyError{
			Property: fmt.Sprintf("%s.tool.%s", action, port.Protocol), // TODO: Can be done better, maybe inline getToolID(...)
			Value:    uploadToolID}
	} else if len(split) == 2 {
		p := pme.FindPlatform(&packagemanager.PlatformReference{
			Package:              split[0],
			PlatformArchitecture: boardPlatform.Platform.Architecture,
		})
		if p == nil {
			return nil, &cmderrors.PlatformNotFoundError{Platform: split[0] + ":" + boardPlatform.Platform.Architecture}
		}
		uploadToolID = split[1]
		uploadToolPlatform = pme.GetBestInstalledPlatformRelease(p)
		if uploadToolPlatform == nil {
			return nil, &cmderrors.PlatformNotFoundError{Platform: split[0] + ":" + boardPlatform.Platform.Architecture}
		}
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

	// Add user provided custom upload properties
	if p, err := properties.LoadFromSlice(requestUploadProperties); err == nil {
		uploadProperties.Merge(p)
	} else {
		return nil, fmt.Errorf("invalid build properties: %w", err)
	}

	// Certain tools require the user to provide custom fields at run time,
	// if they've been provided set them
	// For more info:
	// https://arduino.github.io/arduino-cli/latest/platform-specification/#user-provided-fields
	for name, value := range userFields {
		uploadProperties.Set(fmt.Sprintf("%s.field.%s", action, name), value)
	}

	if !uploadProperties.ContainsKey("upload.protocol") && programmer == nil {
		return nil, &cmderrors.ProgrammerRequiredForUploadError{}
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
		importPath, sketchName, err := s.determineBuildPathAndSketchName(importFile, importDir, sk)
		if err != nil {
			return nil, &cmderrors.NotFoundError{Message: i18n.Tr("Error finding build artifacts"), Cause: err}
		}
		if !importPath.Exist() {
			return nil, &cmderrors.NotFoundError{Message: i18n.Tr("Compiled sketch not found in %s", importPath)}
		}
		if !importPath.IsDir() {
			return nil, &cmderrors.NotFoundError{Message: i18n.Tr("Expected compiled sketch in directory %s, but is a file instead", importPath)}
		}
		uploadProperties.SetPath("build.path", importPath)
		uploadProperties.Set("build.project_name", sketchName)
	}

	// This context is kept alive for the entire duration of the upload
	uploadCtx, uploadCompleted := context.WithCancel(ctx)
	defer uploadCompleted()

	// Start the upload port change detector.
	watcher, err := pme.DiscoveryManager().Watch()
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	updatedUploadPort := f.NewFuture[*discovery.Port]()
	go detectUploadPort(
		uploadCtx,
		port, watcher.Feed(),
		uploadProperties.GetBoolean("upload.wait_for_upload_port"),
		updatedUploadPort)

	// Force port wait to make easier to unbrick boards like the Arduino Leonardo, or similar with native USB,
	// when a sketch causes a crash and the native USB serial port is lost.
	// See https://github.com/arduino/arduino-cli/issues/1943 for the details.
	//
	// In order to trigger the forced serial-port-wait the following conditions must be met:
	// - No upload port specified (protocol == "default")
	// - "upload.wait_for_upload_port" == true (developers requested the touch + port wait)
	// - "upload.tool.serial" not defained, or
	//   "upload.tool.serial" is the same as "upload.tool.default"
	forcedSerialPortWait := port.Protocol == "default" && // this is the value when no port is specified
		uploadProperties.GetBoolean("upload.wait_for_upload_port") &&
		(!uploadProperties.ContainsKey("upload.tool.serial") ||
			uploadProperties.Get("upload.tool.serial") == uploadProperties.Get("upload.tool.default"))

	// If not using programmer perform some action required
	// to set the board in bootloader mode
	actualPort := port.Clone()
	if programmer == nil && !burnBootloader && (port.Protocol == "serial" || forcedSerialPortWait) {
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
			fmt.Fprintln(outStream, i18n.Tr("Skipping 1200-bps touch reset: no serial port selected!"))
		}

		cb := &serialutils.ResetProgressCallbacks{
			TouchingPort: func(portAddress string) {
				logrus.WithField("phase", "board reset").Infof("Performing 1200-bps touch reset on serial port %s", portAddress)
				if verbose {
					fmt.Fprintln(outStream, i18n.Tr("Performing 1200-bps touch reset on serial port %s", portAddress))
				}
			},
			WaitingForNewSerial: func() {
				logrus.WithField("phase", "board reset").Info("Waiting for upload port...")
				if verbose {
					fmt.Fprintln(outStream, i18n.Tr("Waiting for upload port..."))
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
						fmt.Fprintln(outStream, i18n.Tr("Upload port found on %s", portAddress))
					} else {
						fmt.Fprintln(outStream, i18n.Tr("No upload port found, using %s as fallback", actualPort.Address))
					}
				}
			},
			Debug: func(msg string) {
				logrus.WithField("phase", "board reset").Debug(msg)
			},
		}

		if newPortAddress, err := serialutils.Reset(portToTouch, wait, dryRun, nil, cb); err != nil {
			fmt.Fprintln(errStream, i18n.Tr("Cannot perform port reset: %s", err))
		} else {
			if newPortAddress != "" {
				actualPort.Address = newPortAddress
				actualPort.AddressLabel = newPortAddress
			}
		}
	}

	if actualPort.Address != "" {
		// Set serial port property
		uploadProperties.Set("serial.port", actualPort.Address)
		if actualPort.Protocol == "serial" || actualPort.Protocol == "default" {
			// This must be done only for serial ports
			portFile := strings.TrimPrefix(actualPort.Address, "/dev/")
			uploadProperties.Set("serial.port.file", portFile)
		}
	}

	// Get Port properties gathered using pluggable discovery
	uploadProperties.Set("upload.port.address", port.Address)
	uploadProperties.Set("upload.port.label", port.AddressLabel)
	uploadProperties.Set("upload.port.protocol", port.Protocol)
	uploadProperties.Set("upload.port.protocolLabel", port.ProtocolLabel)
	if actualPort.Properties != nil {
		for prop, value := range actualPort.Properties.AsMap() {
			uploadProperties.Set(fmt.Sprintf("upload.port.properties.%s", prop), value)
		}
	}

	// Run recipes for upload
	toolEnv := pme.GetEnvVarsForSpawnedProcess()
	if burnBootloader {
		if err := runTool(uploadCtx, "erase.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return nil, &cmderrors.FailedUploadError{Message: i18n.Tr("Failed chip erase"), Cause: err}
		}
		if err := runTool(uploadCtx, "bootloader.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return nil, &cmderrors.FailedUploadError{Message: i18n.Tr("Failed to burn bootloader"), Cause: err}
		}
	} else if programmer != nil {
		if err := runTool(uploadCtx, "program.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return nil, &cmderrors.FailedUploadError{Message: i18n.Tr("Failed programming"), Cause: err}
		}
	} else {
		if err := runTool(uploadCtx, "upload.pattern", uploadProperties, outStream, errStream, verbose, dryRun, toolEnv); err != nil {
			return nil, &cmderrors.FailedUploadError{Message: i18n.Tr("Failed uploading"), Cause: err}
		}
	}

	uploadCompleted()
	logrus.Tracef("Upload successful")

	updatedPort := updatedUploadPort.Await()
	if updatedPort == nil {
		// If the algorithms can not detect the new port, fallback to the user-provided port.
		return userPort, nil
	}
	return rpc.DiscoveryPortToRPC(updatedPort), nil
}

func detectUploadPort(
	uploadCtx context.Context,
	uploadPort *discovery.Port, watch <-chan *discovery.Event,
	waitForUploadPort bool,
	result f.Future[*discovery.Port],
) {
	log := logrus.WithField("task", "port_detection")
	log.Debugf("Detecting new board port after upload")

	candidate := uploadPort.Clone()
	defer func() {
		result.Send(candidate)
	}()

	// Ignore all events during the upload
	for {
		select {
		case ev, ok := <-watch:
			if !ok {
				log.Error("Upload port detection failed, watcher closed")
				return
			}
			if candidate != nil && ev.Type == "remove" && ev.Port.Equals(candidate) {
				log.WithField("event", ev).Debug("User-specified port has been disconnected, forcing wait for upload port")
				waitForUploadPort = true
				candidate = nil
			} else {
				log.WithField("event", ev).Debug("Ignored watcher event before upload")
			}
			continue
		case <-uploadCtx.Done():
			// Upload completed, move to the next phase
		}
		break
	}

	// Pick the first port that is detected after the upload
	timeout := time.After(5 * time.Second)
	if !waitForUploadPort {
		timeout = time.After(time.Second)
	}
	for {
		select {
		case ev, ok := <-watch:
			if !ok {
				log.Error("Upload port detection failed, watcher closed")
				return
			}
			if candidate != nil && ev.Type == "remove" && candidate.Equals(ev.Port) {
				log.WithField("event", ev).Debug("Candidate port is no longer available")
				candidate = nil
				if !waitForUploadPort {
					waitForUploadPort = true
					timeout = time.After(5 * time.Second)
					log.Debug("User-specified port has been disconnected, now waiting for upload port, timeout extended by 5 seconds")
				}
				continue
			}
			if ev.Type != "add" {
				log.WithField("event", ev).Debug("Ignored non-add event")
				continue
			}

			portPriority := func(port *discovery.Port) int {
				if port == nil {
					return 0
				}
				prio := 0
				if port.HardwareID == uploadPort.HardwareID {
					prio += 1000
				}
				if port.Protocol == uploadPort.Protocol {
					prio += 100
				}
				if port.Address == uploadPort.Address {
					prio += 10
				}
				return prio
			}
			evPortPriority := portPriority(ev.Port)
			candidatePriority := portPriority(candidate)
			if evPortPriority <= candidatePriority {
				log.WithField("event", ev).Debugf("New upload port candidate is worse than the current one (prio=%d)", evPortPriority)
				continue
			}
			log.WithField("event", ev).Debugf("Found new upload port candidate (prio=%d)", evPortPriority)
			candidate = ev.Port

			// If the current candidate have the desired HW-ID return it quickly.
			if candidate.HardwareID == ev.Port.HardwareID {
				timeout = time.After(time.Second)
				log.Debug("New candidate port match the desired HW ID, timeout reduced to 1 second.")
				continue
			}

		case <-timeout:
			log.WithField("selected_port", candidate).Debug("Timeout waiting for candidate port")
			return
		}
	}
}

func runTool(ctx context.Context, recipeID string, props *properties.Map, outStream, errStream io.Writer, verbose bool, dryRun bool, toolEnv []string) error {
	// if ctx is already canceled just exit
	if err := ctx.Err(); err != nil {
		return err
	}

	recipe, ok := props.GetOk(recipeID)
	if !ok {
		return errors.New(i18n.Tr("recipe not found '%s'", recipeID))
	}
	if strings.TrimSpace(recipe) == "" {
		return nil // Nothing to run
	}
	if props.IsPropertyMissingInExpandPropsInString("serial.port", recipe) || props.IsPropertyMissingInExpandPropsInString("serial.port.file", recipe) {
		return errors.New(i18n.Tr("no upload port provided"))
	}
	cmdLine := props.ExpandPropsInString(recipe)
	cmdArgs, _ := properties.SplitQuotedString(cmdLine, `"'`, false)

	// Run Tool
	logrus.WithField("phase", "upload").Tracef("Executing upload tool: %s", cmdLine)
	if verbose {
		fmt.Fprintln(outStream, cmdLine)
	}
	if dryRun {
		return nil
	}
	cmd, err := paths.NewProcess(toolEnv, cmdArgs...)
	if err != nil {
		return errors.New(i18n.Tr("cannot execute upload tool: %s", err))
	}

	cmd.RedirectStdoutTo(outStream)
	cmd.RedirectStderrTo(errStream)

	if err := cmd.Start(); err != nil {
		return errors.New(i18n.Tr("cannot execute upload tool: %s", err))
	}

	// If the ctx is canceled, kill the running command
	completed := make(chan struct{})
	defer close(completed)
	go func() {
		select {
		case <-ctx.Done():
			_ = cmd.Kill()
		case <-completed:
		}
	}()

	if err := cmd.Wait(); err != nil {
		return errors.New(i18n.Tr("uploading error: %s", err))
	}

	return nil
}

func (s *arduinoCoreServerImpl) determineBuildPathAndSketchName(importFile, importDir string, sk *sketch.Sketch) (*paths.Path, string, error) {
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
			return nil, "", errors.New(i18n.Tr("%s and %s cannot be used together", "importFile", "importDir"))
		}

		// We have a path like "path/to/my/build/SketchName.ino.bin". We are going to
		// ignore the extension and set:
		// - "build.path" as "path/to/my/build"
		// - "build.project_name" as "SketchName.ino"

		importFilePath := paths.New(importFile)
		if !importFilePath.Exist() {
			return nil, "", errors.New(i18n.Tr("binary file not found in %s", importFilePath))
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
			return nil, "", fmt.Errorf("%s: %w", i18n.Tr("looking for build artifacts"), err)
		}
		return buildPath, sketchName, nil
	}

	// Case 3: nothing given...
	if sk == nil {
		return nil, "", errors.New(i18n.Tr("no sketch or build directory/file specified"))
	}

	// Case 4: only sketch specified. In this case we use the generated build path
	// and the given sketch name.
	return s.getDefaultSketchBuildPath(sk, nil), sk.Name + sk.MainFile.Ext(), nil
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
		if !globals.MainFileValidExtensions[filepath.Ext(name)] {
			// just ignore those files
			continue
		}

		if candidateName == "" {
			candidateName = name
			candidateFile = file
		}

		if candidateName != name {
			return "", errors.New(i18n.Tr("multiple build artifacts found: '%[1]s' and '%[2]s'", candidateFile, file))
		}
	}

	if candidateName == "" {
		return "", errors.New(i18n.Tr("could not find a valid build artifact"))
	}
	return candidateName, nil
}

// overrideProtocolProperties returns a copy of props overriding action properties with
// specified protocol properties.
//
// For example passing the below properties and "upload" as action and "serial" as protocol:
//
//	upload.speed=256
//	upload.serial.speed=57600
//	upload.network.speed=19200
//
// will return:
//
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
