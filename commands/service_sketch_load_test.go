// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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
	"context"
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func TestLoadSketchProfiles(t *testing.T) {
	srv := NewArduinoCoreServer()
	loadResp, err := srv.LoadSketch(context.Background(), &commands.LoadSketchRequest{
		SketchPath: "./testdata/sketch_with_profile",
	})
	require.NoError(t, err)
	require.Equal(t, loadResp.GetSketch().GetDefaultProfile().GetName(), "nanorp")
	require.Len(t, loadResp.GetSketch().GetProfiles(), 3)

	// nanorp:
	//   fqbn: arduino:avr:uno
	//   platforms:
	//     - platform: arduino:mbed_nano (4.0.2)
	//   libraries:
	//     - ArduinoIoTCloud (1.0.2)
	//     - Arduino_ConnectionHandler (0.6.4)
	//     - TinyDHT sensor library (1.1.0)
	nanorp := loadResp.GetSketch().GetProfiles()[0]
	require.Equal(t, nanorp.GetName(), "nanorp")
	require.Equal(t, nanorp.GetFqbn(), "arduino:avr:uno")
	require.Len(t, nanorp.GetPlatforms(), 1)
	require.Equal(t, nanorp.GetPlatforms()[0].GetId(), "arduino:mbed_nano")
	require.Equal(t, nanorp.GetPlatforms()[0].GetVersion(), "4.0.2")
	require.Len(t, nanorp.GetLibraries(), 3)
	require.Equal(t, nanorp.GetLibraries()[0].GetIndexLibrary().GetName(), "ArduinoIoTCloud")
	require.Equal(t, nanorp.GetLibraries()[0].GetIndexLibrary().GetVersion(), "1.0.2")
	require.Equal(t, nanorp.GetLibraries()[1].GetIndexLibrary().GetName(), "Arduino_ConnectionHandler")
	require.Equal(t, nanorp.GetLibraries()[1].GetIndexLibrary().GetVersion(), "0.6.4")
	require.Equal(t, nanorp.GetLibraries()[2].GetIndexLibrary().GetName(), "TinyDHT sensor library")
	require.Equal(t, nanorp.GetLibraries()[2].GetIndexLibrary().GetVersion(), "1.1.0")

	// profile2:
	//   fqbn: arduino:avr:uno
	//   platforms:
	//     - platform: arduino:mbed_nano (4.0.2)
	//     - platform: test:mbed_nano (1.2.3)
	//       platform_index_url: https://test.com/mbed_nano_index.json
	//   libraries:
	//     - ArduinoIoTCloud (1.0.2)
	//     - dir: libraries/Arduino_ConnectionHandler
	profile2 := loadResp.GetSketch().GetProfiles()[2]
	require.Equal(t, profile2.GetName(), "profile2")
	require.Equal(t, profile2.GetFqbn(), "arduino:avr:uno")
	require.Len(t, profile2.GetPlatforms(), 2)
	require.Equal(t, profile2.GetPlatforms()[0].GetId(), "arduino:mbed_nano")
	require.Equal(t, profile2.GetPlatforms()[0].GetVersion(), "4.0.2")
	require.Equal(t, profile2.GetPlatforms()[1].GetId(), "test:mbed_nano")
	require.Equal(t, profile2.GetPlatforms()[1].GetVersion(), "1.2.3")
	require.Equal(t, profile2.GetPlatforms()[1].GetIndexUrl(), "https://test.com/mbed_nano_index.json")
	require.Len(t, profile2.GetLibraries(), 2)
	require.Equal(t, profile2.GetLibraries()[0].GetIndexLibrary().GetName(), "ArduinoIoTCloud")
	require.Equal(t, profile2.GetLibraries()[0].GetIndexLibrary().GetVersion(), "1.0.2")
	require.Equal(t, profile2.GetLibraries()[1].GetLocalLibrary().GetPath(), "libraries/Arduino_ConnectionHandler")
}
