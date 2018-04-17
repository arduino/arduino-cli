/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package output

import (
	"fmt"
	"sort"

	"github.com/bcmi-labs/arduino-cli/arduino/cores"
	"github.com/gosuri/uitable"
)

// InstalledPlatformReleases represents an output set of tools or cores to format output from `core list` command.
type InstalledPlatformReleases []*cores.PlatformRelease

func (is InstalledPlatformReleases) Len() int      { return len(is) }
func (is InstalledPlatformReleases) Swap(i, j int) { is[i], is[j] = is[j], is[i] }
func (is InstalledPlatformReleases) Less(i, j int) bool {
	return is[i].Platform.String() < is[j].Platform.String()
}

func (is InstalledPlatformReleases) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	table.AddRow("FQBN", "Version", "Latest", "Name")
	sort.Sort(is)
	for _, item := range is {
		table.AddRow(item.Platform.String(), item.Version, item.Platform.GetRelease("latest").Version, item.Platform.Name)
	}
	return fmt.Sprintln(table)
}
