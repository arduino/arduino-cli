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

package outdated

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/i18n"
	"github.com/arduino/arduino-cli/internal/cli/core"
	"github.com/arduino/arduino-cli/internal/cli/feedback"
	"github.com/arduino/arduino-cli/internal/cli/instance"
	"github.com/arduino/arduino-cli/internal/cli/lib"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tr = i18n.Tr

// NewCommand creates a new `outdated` command
func NewCommand() *cobra.Command {
	outdatedCommand := &cobra.Command{
		Use:   "outdated",
		Short: tr("Lists cores and libraries that can be upgraded"),
		Long: tr(`This commands shows a list of installed cores and/or libraries
that can be upgraded. If nothing needs to be updated the output is empty.`),
		Example: "  " + os.Args[0] + " outdated\n",
		Args:    cobra.NoArgs,
		Run:     runOutdatedCommand,
	}
	return outdatedCommand
}

func runOutdatedCommand(cmd *cobra.Command, args []string) {
	inst := instance.CreateAndInit()
	logrus.Info("Executing `arduino-cli outdated`")
	Outdated(inst)
}

// Outdated prints a list of outdated platforms and libraries
func Outdated(inst *rpc.Instance) {
	feedback.PrintResult(
		outdatedResult{core.GetList(inst, false, true), lib.GetList(inst, []string{}, false, true)},
	)
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type outdatedResult struct {
	Platforms     []*rpc.PlatformSummary  `json:"platforms,omitempty"`
	InstalledLibs []*rpc.InstalledLibrary `json:"libraries,omitempty"`
}

func (ir outdatedResult) Data() interface{} {
	return &ir
}

func (ir outdatedResult) String() string {
	if len(ir.Platforms) == 0 && len(ir.InstalledLibs) == 0 {
		return tr("No outdated platforms or libraries found.")
	}

	// A table useful both for platforms and libraries, where some of the fields will be blank.
	t := table.New()
	t.SetHeader(
		tr("ID"),
		tr("Name"),
		tr("Installed"),
		tr("Latest"),
		tr("Location"),
		tr("Description"),
	)
	t.SetColumnWidthMode(2, table.Average)
	t.SetColumnWidthMode(3, table.Average)
	t.SetColumnWidthMode(5, table.Average)

	// Based on internal/cli/core/list.go
	for _, p := range ir.Platforms {
		name := p.GetLatestRelease().GetName()
		if p.GetMetadata().Deprecated {
			name = fmt.Sprintf("[%s] %s", tr("DEPRECATED"), name)
		}
		t.AddRow(p.GetMetadata().Id, name, p.InstalledVersion, p.LatestVersion, "", "")
	}

	// Based on internal/cli/lib/list.go
	sort.Slice(ir.InstalledLibs, func(i, j int) bool {
		return strings.ToLower(
			ir.InstalledLibs[i].Library.Name,
		) < strings.ToLower(
			ir.InstalledLibs[j].Library.Name,
		) ||
			strings.ToLower(
				ir.InstalledLibs[i].Library.ContainerPlatform,
			) < strings.ToLower(
				ir.InstalledLibs[j].Library.ContainerPlatform,
			)
	})
	lastName := ""
	for _, libMeta := range ir.InstalledLibs {
		lib := libMeta.GetLibrary()
		name := lib.Name
		if name == lastName {
			name = ` "`
		} else {
			lastName = name
		}

		location := lib.GetLocation().String()
		if lib.ContainerPlatform != "" {
			location = lib.GetContainerPlatform()
		}

		available := ""
		sentence := ""
		if libMeta.GetRelease() != nil {
			available = libMeta.GetRelease().GetVersion()
			sentence = lib.Sentence
		}

		if available == "" {
			available = "-"
		}
		if sentence == "" {
			sentence = "-"
		} else if len(sentence) > 40 {
			sentence = sentence[:37] + "..."
		}
		t.AddRow("", name, lib.Version, available, location, sentence)
	}

	return t.Render()
}
