// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package daemon

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func TestBoardListMock(t *testing.T) {
	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)

	var tmp1, tmp2 string

	{
		// Create a new instance of the daemon
		grpcInst := cli.Create()
		require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
			fmt.Printf("INIT> %v\n", ir.GetMessage())
		}))

		// Run a BoardList
		resp, err := grpcInst.BoardList(time.Second)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Ports)
		for _, port := range resp.Ports {
			if port.GetPort().GetProtocol() == "serial" {
				tmp1 = port.Port.GetProperties()["discovery_tmp"]
			}
		}
		require.NotEmpty(t, tmp1)

		// Close instance
		require.NoError(t, grpcInst.Destroy(t.Context()))
	}

	{
		// Create a second instance of the daemon
		grpcInst := cli.Create()
		require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
			fmt.Printf("INIT> %v\n", ir.GetMessage())
		}))

		// Run a BoardList
		var resp *commands.BoardListResponse
		for range 5 { // wait up to 5 seconds
			resp, err = grpcInst.BoardList(time.Second)
			require.NoError(t, err)
			if len(resp.Ports) > 0 {
				break
			}
			time.Sleep(time.Second)
		}
		require.NotEmpty(t, resp.Ports)
		for _, port := range resp.Ports {
			if port.GetPort().GetProtocol() == "serial" {
				tmp2 = port.Port.GetProperties()["discovery_tmp"]
			}
		}
		require.NotEmpty(t, tmp2)

		// Close instance
		require.NoError(t, grpcInst.Destroy(t.Context()))
	}

	// Check if the discoveries have been successfully close
	for range 5 { // wait up to 5 seconds
		_, err1 := os.Lstat(tmp1)
		_, err2 := os.Lstat(tmp2)
		if err1 != nil && err2 != nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.NoFileExists(t, tmp1, "discovery has not been closed")
	require.NoFileExists(t, tmp2, "discovery has not been closed")
}
