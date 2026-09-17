// This file is part of arduino-cli.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package upload

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/feedback/result"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

func NewUploadFirmwareCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	var portArgs arguments.Port
	var verbose bool
	var verify bool
	var dryRun bool
	var uploadFields = argumentsUploadFields{}
	uploadCommand := &cobra.Command{
		Use:   "firmware",
		Short: i18n.Tr("Upload a firmware file."),
		Long:  i18n.Tr("Upload a firmware file."),
		Example: "" +
			"  " + os.Args[0] + " firmware /home/user/Blink.fw -p /dev/ttyACM0\n" +
			"  " + os.Args[0] + " firmware Blink.fw -p 192.168.10.1 --upload-field password=abc",
		Args: cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			runUploadFirmwareCommand(cmd.Context(), srv, args, portArgs, verbose, verify, dryRun, uploadFields)
		},
	}
	portArgs.AddToCommand(uploadCommand, srv)
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, i18n.Tr("Verify uploaded binary after the upload."))
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Optional, turns on verbose mode."))
	uploadCommand.Flags().BoolVar(&dryRun, "dry-run", false, i18n.Tr("Do not perform the actual upload, just log out actions"))
	uploadFields.AddToCommand(uploadCommand)
	return uploadCommand
}

func runUploadFirmwareCommand(
	ctx context.Context,
	srv rpc.ArduinoCoreServiceServer,
	args []string,
	portArgs arguments.Port,
	verbose bool,
	verify bool,
	dryRun bool,
	uploadFields argumentsUploadFields,
) {
	inst := instance.CreateAndInit(ctx, srv)
	fwFilePath := args[0]

	resp, err := srv.ReadFirmwareFileDetails(ctx, &rpc.ReadFirmwareFileDetailsRequest{
		FirmwareFile: fwFilePath,
	})
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}
	fwDetails := resp.GetFirmwareFileDetails()

	port, err := portArgs.GetPort(ctx, inst, srv, "", fwDetails.GetProtocol(), nil)
	if err != nil {
		feedback.FatalError(err, feedback.ErrGeneric)
	}

	userFields := uploadFields.ResolveUserFields(fwDetails.GetUserFields())

	uploadReq := &rpc.UploadFirmwareFileRequest{
		Instance:     inst,
		FirmwareFile: fwFilePath,
		Port:         port,
		Verbose:      verbose,
		Verify:       verify,
		DryRun:       dryRun,
		UserFields:   userFields,
	}
	stdOut, stdErr, stdIOResult := feedback.OutputStreams()
	serv, streamResp := commands.UploadFirmwareFileToServerStreams(ctx, stdOut, stdErr)
	if err := srv.UploadFirmwareFile(uploadReq, serv); err != nil {
		errcode := feedback.ErrGeneric
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
