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

package board

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/i18n"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr
var showFullDetails bool

func initDetailsCommand() *cobra.Command {
	var detailsCommand = &cobra.Command{
		Use:     "details <FQBN>",
		Short:   tr("Print details about a board."),
		Long:    tr("Show information about a board, in particular if the board has options to be specified in the FQBN."),
		Example: "  " + os.Args[0] + " board details arduino:avr:nano",
		Args:    cobra.ExactArgs(1),
		Run:     runDetailsCommand,
	}

	detailsCommand.Flags().BoolVarP(&showFullDetails, "full", "f", false, tr("Show full board details"))

	return detailsCommand
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf(tr("Error getting board details: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	res, err := board.Details(context.Background(), &rpc.BoardDetailsReq{
		Instance: inst,
		Fqbn:     args[0],
	})

	if err != nil {
		feedback.Errorf(tr("Error getting board details: %v"), err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(detailsResult{details: res})
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type detailsResult struct {
	details *rpc.BoardDetailsResp
}

func (dr detailsResult) Data() interface{} {
	return dr.details
}

func (dr detailsResult) String() string {
	details := dr.details
	// Table is 4 columns wide:
	// |               |                             | |                       |
	// Board name:     Arduino Nano
	//
	// Required tools: arduino:avr-gcc                 5.4.0-atmel3.6.1-arduino2
	//                 arduino:avrdude                 6.3.0-arduino14
	//                 arduino:arduinoOTA              1.2.1
	//
	// Option:         Processor                       cpu
	//                 ATmega328P                    ✔ cpu=atmega328
	//                 ATmega328P (Old Bootloader)     cpu=atmega328old
	//                 ATmega168                       cpu=atmega168
	t := table.New()
	t.SetColumnWidthMode(1, table.Average)
	t.AddRow(tr("Board name:"), details.Name)
	t.AddRow("FQBN:", details.Fqbn)
	t.AddRow(tr("Identification properties:"), details.PropertiesId)
	t.AddRow(tr("Board version:"), details.Version)

	if details.Official {
		t.AddRow() // get some space from above
		t.AddRow(tr("Official Arduino board:"),
			table.NewCell("✔", color.New(color.FgGreen)))
	}

	for i, idp := range details.IdentificationPref {
		if i == 0 {
			t.AddRow() // get some space from above
			t.AddRow(tr("Identification properties:"), "VID:"+idp.UsbID.VID+" PID:"+idp.UsbID.PID)
			continue
		}
		t.AddRow("", "VID:"+idp.UsbID.VID+" PID:"+idp.UsbID.PID)
	}

	t.AddRow() // get some space from above
	t.AddRow(tr("Package name:"), details.Package.Name)
	t.AddRow(tr("Package maintainer:"), details.Package.Maintainer)
	t.AddRow(tr("Package URL:"), details.Package.Url)
	t.AddRow(tr("Package website:"), details.Package.WebsiteURL)
	t.AddRow(tr("Package online help:"), details.Package.Help.Online)

	t.AddRow() // get some space from above
	t.AddRow(tr("Platform name:"), details.Platform.Name)
	t.AddRow(tr("Platform category:"), details.Platform.Category)
	t.AddRow(tr("Platform architecture:"), details.Platform.Architecture)
	t.AddRow(tr("Platform URL:"), details.Platform.Url)
	t.AddRow(tr("Platform file name:"), details.Platform.ArchiveFileName)
	t.AddRow(tr("Platform size (bytes):"), fmt.Sprint(details.Platform.Size))
	t.AddRow(tr("Platform checksum:"), details.Platform.Checksum)

	t.AddRow() // get some space from above
	for _, tool := range details.ToolsDependencies {
		t.AddRow(tr("Required tools:"), tool.Packager+":"+tool.Name, "", tool.Version)
		if showFullDetails {
			for _, sys := range tool.Systems {
				t.AddRow("", tr("OS:"), "", sys.Host)
				t.AddRow("", tr("File:"), "", sys.ArchiveFileName)
				t.AddRow("", tr("Size (bytes):"), "", fmt.Sprint(sys.Size))
				t.AddRow("", tr("Checksum:"), "", sys.Checksum)
				t.AddRow("", "URL:", "", sys.Url)
				t.AddRow() // get some space from above
			}
		}
		t.AddRow() // get some space from above
	}

	for _, option := range details.ConfigOptions {
		t.AddRow(tr("Option:"), option.OptionLabel, "", option.Option)
		for _, value := range option.Values {
			green := color.New(color.FgGreen)
			if value.Selected {
				t.AddRow("",
					table.NewCell(value.ValueLabel, green),
					table.NewCell("✔", green),
					table.NewCell(option.Option+"="+value.Value, green))
			} else {
				t.AddRow("",
					value.ValueLabel,
					"",
					option.Option+"="+value.Value)
			}
		}
	}

	return t.Render()
}
