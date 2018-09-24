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

	"github.com/gosuri/uitable"
	semver "go.bug.st/relaxed-semver"
)

// InstalledPlatforms represents an output of a set of installed platforms.
type InstalledPlatforms struct {
	Platforms []*InstalledPlatform
}

// InstalledPlatform represents an output of an installed plaform.
type InstalledPlatform struct {
	ID        string
	Installed *semver.Version
	Latest    *semver.Version
	Name      string
}

func (is InstalledPlatforms) less(i, j int) bool {
	return is.Platforms[i].ID < is.Platforms[j].ID
}

func (is InstalledPlatforms) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	table.AddRow("ID", "Installed", "Latest", "Name")
	sort.Slice(is.Platforms, is.less)
	for _, item := range is.Platforms {
		table.AddRow(item.ID, item.Installed, item.Latest, item.Name)
	}
	return fmt.Sprintln(table)
}

// SearchedPlatforms represents an output of a set of searched platforms
type SearchedPlatforms struct {
	Platforms []*SearchedPlatform
}

func (is SearchedPlatforms) less(i, j int) bool {
	return is.Platforms[i].ID < is.Platforms[j].ID
}

// SearchedPlatform represents an output of a searched platform
type SearchedPlatform struct {
	ID      string
	Version *semver.Version
	Name    string
}

func (is SearchedPlatforms) String() string {
	table := uitable.New()
	table.MaxColWidth = 100
	table.Wrap = true

	sort.Slice(is.Platforms, is.less)
	table.AddRow("ID", "Version", "Name")
	for _, item := range is.Platforms {
		table.AddRow(item.ID, item.Version, item.Name)
	}
	return fmt.Sprintln(table)
}
