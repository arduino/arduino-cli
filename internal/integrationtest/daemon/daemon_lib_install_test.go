// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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

package daemon_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func TestDaemonBundleLibInstall(t *testing.T) {
	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Install libraries in bundled dir
	{
		instCl, err := grpcInst.LibraryInstall(context.Background(), "Arduino_BuiltIn", "", false, false, true)
		require.NoError(t, err)
		for {
			msg, err := instCl.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			fmt.Printf("LIB INSTALL> %+v\n", msg)
		}
	}

	// Check if libraries are installed as expected
	{
		resp, err := grpcInst.LibraryList(context.Background(), "", "", true, false)
		require.NoError(t, err)
		libsAndLocation := map[string]commands.LibraryLocation{}
		for _, lib := range resp.GetInstalledLibraries() {
			libsAndLocation[lib.Library.Name] = lib.Library.Location
		}
		require.Contains(t, libsAndLocation, "Ethernet")
		require.Contains(t, libsAndLocation, "SD")
		require.Contains(t, libsAndLocation, "Firmata")
		require.Equal(t, libsAndLocation["Ethernet"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
		require.Equal(t, libsAndLocation["SD"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
		require.Equal(t, libsAndLocation["Firmata"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
	}

	// Install a library in sketchbook to override bundled
	{
		instCl, err := grpcInst.LibraryInstall(context.Background(), "Ethernet", "", false, false, false)
		require.NoError(t, err)
		for {
			msg, err := instCl.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			fmt.Printf("LIB INSTALL> %+v\n", msg)
		}
	}

	// Check if libraries are installed as expected
	{
		resp, err := grpcInst.LibraryList(context.Background(), "", "", true, false)
		require.NoError(t, err)
		libsAndLocation := map[string]commands.LibraryLocation{}
		for _, lib := range resp.GetInstalledLibraries() {
			libsAndLocation[lib.Library.Name] = lib.Library.Location
		}
		require.Contains(t, libsAndLocation, "Ethernet")
		require.Contains(t, libsAndLocation, "SD")
		require.Contains(t, libsAndLocation, "Firmata")
		require.Equal(t, libsAndLocation["Ethernet"], commands.LibraryLocation_LIBRARY_LOCATION_USER)
		require.Equal(t, libsAndLocation["SD"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
		require.Equal(t, libsAndLocation["Firmata"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
	}

	// Un-Set builtin libraries dir
	err := cli.SetValue("directories.builtin.libraries", `""`)
	require.NoError(t, err)

	// Re-init
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Install libraries in bundled dir (should now fail)
	{
		instCl, err := grpcInst.LibraryInstall(context.Background(), "Arduino_BuiltIn", "", false, false, true)
		require.NoError(t, err)
		for {
			msg, err := instCl.Recv()
			if err == io.EOF {
				require.FailNow(t, "LibraryInstall is supposed to fail because builtin libraries directory is not set")
			}
			if err != nil {
				fmt.Println("LIB INSTALL ERROR:", err)
				break
			}
			fmt.Printf("LIB INSTALL> %+v\n", msg)
		}
	}
}
