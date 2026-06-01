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

package burnbootloader

import (
	"context"
	"errors"
	"os"

	"github.com/arduino/arduino-cli/commands"
	"github.com/arduino/arduino-cli/commands/cmderrors"
	"github.com/arduino/arduino-cli/internal/cli/arguments"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn             arguments.Fqbn
	port             arguments.Port
	verbose          bool
	verify           bool
	programmer       arguments.Programmer
	dryRun           bool
	uploadProperties []string
)

// NewCommand created a new `burn-bootloader` command
func NewCommand(srv rpc.ArduinoCoreServiceServer) *cobra.Command {
	burnBootloaderCommand := &cobra.Command{
		Use:     "burn-bootloader",
		Short:   i18n.Tr("Upload the bootloader."),
		Long:    i18n.Tr("Upload the bootloader on the board using an external programmer."),
		Example: "  " + os.Args[0] + " burn-bootloader -b arduino:avr:uno -P atmel_ice",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runBootloaderCommand(cmd.Context(), srv)
		},
	}

	fqbn.AddToCommand(burnBootloaderCommand, srv)
	port.AddToCommand(burnBootloaderCommand, srv)
	programmer.AddToCommand(burnBootloaderCommand, srv)
	burnBootloaderCommand.Flags().StringArrayVar(&uploadProperties, "upload-property", []string{},
		i18n.Tr("Override an upload property with a custom value. Can be used multiple times for multiple properties."))
	burnBootloaderCommand.Flags().BoolVarP(&verify, "verify", "t", false, i18n.Tr("Verify uploaded binary after the upload."))
	burnBootloaderCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, i18n.Tr("Turns on verbose mode."))
	burnBootloaderCommand.Flags().BoolVar(&dryRun, "dry-run", false, i18n.Tr("Do not perform the actual upload, just log out actions"))
	burnBootloaderCommand.Flags().MarkHidden("dry-run")

	return burnBootloaderCommand
}

func runBootloaderCommand(ctx context.Context, srv rpc.ArduinoCoreServiceServer) {
	instance := instance.CreateAndInit(ctx, srv)

	logrus.Info("Executing `arduino-cli burn-bootloader`")

	// We don't need a Sketch to upload a board's bootloader
	discoveryPort, err := port.GetPort(ctx, instance, srv, "", "", nil)
	if err != nil {
		feedback.Fatal(i18n.Tr("Error during Upload: %v", err), feedback.ErrGeneric)
	}

	stdOut, stdErr, res := feedback.OutputStreams()
	stream := commands.BurnBootloaderToServerStreams(ctx, stdOut, stdErr)
	if err := srv.BurnBootloader(&rpc.BurnBootloaderRequest{
		Instance:         instance,
		Fqbn:             fqbn.String(),
		Port:             discoveryPort,
		Verbose:          verbose,
		Verify:           verify,
		Programmer:       programmer.String(ctx, instance, srv, fqbn.String()),
		UploadProperties: uploadProperties,
		DryRun:           dryRun,
	}, stream); err != nil {
		errcode := feedback.ErrGeneric
		if errors.Is(err, &cmderrors.ProgrammerRequiredForUploadError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		if errors.Is(err, &cmderrors.MissingProgrammerError{}) {
			errcode = feedback.ErrMissingProgrammer
		}
		feedback.Fatal(i18n.Tr("Error during Upload: %v", err), errcode)
	}
	feedback.PrintResult(res())
}
