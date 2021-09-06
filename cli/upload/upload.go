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
	"os"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

var (
	fqbn       string
	port       arguments.Port
	verbose    bool
	verify     bool
	importDir  string
	importFile string
	programmer string
	dryRun     bool
	tr         = i18n.Tr
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   tr("Upload Arduino sketches."),
		Long:    tr("Upload Arduino sketches. This does NOT compile the sketch prior to upload."),
		Example: "  " + os.Args[0] + " upload /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		PreRun:  checkFlagsConflicts,
		Run:     run,
	}

	uploadCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", tr("Fully Qualified Board Name, e.g.: arduino:avr:uno"))
	port.AddToCommand(uploadCommand)
	uploadCommand.Flags().StringVarP(&importDir, "input-dir", "", "", tr("Directory containing binaries to upload."))
	uploadCommand.Flags().StringVarP(&importFile, "input-file", "i", "", tr("Binary file to upload."))
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, tr("Verify uploaded binary after the upload."))
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, tr("Optional, turns on verbose mode."))
	uploadCommand.Flags().StringVarP(&programmer, "programmer", "P", "", tr("Optional, use the specified programmer to upload."))
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, tr("Do not perform the actual upload, just log out actions"))
	uploadCommand.Flags().MarkHidden("dry-run")
	return uploadCommand
}

func checkFlagsConflicts(command *cobra.Command, args []string) {
	if importFile != "" && importDir != "" {
		feedback.Errorf(tr("error: %s and %s flags cannot be used together", "--input-file", "--input-dir"))
		os.Exit(errorcodes.ErrBadArgument)
	}
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	sketchPath := arguments.InitSketchPath(path)

	// .pde files are still supported but deprecated, this warning urges the user to rename them
	if files := sketch.CheckForPdeFiles(sketchPath); len(files) > 0 && importDir == "" && importFile == "" {
		feedback.Error(tr("Sketches with .pde extension are deprecated, please rename the following files to .ino:"))
		for _, f := range files {
			feedback.Error(f)
		}
	}

	sk, err := sketch.New(sketchPath)
	if err != nil && importDir == "" && importFile == "" {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	discoveryPort, err := port.GetPort(instance, sk)
	if err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if fqbn == "" && sk != nil && sk.Metadata != nil {
		// If the user didn't specify an FQBN and a sketch.json file is present
		// read it from there.
		fqbn = sk.Metadata.CPU.Fqbn
	}

	userFieldRes, err := upload.SupportedUserFields(context.Background(), &rpc.SupportedUserFieldsRequest{
		Instance: instance,
		Fqbn:     fqbn,
		Protocol: discoveryPort.Protocol,
	})
	if err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	fields := map[string]string{}
	if len(userFieldRes.UserFields) > 0 {
		feedback.Print(tr("Uploading to specified board using %s protocol requires the following info:", discoveryPort.Protocol))
		fields = arguments.AskForUserFields(userFieldRes.UserFields)
	}

	if sketchPath != nil {
		path = sketchPath.String()
	}

	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: path,
		Port:       discoveryPort.ToRPC(),
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
		ImportDir:  importDir,
		Programmer: programmer,
		DryRun:     dryRun,
		UserFields: fields,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
