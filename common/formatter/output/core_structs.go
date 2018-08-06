/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package output

import (
	"fmt"
	"sort"

	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/gosuri/uitable"
)

// InstalledPlatformReleases represents an output set of installed platforms.
type InstalledPlatformReleases []*cores.PlatformRelease

func (is InstalledPlatformReleases) Len() int      { return len(is) }
func (is InstalledPlatformReleases) Swap(i, j int) { is[i], is[j] = is[j], is[i] }
func (is InstalledPlatformReleases) Less(i, j int) bool {
	return is[i].Platform.String() < is[j].Platform.String()
}

// PlatformReleases represents an output set of tools of platforms.
type PlatformReleases []*cores.PlatformRelease

func (is PlatformReleases) Len() int      { return len(is) }
func (is PlatformReleases) Swap(i, j int) { is[i], is[j] = is[j], is[i] }
func (is PlatformReleases) Less(i, j int) bool {
	return is[i].Platform.String() < is[j].Platform.String()
}

func (is InstalledPlatformReleases) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	table.AddRow("ID", "Installed", "Latest", "Name")
	sort.Sort(is)
	for _, item := range is {
		table.AddRow(item.Platform.String(), item.Version, item.Platform.GetLatestRelease().Version, item.Platform.Name)
	}
	return fmt.Sprintln(table)
}

func (is PlatformReleases) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	table.AddRow("ID", "Version", "Installed", "Name")
	sort.Sort(is)
	for _, item := range is {
		installed := "No"
		if item.InstallDir != nil {
			installed = "Yes"
		}
		table.AddRow(item.Platform.String(), item.Version, installed, item.Platform.Name)
	}
	return fmt.Sprintln(table)
}
