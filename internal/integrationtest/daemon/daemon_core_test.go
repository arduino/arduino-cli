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

func TestDaemonCoreUpdateIndex(t *testing.T) {
	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Set extra indexes
	err := cli.SetValue(
		"board_manager.additional_urls", ""+
			`["http://arduino.esp8266.com/stable/package_esp8266com_index.json",`+
			` "http://downloads.arduino.cc/package_inexistent_index.json"]`)
	require.NoError(t, err)

	analyzeUpdateIndexClient := func(cl commands.ArduinoCoreService_UpdateIndexClient) (error, map[string]*commands.DownloadProgressEnd) {
		analyzer := NewDownloadProgressAnalyzer(t)
		for {
			msg, err := cl.Recv()
			// fmt.Println("DOWNLOAD>", msg)
			if err == io.EOF {
				return nil, analyzer.Results
			}
			if err != nil {
				return err, analyzer.Results
			}
			require.NoError(t, err)
			analyzer.Process(msg.GetDownloadProgress())
		}
	}

	{
		cl, err := grpcInst.UpdateIndex(context.Background(), true)
		require.NoError(t, err)
		err, res := analyzeUpdateIndexClient(cl)
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Success)
	}
	{
		cl, err := grpcInst.UpdateIndex(context.Background(), false)
		require.NoError(t, err)
		err, res := analyzeUpdateIndexClient(cl)
		require.Error(t, err)
		require.Len(t, res, 3)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Success)
		require.True(t, res["http://arduino.esp8266.com/stable/package_esp8266com_index.json"].Success)
		require.False(t, res["http://downloads.arduino.cc/package_inexistent_index.json"].Success)
	}
}
