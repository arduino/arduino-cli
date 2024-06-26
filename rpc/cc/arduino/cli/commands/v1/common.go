// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (https://www.arduino.cc/)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"sort"

	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	semver "go.bug.st/relaxed-semver"
)

// DownloadProgressCB is a callback to get updates on download progress
type DownloadProgressCB func(curr *DownloadProgress)

// Start sends a "start" DownloadProgress message to the callback function
func (d DownloadProgressCB) Start(url, label string) {
	d(&DownloadProgress{
		Message: &DownloadProgress_Start{
			Start: &DownloadProgressStart{
				Url:   url,
				Label: label,
			},
		},
	})
}

// Update sends an "update" DownloadProgress message to the callback function
func (d DownloadProgressCB) Update(downloaded int64, totalSize int64) {
	d(&DownloadProgress{
		Message: &DownloadProgress_Update{
			Update: &DownloadProgressUpdate{
				Downloaded: downloaded,
				TotalSize:  totalSize,
			},
		},
	})
}

// End sends an "end" DownloadProgress message to the callback function
func (d DownloadProgressCB) End(success bool, message string) {
	d(&DownloadProgress{
		Message: &DownloadProgress_End{
			End: &DownloadProgressEnd{
				Success: success,
				Message: message,
			},
		},
	})
}

// TaskProgressCB is a callback to receive progress messages
type TaskProgressCB func(msg *TaskProgress)

// InstanceCommand is an interface that represents a gRPC command with
// a gRPC Instance.
type InstanceCommand interface {
	GetInstance() *Instance
}

// GetLatestRelease returns the latest release in this PlatformSummary,
// or nil if not available.
func (s *PlatformSummary) GetLatestRelease() *PlatformRelease {
	if s.GetLatestVersion() == "" {
		return nil
	}
	return s.GetReleases()[s.GetLatestVersion()]
}

// GetInstalledRelease returns the latest release in this PlatformSummary,
// or nil if not available.
func (s *PlatformSummary) GetInstalledRelease() *PlatformRelease {
	if s.GetInstalledVersion() == "" {
		return nil
	}
	return s.GetReleases()[s.GetInstalledVersion()]
}

// GetSortedReleases returns the releases in order of version.
func (s *PlatformSummary) GetSortedReleases() []*PlatformRelease {
	res := []*PlatformRelease{}
	for _, release := range s.GetReleases() {
		res = append(res, release)
	}
	sort.SliceStable(res, func(i, j int) bool {
		return semver.ParseRelaxed(res[i].GetVersion()).LessThan(semver.ParseRelaxed(res[j].GetVersion()))
	})
	return res
}

// DiscoveryPortToRPC converts a *discovery.Port into an *rpc.Port
func DiscoveryPortToRPC(p *discovery.Port) *Port {
	props := p.Properties
	if props == nil {
		props = properties.NewMap()
	}
	return &Port{
		Address:       p.Address,
		Label:         p.AddressLabel,
		Protocol:      p.Protocol,
		ProtocolLabel: p.ProtocolLabel,
		HardwareId:    p.HardwareID,
		Properties:    props.AsMap(),
	}
}

// DiscoveryPortFromRPCPort converts an *rpc.Port into a *discovery.Port
func DiscoveryPortFromRPCPort(o *Port) (p *discovery.Port) {
	if o == nil {
		return nil
	}
	res := &discovery.Port{
		Address:       o.GetAddress(),
		AddressLabel:  o.GetLabel(),
		Protocol:      o.GetProtocol(),
		ProtocolLabel: o.GetProtocolLabel(),
		HardwareID:    o.GetHardwareId(),
	}
	if o.GetProperties() != nil {
		res.Properties = properties.NewFromHashmap(o.GetProperties())
	}
	return res
}
