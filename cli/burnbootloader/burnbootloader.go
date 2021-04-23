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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/spf13/cobra"
)

var (
	fqbn       string
	port       string
	verbose    bool
	verify     bool
	programmer string
)

// NewCommand created a new `burn-bootloader` command
func NewCommand() *cobra.Command {
	burnBootloaderCommand := &cobra.Command{
		Use:     "burn-bootloader",
		Short:   "Upload the bootloader.",
		Long:    "Upload the bootloader on the board using an external programmer.",
		Example: "  " + os.Args[0] + " burn-bootloader -b arduino:avr:uno -P atmel-ice",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	burnBootloaderCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	burnBootloaderCommand.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	burnBootloaderCommand.Flags().BoolVarP(&verify, "verify", "t", false, "Verify uploaded binary after the upload.")
	burnBootloaderCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Turns on verbose mode.")
	burnBootloaderCommand.Flags().StringVarP(&programmer, "programmer", "P", "", "Use the specified programmer to upload.")

	return burnBootloaderCommand
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateAndInit()

	if _, err := upload.BurnBootloader(context.Background(), &rpc.BurnBootloaderRequest{
		Instance:   instance,
		Fqbn:       fqbn,
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		Programmer: programmer,
	}, os.Stdout, os.Stderr); err != nil {
		feedback.Errorf("Error during Upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	os.Exit(0)
}
