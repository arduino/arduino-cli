// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/stretchr/testify/require"
)

func TestBoardIdentification(t *testing.T) {
	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	plInst, err := grpcInst.PlatformInstall(context.Background(), "arduino", "avr", "1.8.6", true)
	require.NoError(t, err)
	for {
		msg, err := plInst.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		fmt.Printf("INSTALL> %v\n", msg)
	}

	resp, err := grpcInst.BoardIdentify(context.Background(), map[string]string{"vid": "0x2341", "pid": "0x0043"}, false)
	require.NoError(t, err)
	require.Len(t, resp.GetBoards(), 1)
	require.Equal(t, "arduino:avr:uno", resp.GetBoards()[0].GetFqbn())
}
