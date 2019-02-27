/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/common/formatter"
	"github.com/arduino/arduino-cli/executils"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	serial "go.bug.st/serial.v1"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   "Upload Arduino sketches.",
		Long:    "Upload Arduino sketches.",
		Example: "  " + cli.AppName + " upload /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}
	uploadCommand.Flags().StringVarP(
		&flags.fqbn, "fqbn", "b", "",
		"Fully Qualified Board Name, e.g.: arduino:avr:uno")
	uploadCommand.Flags().StringVarP(
		&flags.port, "port", "p", "",
		"Upload port, e.g.: COM10 or /dev/ttyACM0")
	uploadCommand.Flags().StringVarP(
		&flags.importFile, "input", "i", "",
		"Input file to be uploaded.")
	uploadCommand.Flags().BoolVarP(
		&flags.verify, "verify", "t", false,
		"Verify uploaded binary after the upload.")
	uploadCommand.Flags().BoolVarP(
		&flags.verbose, "verbose", "v", false,
		"Optional, turns on verbose mode.")
	return uploadCommand
}

var flags struct {
	fqbn       string
	port       string
	verbose    bool
	verify     bool
	importFile string
}

func run(command *cobra.Command, args []string) {
	var sketchPath *paths.Path
	if len(args) > 0 {
		sketchPath = paths.New(args[0])
	}
	sketch, err := cli.InitSketch(sketchPath)
	if err != nil {
		formatter.PrintError(err, "Error opening sketch.")
		os.Exit(cli.ErrGeneric)
	}

	// FIXME: make a specification on how a port is specified via command line
	port := flags.port
	if port == "" {
		formatter.PrintErrorMessage("No port provided.")
		os.Exit(cli.ErrBadCall)
	}

	if flags.fqbn == "" && sketch != nil {
		flags.fqbn = sketch.Metadata.CPU.Fqbn
	}
	if flags.fqbn == "" {
		formatter.PrintErrorMessage("No Fully Qualified Board Name provided.")
		os.Exit(cli.ErrBadCall)
	}
	fqbn, err := cores.ParseFQBN(flags.fqbn)
	if err != nil {
		formatter.PrintError(err, "Invalid FQBN.")
		os.Exit(cli.ErrBadCall)
	}

	pm, _ := cli.InitPackageAndLibraryManager()

	// Find target board and board properties
	_, _, board, boardProperties, _, err := pm.ResolveFQBN(fqbn)
	if err != nil {
		formatter.PrintError(err, "Invalid FQBN.")
		os.Exit(cli.ErrBadCall)
	}

	// Load programmer tool
	uploadToolPattern, have := boardProperties.GetOk("upload.tool")
	if !have || uploadToolPattern == "" {
		formatter.PrintErrorMessage("The board does not define an 'upload.tool' property.")
		os.Exit(cli.ErrGeneric)
	}

	var referencedPlatformRelease *cores.PlatformRelease
	if split := strings.Split(uploadToolPattern, ":"); len(split) > 2 {
		formatter.PrintErrorMessage("The board defines an invalid 'upload.tool' property: " + uploadToolPattern)
		os.Exit(cli.ErrGeneric)
	} else if len(split) == 2 {
		referencedPackageName := split[0]
		uploadToolPattern = split[1]
		architecture := board.PlatformRelease.Platform.Architecture

		if referencedPackage := pm.GetPackages().Packages[referencedPackageName]; referencedPackage == nil {
			formatter.PrintErrorMessage("The board requires platform '" + referencedPackageName + ":" + architecture + "' that is not installed.")
			os.Exit(cli.ErrGeneric)
		} else if referencedPlatform := referencedPackage.Platforms[architecture]; referencedPlatform == nil {
			formatter.PrintErrorMessage("The board requires platform '" + referencedPackageName + ":" + architecture + "' that is not installed.")
			os.Exit(cli.ErrGeneric)
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
			uploadProperties.Merge(requiredTool.RuntimeProperties())
		}
	}

	// Set properties for verbose upload
	if flags.verbose {
		if v, ok := uploadProperties.GetOk("upload.params.verbose"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
	} else {
		if v, ok := uploadProperties.GetOk("upload.params.quiet"); ok {
			uploadProperties.Set("upload.verbose", v)
		}
	}

	// Set properties for verify
	if flags.verify {
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
	if flags.importFile == "" {
		importPath = sketch.FullPath
		importFile = sketch.Name + "." + fqbnSuffix
	} else {
		importPath = paths.New(flags.importFile).Parent()
		importFile = paths.New(flags.importFile).Base()
	}

	outputTmpFile, ok := uploadProperties.GetOk("recipe.output.tmp_file")
	outputTmpFile = uploadProperties.ExpandPropsInString(outputTmpFile)
	if !ok {
		formatter.PrintErrorMessage("The platform does not define the required property 'recipe.output.tmp_file'.")
		os.Exit(cli.ErrGeneric)
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
			formatter.PrintErrorMessage("Compiled sketch not found: " + uploadFile.String() + ". Please compile first.")
		} else {
			formatter.PrintError(err, "Could not open compiled sketch.")
		}
		os.Exit(cli.ErrGeneric)
	}

	// Perform reset via 1200bps touch if requested
	if uploadProperties.GetBoolean("upload.use_1200bps_touch") {
		ports, err := serial.GetPortsList()
		if err != nil {
			formatter.PrintError(err, "Can't get serial port list")
			os.Exit(cli.ErrGeneric)
		}
		for _, p := range ports {
			if p == port {
				if err := touchSerialPortAt1200bps(p); err != nil {
					formatter.PrintError(err, "Can't perform reset via 1200bps-touch on serial port")
					os.Exit(cli.ErrGeneric)
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
			formatter.PrintError(err, "Could not detect serial ports")
			os.Exit(cli.ErrGeneric)
		} else if p == "" {
			formatter.Print("No new serial port detected.")
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
		formatter.PrintError(err, "Invalid recipe in platform.")
		os.Exit(cli.ErrCoreConfig)
	}

	// Run Tool
	cmd, err := executils.Command(cmdArgs)
	if err != nil {
		formatter.PrintError(err, "Could not execute upload tool.")
		os.Exit(cli.ErrGeneric)
	}

	executils.AttachStdoutListener(cmd, executils.PrintToStdout)
	executils.AttachStderrListener(cmd, executils.PrintToStderr)

	if err := cmd.Start(); err != nil {
		formatter.PrintError(err, "Could not execute upload tool.")
		os.Exit(cli.ErrGeneric)
	}
	if err := cmd.Wait(); err != nil {
		formatter.PrintError(err, "Error during upload.")
		os.Exit(cli.ErrGeneric)
	}
}

func touchSerialPortAt1200bps(port string) error {
	logrus.Infof("Touching port %s at 1200bps", port)

	// Open port
	p, err := serial.Open(port, &serial.Mode{BaudRate: 1200})
	if err != nil {
		return fmt.Errorf("open port: %s", err)
	}
	defer p.Close()

	if err = p.SetDTR(false); err != nil {
		return fmt.Errorf("can't set DTR")
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
