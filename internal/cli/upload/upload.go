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

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
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
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	uploadFields := map[string]string{}
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: i18n.Tr("Upload Arduino sketches."),
		Long:  i18n.Tr("Upload Arduino sketches. This does NOT compile the sketch prior to upload."),
		Example: "" +
			"  " + os.Args[0] + " upload /home/user/Arduino/MySketch -p /dev/ttyACM0 -b arduino:avr:uno\n" +
			"  " + os.Args[0] + " upload -p 192.168.10.1 -b arduino:avr:uno --upload-field password=abc",
		Args: cobra.MaximumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			arguments.CheckFlagsConflicts(cmd, "input-file", "input-dir")
		},
		Run: func(cmd *cobra.Command, args []string) {
			runUploadCommand(cmd.Context(), srv, args, uploadFields)
		},
	}

	fqbnArg.AddToCommand(uploadCommand, srv)
	portArgs.AddToCommand(uploadCommand, srv)
	profileArg.AddToCommand(uploadCommand, srv)
	uploadCommand.Flags().StringVarP(&importDir, "input-dir", "", "", i18n.Tr("Directory containing binaries to upload."))
	uploadCommand.Flags().StringVarP(&importFile, "input-file", "i", "", i18n.Tr("Binary file to upload."))
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, i18n.Tr("Verify uploaded binary after the upload."))
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Optional, turns on verbose mode."))
	programmer.AddToCommand(uploadCommand, srv)
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, i18n.Tr("Do not perform the actual upload, just log out actions"))
	uploadCommand.Flags().MarkHidden("dry-run")
	arguments.AddKeyValuePFlag(uploadCommand, &uploadFields, "upload-field", "F", nil, i18n.Tr("Set a value for a field required to upload."))
	return uploadCommand
}

func runUploadCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer, args []string, uploadFieldsArgs map[string]string) {
	logrus.Info("Executing `arduino-cli upload`")

	path := ""
	if len(args) > 0 {
		path = args[0]
	}
	sketchPath := arguments.InitSketchPath(path)
	resp, err := srv.LoadSketch(ctx, &rpc.LoadSketchRequest{SketchPath: sketchPath.String()})
	sketch := resp.GetSketch()
	if importDir == "" && importFile == "" {
		if err != nil {
			feedback.Fatal(i18n.Tr("Error during Upload: %v", err), feedback.ErrGeneric)
		}
		feedback.WarnAboutDeprecatedFiles(sketch)
	}

	var inst *rpc.Instance
	var profile *rpc.SketchProfile

	if profileArg.Get() == "" {
		inst, profile = instance.CreateAndInitWithProfile(ctx, srv, sketch.GetDefaultProfile().GetName(), sketchPath)
	} else {
		inst, profile = instance.CreateAndInitWithProfile(ctx, srv, profileArg.Get(), sketchPath)
	}

	if fqbnArg.String() == "" {
		fqbnArg.Set(profile.GetFqbn())
	}

	defaultFQBN := sketch.GetDefaultFqbn()
	defaultAddress := sketch.GetDefaultPort()
	defaultProtocol := sketch.GetDefaultProtocol()
	fqbn, port := arguments.CalculateFQBNAndPort(ctx, &portArgs, &fqbnArg, inst, srv, defaultFQBN, defaultAddress, defaultProtocol)

	userFieldRes, err := srv.SupportedUserFields(ctx, &rpc.SupportedUserFieldsRequest{
		Instance: inst,
		Fqbn:     fqbn,
		Protocol: port.GetProtocol(),
	})
	if err != nil {
		msg := i18n.Tr("Error during Upload: %v", err)

		// Check the error type to give the user better feedback on how
		// to resolve it
		var platformErr *cmderrors.PlatformNotFoundError
		if errors.As(err, &platformErr) {
			split := strings.Split(platformErr.Platform, ":")
			if len(split) < 2 {
				panic(i18n.Tr("Platform ID is not correct"))
			}

			msg += "\n"
			if platform, err := srv.PlatformSearch(ctx, &rpc.PlatformSearchRequest{
				Instance:   inst,
				SearchArgs: platformErr.Platform,
			}); err != nil {
				msg += err.Error()
			} else if len(platform.GetSearchOutput()) > 0 {
				msg += i18n.Tr("Try running %s", fmt.Sprintf("`%s core install %s`", version.VersionInfo.Application, platformErr.Platform))
			} else {
				msg += i18n.Tr("Platform %s is not found in any known index\nMaybe you need to add a 3rd party URL?", platformErr.Platform)
			}
		}
		feedback.Fatal(msg, feedback.ErrGeneric)
	}

	fields := map[string]string{}
	if len(userFieldRes.GetUserFields()) > 0 {
		if len(uploadFieldsArgs) > 0 {
			// If the user has specified some fields via cmd-line, we don't ask for them
			for _, field := range userFieldRes.GetUserFields() {
				if value, ok := uploadFieldsArgs[field.GetName()]; ok {
					fields[field.GetName()] = value
				} else {
					feedback.Fatal(i18n.Tr("Missing required upload field: %s", field.GetName()), feedback.ErrBadArgument)
				}
			}
		} else {
			// Otherwise prompt the user for them
			feedback.Print(i18n.Tr("Uploading to specified board using %s protocol requires the following info:", port.GetProtocol()))
			if f, err := arguments.AskForUserFields(userFieldRes.GetUserFields()); err != nil {
				msg := fmt.Sprintf("%s: %s", i18n.Tr("Error getting user input"), err)
				feedback.Fatal(msg, feedback.ErrGeneric)
			} else {
				fields = f
			}
		}
	}

	if sketchPath != nil {
		path = sketchPath.String()
	}

	prog := profile.GetProgrammer()
	if prog == "" || programmer.GetProgrammer() != "" {
		prog = programmer.String(ctx, inst, srv, fqbn)
	}
	if prog == "" {
		prog = sketch.GetDefaultProgrammer()
	}

	stdOut, stdErr, stdIOResult := feedback.OutputStreams()
	req := &rpc.UploadRequest{
		Instance:   inst,
		Fqbn:       fqbn,
		SketchPath: path,
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
		ImportDir:  importDir,
		Programmer: prog,
		DryRun:     dryRun,
		UserFields: fields,
	}
	stream, streamResp := commands.UploadToServerStreams(ctx, stdOut, stdErr)
	if err := srv.Upload(req, stream); err != nil {
		errcode := feedback.ErrGeneric
		if errors.Is(err, &cmderrors.ProgrammerRequiredForUploadError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		if errors.Is(err, &cmderrors.MissingProgrammerError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		feedback.FatalError(err, errcode)
	} else {
		io := stdIOResult()
		feedback.PrintResult(&uploadResult{
			Stdout:            io.Stdout,
			Stderr:            io.Stderr,
			UpdatedUploadPort: result.NewPort(streamResp().GetUpdatedUploadPort()),
		})
	}
}

type uploadResult struct {
	Stdout            string       `json:"stdout"`
	Stderr            string       `json:"stderr"`
	UpdatedUploadPort *result.Port `json:"updated_upload_port,omitempty"`
}

func (r *uploadResult) Data() interface{} {
	return r
}

func (r *uploadResult) String() string {
	if r.UpdatedUploadPort == nil {
		return ""
	}
	return i18n.Tr("New upload port: %[1]s (%[2]s)", r.UpdatedUploadPort.Address, r.UpdatedUploadPort.Protocol)
}
