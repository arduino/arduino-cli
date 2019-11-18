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

package compile

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/compile"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn               string   // Fully Qualified Board Name, e.g.: arduino:avr:uno.
	showProperties     bool     // Show all build preferences used instead of compiling.
	preprocess         bool     // Print preprocessed code to stdout.
	buildCachePath     string   // Builds of 'core.a' are saved into this path to be cached and reused.
	buildPath          string   // Path where to save compiled files.
	buildProperties    []string // List of custom build properties separated by commas. Or can be used multiple times for multiple properties.
	warnings           string   // Used to tell gcc which warning level to use.
	verbose            bool     // Turns on verbose mode.
	quiet              bool     // Suppresses almost every output.
	vidPid             string   // VID/PID specific build properties.
	uploadAfterCompile bool     // Upload the binary after the compilation.
	port               string   // Upload port, e.g.: COM10 or /dev/ttyACM0.
	verify             bool     // Upload, verify uploaded binary after the upload.
	exportFile         string   // The compiled binary is written to this file
)

// NewCommand created a new `compile` command
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:     "compile",
		Short:   "Compiles Arduino sketches.",
		Long:    "Compiles Arduino sketches.",
		Example: "  " + os.Args[0] + " compile -b arduino:avr:uno /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	command.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	command.Flags().BoolVar(&showProperties, "show-properties", false, "Show all build properties used instead of compiling.")
	command.Flags().BoolVar(&preprocess, "preprocess", false, "Print preprocessed code to stdout instead of compiling.")
	command.Flags().StringVar(&buildCachePath, "build-cache-path", "", "Builds of 'core.a' are saved into this path to be cached and reused.")
	command.Flags().StringVarP(&exportFile, "output", "o", "", "Filename of the compile output.")
	command.Flags().StringVar(&buildPath, "build-path", "",
		"Path where to save compiled files. If omitted, a directory will be created in the default temporary path of your OS.")
	command.Flags().StringSliceVar(&buildProperties, "build-properties", []string{},
		"List of custom build properties separated by commas. Or can be used multiple times for multiple properties.")
	command.Flags().StringVar(&warnings, "warnings", "none",
		`Optional, can be "none", "default", "more" and "all". Defaults to "none". Used to tell gcc which warning level to use (-W flag).`)
	command.Flags().BoolVarP(&verbose, "verbose", "v", false, "Optional, turns on verbose mode.")
	command.Flags().BoolVar(&quiet, "quiet", false, "Optional, supresses almost every output.")
	command.Flags().BoolVarP(&uploadAfterCompile, "upload", "u", false, "Upload the binary after the compilation.")
	command.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	command.Flags().BoolVarP(&verify, "verify", "t", false, "Verify uploaded binary after the upload.")
	command.Flags().StringVar(&vidPid, "vid-pid", "", "When specified, VID/PID specific build properties are used, if boards supports them.")

	return command
}

func run(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstance()

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}

	sketchPath := initSketchPath(path)

	_, err := compile.Compile(context.Background(), &rpc.CompileReq{
		Instance:        instance,
		Fqbn:            fqbn,
		SketchPath:      sketchPath.String(),
		ShowProperties:  showProperties,
		Preprocess:      preprocess,
		BuildCachePath:  buildCachePath,
		BuildPath:       buildPath,
		BuildProperties: buildProperties,
		Warnings:        warnings,
		Verbose:         verbose,
		Quiet:           quiet,
		VidPid:          vidPid,
		ExportFile:      exportFile,
	}, os.Stdout, os.Stderr, globals.LogLevel == "debug")

	if err != nil {
		feedback.Errorf("Error during build: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if uploadAfterCompile {
		_, err := upload.Upload(context.Background(), &rpc.UploadReq{
			Instance:   instance,
			Fqbn:       fqbn,
			SketchPath: sketchPath.String(),
			Port:       port,
			Verbose:    verbose,
			Verify:     verify,
			ImportFile: exportFile,
		}, os.Stdout, os.Stderr)

		if err != nil {
			feedback.Errorf("Error during Upload: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}
}

// initSketchPath returns the current working directory
func initSketchPath(sketchPath *paths.Path) *paths.Path {
	if sketchPath != nil {
		return sketchPath
	}

	wd, err := paths.Getwd()
	if err != nil {
		feedback.Errorf("Couldn't get current working directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return wd
}
