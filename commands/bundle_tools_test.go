// This file is part of arduino-cli.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"fmt"
	"testing"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/resources"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v2"
)

func TestBundledToolsDownloadAvailability(t *testing.T) {
	tmp, err := paths.MkTempDir("", "")
	require.NoError(t, err)
	defer tmp.RemoveAll()

	downloadAndTestChecksum := func(r *resources.DownloadResource) {
		fmt.Println("Testing:", r.URL)
		d, err := r.Download(tmp, &downloader.Config{})
		require.NoError(t, err)
		err = d.Run()
		require.NoError(t, err)

		checksum, err := r.TestLocalArchiveIntegrity(tmp)
		require.True(t, checksum)
		require.NoError(t, err)
	}

	toTest := [][]*cores.Flavor{
		serialMonitorFlavors,
		ctagsFlavors,
		serialDiscoveryFlavors,
		mdnsDiscoveryFlavors,
	}
	for _, flavors := range toTest {
		for _, resource := range flavors {
			downloadAndTestChecksum(resource.Resource)
			resource.Resource.ArchivePath(tmp)
		}
	}
}
