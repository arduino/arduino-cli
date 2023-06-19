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
	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbnArg    arguments.Fqbn
	portArgs   arguments.Port
	profileArg arguments.Profile
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
	profileArg.AddToCommand(uploadCommand)
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
		feedback.Fatal(tr("Error during Upload: %v", err), feedback.ErrSketchError)
	}

	instance, profile := instance.CreateAndInitWithProfile(profileArg.Get(), sketchPath)
	if fqbnArg.String() == "" {
		fqbnArg.Set(profile.GetFqbn())
	}

	defaultFQBN := sk.GetDefaultFQBN()
	defaultAddress, defaultProtocol := sk.GetDefaultPortAddressAndProtocol()
	fqbn, port := arguments.CalculateFQBNAndPort(&portArgs, &fqbnArg, instance, defaultFQBN, defaultAddress, defaultProtocol)

	userFieldRes, err := upload.SupportedUserFields(context.Background(), &rpc.SupportedUserFieldsRequest{
		Instance: instance,
		Fqbn:     fqbn,
		Protocol: port.Protocol,
	})
	if err != nil {
		msg := tr("Error during Upload: %v", err)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *arduino.PlatformNotFoundError
		if errors.As(err, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(tr("Platform ID is not correct"))
			}

			// FIXME: Here we must not access package manager...
			pme, release := commands.GetPackageManagerExplorer(&rpc.UploadRequest{Instance: instance})
			platform := pme.FindPlatform(&packagemanager.PlatformReference{
				Package:              split[0],
				PlatformArchitecture: split[1],
			})
			release()

			msg += "\n"
			if platform != nil {
				msg += tr("Try running %s", fmt.Sprintf("`%s core install %s`", version.VersionInfo.Application, platformErr.Platform))
			} else {
				msg += tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform)
			}
		}
		feedback.Fatal(msg, feedback.ErrBoardUpload)
	}

	fields := map[string]string{}
	if len(userFieldRes.UserFields) > 0 {
		feedback.Print(tr("Uploading to specified board using %s protocol requires the following info:", port.Protocol))
		if f, err := arguments.AskForUserFields(userFieldRes.UserFields); err != nil {
			msg := fmt.Sprintf("%s: %s", tr("Error getting user input"), err)
			feedback.Fatal(msg, feedback.ErrGeneric)
		} else {
			fields = f
		}
	}

	if sketchPath != nil {
		path = sketchPath.String()
	}

	stdOut, stdErr, stdIOResult := feedback.OutputStreams()
	req := &rpc.UploadRequest{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: path,
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
		ImportDir:  importDir,
		Programmer: programmer.String(),
		DryRun:     dryRun,
		UserFields: fields,
	}
	if err := upload.Upload(context.Background(), req, stdOut, stdErr); err != nil {
		feedback.FatalError(err, feedback.ErrBoardUpload)
	}
	feedback.PrintResult(stdIOResult())
}
