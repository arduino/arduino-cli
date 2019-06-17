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
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/common/formatter"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
)

// InitCommand prepares the command.
func InitCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   "Upload Arduino sketches.",
		Long:    "Upload Arduino sketches.",
		Example: "  " + cli.VersionInfo.Application + " upload /home/user/Arduino/MySketch",
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
	instance := cli.CreateInstance()

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := cli.InitSketchPath(path)

	uploadRes, err := upload.Upload(context.Background(), &rpc.UploadReq{
		Instance:   instance,
		Fqbn:       flags.fqbn,
		SketchPath: sketchPath.String(),
		Port:       flags.port,
		Verbose:    flags.verbose,
		Verify:     flags.verify,
		ImportFile: flags.importFile,
	}, os.Stdout, os.Stderr)
	if err == nil {
		outputUploadResp(uploadRes)
	} else {
		formatter.PrintError(err, "Error during Upload")
		os.Exit(cli.ErrGeneric)
	}
}

func outputUploadResp(details *rpc.UploadResp) {

}
