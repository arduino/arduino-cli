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
	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/arduino-cli/table"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
)

var detailsCommand = &cobra.Command{
	Use:     "details <FQBN>",
	Short:   "Print details about a board.",
	Long:    "Show information about a board, in particular if the board has options to be specified in the FQBN.",
	Example: "  " + os.Args[0] + " board details arduino:avr:nano",
	Args:    cobra.ExactArgs(1),
	Run:     runDetailsCommand,
}

func runDetailsCommand(cmd *cobra.Command, args []string) {
	inst, err := instance.CreateInstance()
	if err != nil {
		feedback.Errorf("Error getting board details: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	res, err := board.Details(context.Background(), &rpc.BoardDetailsReq{
		Instance: inst,
		Fqbn:     args[0],
	})

	if err != nil {
		feedback.Errorf("Error getting board details: %v", err)
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
	t.AddRow("Board name:", details.Name)
	t.AddRow("Board fqbn:", details.Fqbn)
	t.AddRow("Board propertiesId:", details.PropertiesId)
	t.AddRow("Board version:", details.Version)

	t.AddRow() // get some space from above
	if details.Official {
		t.AddRow("Official Arduino board:",
			table.NewCell("✔", color.New(color.FgGreen)))
	}

	for i, idp := range details.IdentificationPref {
		if i == 0 {
			t.AddRow() // get some space from above
			t.AddRow("Identification Preferences:", "VID:"+idp.UsbID.VID+" PID:"+idp.UsbID.PID)
			continue
		}
		t.AddRow("", "VID:"+idp.UsbID.VID+" PID:"+idp.UsbID.PID)
	}

	t.AddRow() // get some space from above
	t.AddRow("Package name:", details.Package.Name)
	t.AddRow("Package maintainer:", details.Package.Maintainer)
	t.AddRow("Package URL:", details.Package.Url)
	t.AddRow("Package websiteURL:", details.Package.WebsiteURL)
	t.AddRow("Package online help:", details.Package.Help.Online)

	t.AddRow() // get some space from above
	t.AddRow("Platform name:", details.Platform.Name)
	t.AddRow("Platform category:", details.Platform.Category)
	t.AddRow("Platform architecture:", details.Platform.Architecture)
	t.AddRow("Platform URL:", details.Platform.Url)
	t.AddRow("Platform file name:", details.Platform.ArchiveFileName)
	t.AddRow("Platform size (bytes):", fmt.Sprint(details.Platform.Size))
	t.AddRow("Platform checksum:", details.Platform.Checksum)

	t.AddRow() // get some space from above
	for _, tool := range details.ToolsDependencies {
		t.AddRow("Required tool: " + tool.Packager + ":" + tool.Name + "  ver: " + tool.Version)
		for _, sys := range tool.Systems {
			t.AddRow("OS: " + sys.Host)
			t.AddRow("File: " + sys.ArchiveFileName)
			t.AddRow("Size (bytes): " + fmt.Sprint(sys.Size))
			t.AddRow("Checksum: " + sys.Checksum)
			t.AddRow("URL: " + sys.Url)
			t.AddRow() // get some space from above
		}
		t.AddRow() // get some space from above
	}

	for _, option := range details.ConfigOptions {
		t.AddRow("Option:", option.OptionLabel, "", option.Option)
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
