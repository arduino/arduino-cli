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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/cli/arguments"
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbnArg    arguments.Fqbn
	portArgs   arguments.Port
	verbose    bool
	verify     bool
	importDir  string
	importFile string
	programmer arguments.Programmer
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
		PreRun: func(cmd *cobra.Command, args []string) {
			arguments.CheckFlagsConflicts(cmd, "input-file", "input-dir")
		},
		Run: runUploadCommand,
	}

	fqbnArg.AddToCommand(uploadCommand)
	portArgs.AddToCommand(uploadCommand)
	uploadCommand.Flags().StringVarP(&importDir, "input-dir", "", "", tr("Directory containing binaries to upload."))
	uploadCommand.Flags().StringVarP(&importFile, "input-file", "i", "", tr("Binary file to upload."))
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, tr("Verify uploaded binary after the upload."))
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, tr("Optional, turns on verbose mode."))
	programmer.AddToCommand(uploadCommand)
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, tr("Do not perform the actual upload, just log out actions"))
	uploadCommand.Flags().MarkHidden("dry-run")
	return uploadCommand
}

func runUploadCommand(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli upload`")

	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	sketchPath := arguments.InitSketchPath(path)

	if importDir == "" && importFile == "" {
		arguments.WarnDeprecatedFiles(sketchPath)
	}

	sk, err := sketch.New(sketchPath)
	if err != nil && importDir == "" && importFile == "" {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	port, err := portArgs.GetPort(instance, sk)
	if err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	if fqbnArg.String() == "" && sk != nil && sk.Metadata != nil {
		// If the user didn't specify an FQBN and a sketch.json file is present
		// read it from there.
		fqbnArg.Set(sk.Metadata.CPU.Fqbn)
	}

	userFieldRes, err := upload.SupportedUserFields(context.Background(), &rpc.SupportedUserFieldsRequest{
		Instance: instance,
		Fqbn:     fqbnArg.String(),
		Address:  port.Address,
		Protocol: port.Protocol,
	})
	if err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *arduino.PlatformNotFoundError
		if errors.As(err, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(tr("Platform ID is not correct"))
			}

			pm := commands.GetPackageManager(instance.GetId())
			platform := pm.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
			})

			if platform != nil {
				feedback.Errorf(tr("Try running %s", fmt.Sprintf("`%s core install %s`", globals.VersionInfo.Application, platformErr.Platform)))
			} else {
				feedback.Errorf(tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform))
			}
		}
		os.Exit(errorcodes.ErrGeneric)
	}

	fields := map[string]string{}
	if len(userFieldRes.UserFields) > 0 {
		feedback.Print(tr("Uploading to specified board using %s protocol requires the following info:", port.Protocol))
		fields = arguments.AskForUserFields(userFieldRes.UserFields)
	}

	if sketchPath != nil {
		path = sketchPath.String()
	}

	if _, err := upload.Upload(context.Background(), &rpc.UploadRequest{
		Instance:   instance,
		Fqbn:       fqbnArg.String(),
		SketchPath: path,
		Port:       port.ToRPC(),
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
		ImportDir:  importDir,
		Programmer: programmer.String(),
		DryRun:     dryRun,
		UserFields: fields,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf(tr("Error during Upload: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}
}
