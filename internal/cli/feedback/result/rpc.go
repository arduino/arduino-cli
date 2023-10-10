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

package result

import (
	"github.com/arduino/arduino-cli/internal/orderedmap"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	semver "go.bug.st/relaxed-semver"
)

// NewPlatformResult creates a new result.Platform from rpc.PlatformSummary
func NewPlatformResult(in *rpc.PlatformSummary) *Platform {
	releases := orderedmap.New[string, *PlatformRelease]()
	for k, v := range in.Releases {
		releases.Set(k, NewPlatformReleaseResult(v))
	}
	releases.SortKeys(func(x, y string) int {
		return semver.ParseRelaxed(x).CompareTo(semver.ParseRelaxed(y))
	})

	return &Platform{
		Id:                in.Metadata.Id,
		Maintainer:        in.Metadata.Maintainer,
		Website:           in.Metadata.Website,
		Email:             in.Metadata.Email,
		ManuallyInstalled: in.Metadata.ManuallyInstalled,
		Deprecated:        in.Metadata.Deprecated,
		Indexed:           in.Metadata.Indexed,
		Releases:          releases,
		InstalledVersion:  in.InstalledVersion,
		LatestVersion:     in.LatestVersion,
	}
}

// Platform maps a rpc.Platform
type Platform struct {
	Id                string `json:"id,omitempty"`
	Maintainer        string `json:"maintainer,omitempty"`
	Website           string `json:"website,omitempty"`
	Email             string `json:"email,omitempty"`
	ManuallyInstalled bool   `json:"manually_installed,omitempty"`
	Deprecated        bool   `json:"deprecated,omitempty"`
	Indexed           bool   `json:"indexed,omitempty"`

	Releases *orderedmap.Map[string, *PlatformRelease] `json:"releases,omitempty"`

	InstalledVersion string `json:"installed_version,omitempty"`
	LatestVersion    string `json:"latest_version,omitempty"`
}

// GetLatestRelease returns the latest relase of this platform or nil if none available.
func (p *Platform) GetLatestRelease() *PlatformRelease {
	return p.Releases.Get(p.LatestVersion)
}

// GetInstalledRelease returns the installed relase of this platform or nil if none available.
func (p *Platform) GetInstalledRelease() *PlatformRelease {
	return p.Releases.Get(p.InstalledVersion)
}

// NewPlatformReleaseResult creates a new result.PlatformRelease from rpc.PlatformRelease
func NewPlatformReleaseResult(in *rpc.PlatformRelease) *PlatformRelease {
	var boards []*Board
	for _, board := range in.Boards {
		boards = append(boards, &Board{
			Name: board.Name,
			Fqbn: board.Fqbn,
		})
	}
	var help *HelpResource
	if in.Help != nil {
		help = &HelpResource{
			Online: in.Help.Online,
		}
	}
	res := &PlatformRelease{
		Name:            in.Name,
		Version:         in.Version,
		Type:            in.Type,
		Installed:       in.Installed,
		Boards:          boards,
		Help:            help,
		MissingMetadata: in.MissingMetadata,
		Deprecated:      in.Deprecated,
	}
	return res
}

// PlatformRelease maps a rpc.PlatformRelease
type PlatformRelease struct {
	Name            string        `json:"name,omitempty"`
	Version         string        `json:"version,omitempty"`
	Type            []string      `json:"type,omitempty"`
	Installed       bool          `json:"installed,omitempty"`
	Boards          []*Board      `json:"boards,omitempty"`
	Help            *HelpResource `json:"help,omitempty"`
	MissingMetadata bool          `json:"missing_metadata,omitempty"`
	Deprecated      bool          `json:"deprecated,omitempty"`
}

// Board maps a rpc.Board
type Board struct {
	Name string `json:"name,omitempty"`
	Fqbn string `json:"fqbn,omitempty"`
}

// HelpResource maps a rpc.HelpResource
type HelpResource struct {
	Online string `json:"online,omitempty"`
}
