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

	{
		cl, err := grpcInst.UpdateIndex(context.Background(), true)
		require.NoError(t, err)
		res, err := analyzeUpdateIndexStream(t, cl)
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Successful)
	}
	{
		cl, err := grpcInst.UpdateIndex(context.Background(), false)
		require.NoError(t, err)
		res, err := analyzeUpdateIndexStream(t, cl)
		require.Error(t, err)
		require.Len(t, res, 3)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Successful)
		require.True(t, res["http://arduino.esp8266.com/stable/package_esp8266com_index.json"].Successful)
		require.False(t, res["http://downloads.arduino.cc/package_inexistent_index.json"].Successful)
	}
}

// analyzeUpdateIndexStream runs an update index checking if the sequence of DownloadProgress and
// DownloadResult messages is correct. It returns a map reporting all the DownloadResults messages
// received (it maps urls to DownloadResults).
func analyzeUpdateIndexStream(t *testing.T, cl commands.ArduinoCoreService_UpdateIndexClient) (map[string]*commands.DownloadResult, error) {
	ongoingDownload := ""
	results := map[string]*commands.DownloadResult{}
	for {
		msg, err := cl.Recv()
		if err == io.EOF {
			return results, nil
		}
		if err != nil {
			return results, err
		}
		require.NoError(t, err)
		fmt.Printf("UPDATE> %+v\n", msg)
		if progress := msg.GetDownloadProgress(); progress != nil {
			if progress.Url != "" && progress.Url != ongoingDownload {
				require.Empty(t, ongoingDownload, "DownloadProgress: initiated a new download with closing the previous one")
				ongoingDownload = progress.Url
			}
			if progress.Completed {
				require.NotEmpty(t, ongoingDownload, "DownloadProgress: sent a 'completed' download message without starting it")
				ongoingDownload = ""
			}
			if progress.Downloaded > 0 {
				require.NotEmpty(t, ongoingDownload, "DownloadProgress: sent an update but never initiated a download")
			}
		} else if result := msg.GetDownloadResult(); result != nil {
			require.Empty(t, ongoingDownload, "DownloadResult: got a download result with closing it first")
			results[result.Url] = result
		} else {
			require.FailNow(t, "DownloadProgress: received a message without a Progress or a Result")
		}
	}
}
