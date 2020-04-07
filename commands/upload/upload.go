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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/sketches"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/executils"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
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

	// FIXME: make a specification on how a port is specified via command line
	port := req.GetPort()
	if port == "" && sketch != nil && sketch.Metadata != nil {
		deviceURI, err := url.Parse(sketch.Metadata.CPU.Port)
		if err != nil {
			return nil, fmt.Errorf("invalid Device URL format: %s", err)
		}
		if deviceURI.Scheme == "serial" {
			port = deviceURI.Host + deviceURI.Path
		}
	}
	if port == "" {
		return nil, fmt.Errorf("no upload port provided")
	}

	fqbnIn := req.GetFqbn()
	if fqbnIn == "" && sketch != nil && sketch.Metadata != nil {
		fqbnIn = sketch.Metadata.CPU.Fqbn
	}
	if fqbnIn == "" {
		return nil, fmt.Errorf("no Fully Qualified Board Name provided")
	}
	fqbn, err := cores.ParseFQBN(fqbnIn)
	if err != nil {
		return nil, fmt.Errorf("incorrect FQBN: %s", err)
	}

	pm := commands.GetPackageManager(req.GetInstance().GetId())

	// Find target board and board properties
	_, _, board, boardProperties, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		return nil, fmt.Errorf("incorrect FQBN: %s", err)
	}

	// Load programmer tool
	uploadToolPattern, have := boardProperties.GetOk("upload.tool")
	if !have || uploadToolPattern == "" {
		return nil, fmt.Errorf("cannot get programmer tool: undefined 'upload.tool' property")
	}

	var referencedPlatformRelease *cores.PlatformRelease
	if split := strings.Split(uploadToolPattern, ":"); len(split) > 2 {
		return nil, fmt.Errorf("invalid 'upload.tool' property: %s", uploadToolPattern)
	} else if len(split) == 2 {
		referencedPackageName := split[0]
		uploadToolPattern = split[1]
		architecture := board.PlatformRelease.Platform.Architecture

		if referencedPackage := pm.Packages[referencedPackageName]; referencedPackage == nil {
			return nil, fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
		} else if referencedPlatform := referencedPackage.Platforms[architecture]; referencedPlatform == nil {
			return nil, fmt.Errorf("required platform %s:%s not installed", referencedPackageName, architecture)
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

	// Set properties for verify
	Verify := req.GetVerify()
	if Verify {
		uploadProperties.Set("upload.verify", uploadProperties.Get("upload.params.verify"))
	} else {
		uploadProperties.Set("upload.verify", uploadProperties.Get("upload.params.noverify"))
	}

	// Set path to compiled binary
	// Make the filename without the FQBN configs part
	fqbn.Configs = properties.NewMap()
	fqbnSuffix := strings.Replace(fqbn.String(), ":", ".", -1)

	var importPath *paths.Path
	var importFile string
	// If no importFile is passed, use sketch path
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
		return nil, fmt.Errorf("property 'recipe.output.tmp_file' not defined")
	}

	ext := filepath.Ext(outputTmpFile)
	if strings.HasSuffix(importFile, ext) {
		importFile = importFile[:len(importFile)-len(ext)]
	}

	// Check if the file ext we calculate is the same that is needed by the upload recipe
	recipet := uploadProperties.Get("upload.pattern")
	cmdLinet := uploadProperties.ExpandPropsInString(recipet)
	cmdArgst, err := properties.SplitQuotedString(cmdLinet, `"'`, false)
	var tPath *paths.Path
	if err != nil {
		return nil, fmt.Errorf("invalid recipe '%s': %s", recipet, err)
	}
	for _, t := range cmdArgst {
		if strings.Contains(t, "build.project_name") {
			tPath = paths.New(t)
		}
	}

	if ext != tPath.Ext() {
		ext = tPath.Ext()
	}
	//uploadRecipeInputFileExt :=
	uploadProperties.SetPath("build.path", importPath)
	uploadProperties.Set("build.project_name", importFile)
	uploadFile := importPath.Join(importFile + ext)
	if _, err := uploadFile.Stat(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot open sketch: %s", err)
		}
		// Built sketch not found in the provided path, let's fallback to the temp compile path
		var fallbackBuildPath *paths.Path
		if req.GetBuildPath() != "" {
			fallbackBuildPath = paths.New(req.GetBuildPath())
		} else {

			fallbackBuildPath = builder.GenBuildPath(sketchPath)
		}

		logrus.Warnf("Built sketch not found in %s, let's fallback to %s", uploadFile, fallbackBuildPath)
		uploadProperties.SetPath("build.path", fallbackBuildPath)
		// If we search inside the build.path, compile artifact do not have the fqbnSuffix in the filename
		uploadFile = fallbackBuildPath.Join(sketch.Name + ".ino" + ext)
		if _, err := uploadFile.Stat(); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("compiled sketch %s not found", uploadFile.String())
			}
			return nil, fmt.Errorf("cannot open sketch: %s", err)
		}
		// Clean from extension
		uploadProperties.Set("build.project_name", sketch.Name+".ino")

	}

	// Perform reset via 1200bps touch if requested
	if uploadProperties.GetBoolean("upload.use_1200bps_touch") {
		ports, err := serial.GetPortsList()
		if err != nil {
			return nil, fmt.Errorf("cannot get serial port list: %s", err)
		}
		for _, p := range ports {
			if p == port {
				if err := touchSerialPortAt1200bps(p); err != nil {
					return nil, fmt.Errorf("cannot perform reset: %s", err)
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
	actualPort := port // default
	if uploadProperties.GetBoolean("upload.wait_for_upload_port") {
		if p, err := waitForNewSerialPort(); err != nil {
			return nil, fmt.Errorf("cannot detect serial ports: %s", err)
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

	// Set serial port property
	uploadProperties.Set("serial.port", actualPort)
	if strings.HasPrefix(actualPort, "/dev/") {
		uploadProperties.Set("serial.port.file", actualPort[5:])
	} else {
		uploadProperties.Set("serial.port.file", actualPort)
	}

	// Build recipe for upload
	recipe := uploadProperties.Get("upload.pattern")
	cmdLine := uploadProperties.ExpandPropsInString(recipe)
	cmdArgs, err := properties.SplitQuotedString(cmdLine, `"'`, false)
	if err != nil {
		return nil, fmt.Errorf("invalid recipe '%s': %s", recipe, err)
	}

	// Run Tool
	cmd, err := executils.Command(cmdArgs)
	if err != nil {
		return nil, fmt.Errorf("cannot execute upload tool: %s", err)
	}

	executils.AttachStdoutListener(cmd, executils.PrintToStdout)
	executils.AttachStderrListener(cmd, executils.PrintToStderr)
	cmd.Stdout = outStream
	cmd.Stderr = errStream

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot execute upload tool: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("uploading error: %s", err)
	}

	logrus.Tracef("Upload %s on %s successful", sketch.Name, fqbnIn)

	return &rpc.UploadResp{}, nil
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
