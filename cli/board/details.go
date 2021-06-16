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
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr
var showFullDetails bool
var fqbn string
var listProgrammers bool

func initDetailsCommand() *cobra.Command {
	var detailsCommand = &cobra.Command{
		Use:     "details -b <FQBN>",
		Short:   tr("Print details about a board."),
		Long:    tr("Show information about a board, in particular if the board has options to be specified in the FQBN."),
		Example: "  " + os.Args[0] + " board details -b arduino:avr:nano",
		Args:    cobra.MaximumNArgs(1),
		Run:     runDetailsCommand,
	}

	detailsCommand.Flags().BoolVarP(&showFullDetails, "full", "f", false, tr("Show full board details"))
	detailsCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	detailsCommand.Flags().BoolVarP(&listProgrammers, "list-programmers", "", false, tr("Show list of available programmers"))
	// detailsCommand.MarkFlagRequired("fqbn") // enable once `board details <fqbn>` is removed

	return detailsCommand
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()

	// remove once `board details <fqbn>` is removed
	if fqbn == "" && len(args) > 0 {
		fqbn = args[0]
	}

	res, err := board.Details(context.Background(), &rpc.BoardDetailsRequest{
		Instance: inst,
		Fqbn:     fqbn,
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
	details *rpc.BoardDetailsResponse
}

func (dr detailsResult) Data() interface{} {
	return dr.details
}

func (dr detailsResult) String() string {
	details := dr.details

	if listProgrammers {
		t := table.New()
		t.AddRow(tr("Id"), tr("Programmer name"))
		for _, programmer := range details.Programmers {
			t.AddRow(programmer.GetId(), programmer.GetName())
		}
		return t.Render()
	}

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
	addIfNotEmpty := func(label, content string) {
		if content != "" {
			t.AddRow(label, content)
		}
	}

	t.SetColumnWidthMode(1, table.Average)
	t.AddRow(tr("Board name:"), details.Name)
	t.AddRow("FQBN:", details.Fqbn)
	addIfNotEmpty(tr("Board version:"), details.Version)
	if details.GetDebuggingSupported() {
		t.AddRow(tr("Debugging supported:"), table.NewCell("✔", color.New(color.FgGreen)))
	}

	if details.Official {
		t.AddRow() // get some space from above
		t.AddRow(tr("Official Arduino board:"),
			table.NewCell("✔", color.New(color.FgGreen)))
	}

	for i, idp := range details.IdentificationPrefs {
		if i == 0 {
			t.AddRow() // get some space from above
			t.AddRow(tr("Identification properties:"), "VID:"+idp.UsbId.Vid+" PID:"+idp.UsbId.Pid)
			continue
		}
		t.AddRow("", "VID:"+idp.UsbId.Vid+" PID:"+idp.UsbId.Pid)
	}

	t.AddRow() // get some space from above
	addIfNotEmpty(tr("Package name:"), details.Package.Name)
	addIfNotEmpty(tr("Package maintainer:"), details.Package.Maintainer)
	addIfNotEmpty(tr("Package URL:"), details.Package.Url)
	addIfNotEmpty(tr("Package website:"), details.Package.WebsiteUrl)
	addIfNotEmpty(tr("Package online help:"), details.Package.Help.Online)

	t.AddRow() // get some space from above
	addIfNotEmpty(tr("Platform name:"), details.Platform.Name)
	addIfNotEmpty(tr("Platform category:"), details.Platform.Category)
	addIfNotEmpty(tr("Platform architecture:"), details.Platform.Architecture)
	addIfNotEmpty(tr("Platform URL:"), details.Platform.Url)
	addIfNotEmpty(tr("Platform file name:"), details.Platform.ArchiveFilename)
	if details.Platform.Size != 0 {
		addIfNotEmpty(tr("Platform size (bytes):"), fmt.Sprint(details.Platform.Size))
	}
	addIfNotEmpty(tr("Platform checksum:"), details.Platform.Checksum)

	t.AddRow() // get some space from above
	for _, tool := range details.ToolsDependencies {
		t.AddRow(tr("Required tool:"), tool.Packager+":"+tool.Name, "", tool.Version)
		if showFullDetails {
			for _, sys := range tool.Systems {
				t.AddRow("", tr("OS:"), "", sys.Host)
				t.AddRow("", tr("File:"), "", sys.ArchiveFilename)
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

	t.AddRow(tr("Programmers:"), tr("Id"), tr("Name"))
	for _, programmer := range details.Programmers {
		t.AddRow("", programmer.GetId(), programmer.GetName())
	}

	return t.Render()
}
