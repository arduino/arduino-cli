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

package monitor_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestMonitorGRPCClose(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2271

	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)
	cli.InstallMockedSerialMonitor(t)

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Run a one-shot board list
	boardListResp, err := grpcInst.BoardList(time.Second)
	require.NoError(t, err)
	ports := boardListResp.GetPorts()
	require.NotEmpty(t, ports)
	fmt.Printf("Got boardlist response with %d ports\n", len(ports))

	// Open mocked serial-monitor and close it client-side
	tmpFileMatcher := regexp.MustCompile("Tmpfile: (.*)\n")
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		mon, err := grpcInst.Monitor(ctx, ports[0].GetPort())
		var tmpFile *paths.Path
		for {
			monResp, err := mon.Recv()
			if err != nil {
				fmt.Println("MON>", err)
				break
			}
			fmt.Printf("MON> %v\n", monResp)
			if rx := monResp.GetRxData(); rx != nil {
				if matches := tmpFileMatcher.FindAllStringSubmatch(string(rx), -1); len(matches) > 0 {
					fmt.Println("Found tmpFile", matches[0][1])
					tmpFile = paths.New(matches[0][1])
				}
			}
		}
		require.NotNil(t, tmpFile)
		// The port is close client-side, it may be still open server-side
		require.True(t, tmpFile.Exist())
		cancel()
		require.NoError(t, err)
	}

	// Now close the monitor using MonitorRequest_Close
	{
		// Keep a timeout to allow the test to exit in any case
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		mon, err := grpcInst.Monitor(ctx, ports[0].GetPort())
		var tmpFile *paths.Path
		for {
			monResp, err := mon.Recv()
			if errors.Is(err, io.EOF) {
				fmt.Println("MON>", err)
				break
			}

			require.NoError(t, err)
			fmt.Printf("MON> %v\n", monResp)
			if rx := monResp.GetRxData(); rx != nil {
				if matches := tmpFileMatcher.FindAllStringSubmatch(string(rx), -1); len(matches) > 0 {
					fmt.Println("Found tmpFile", matches[0][1])
					tmpFile = paths.New(matches[0][1])
					go func() {
						time.Sleep(time.Second)
						fmt.Println("<MON Sent close command")
						mon.Send(&commands.MonitorRequest{Message: &commands.MonitorRequest_Close{Close: true}})
					}()
				}
			}
		}
		require.NotNil(t, tmpFile)
		// The port is closed serverd-side, it must be already closed once the client has received the EOF
		require.False(t, tmpFile.Exist())
		cancel()
		require.NoError(t, err)
	}
}

func TestMonitorGRPCAppliedSettings(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/2965

	env, cli := integrationtest.CreateEnvForDaemon(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	cli.InstallMockedSerialDiscovery(t)
	cli.InstallMockedSerialMonitor(t)

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	boardListResp, err := grpcInst.BoardList(time.Second)
	require.NoError(t, err)
	ports := boardListResp.GetPorts()
	require.NotEmpty(t, ports)

	// recvUntilAppliedSettings reads from the monitor stream, skipping rxData and other
	// messages, until an applied_settings response is received or the test times out.
	recvUntilAppliedSettings := func(mon commands.ArduinoCoreService_MonitorClient) *commands.MonitorPortConfiguration {
		t.Helper()
		for {
			monResp, err := mon.Recv()
			require.NoError(t, err)
			fmt.Printf("MON> %v\n", monResp)
			if as := monResp.GetAppliedSettings(); as != nil {
				return as
			}
		}
	}

	// Open the monitor with a non-default baudrate (115200 vs the default 9600).
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mon, err := grpcInst.MonitorWithConfig(ctx, ports[0].GetPort(), &commands.MonitorPortConfiguration{
		Settings: []*commands.MonitorPortSetting{
			{SettingId: "baudrate", Value: "115200"},
		},
	})
	require.NoError(t, err)

	// The server must emit applied_settings right after opening the port.
	openApplied := recvUntilAppliedSettings(mon)
	openSettings := map[string]string{}
	for _, s := range openApplied.GetSettings() {
		openSettings[s.GetSettingId()] = s.GetValue()
	}
	require.Equal(t, "115200", openSettings["baudrate"], "applied_settings after open must reflect the configured baudrate")

	// Send an updated configuration to change the baudrate back to 9600.
	err = mon.Send(&commands.MonitorRequest{
		Message: &commands.MonitorRequest_UpdatedConfiguration{
			UpdatedConfiguration: &commands.MonitorPortConfiguration{
				Settings: []*commands.MonitorPortSetting{
					{SettingId: "baudrate", Value: "9600"},
				},
			},
		},
	})
	require.NoError(t, err)

	// The server must emit applied_settings after the configuration update.
	updateApplied := recvUntilAppliedSettings(mon)
	updateSettings := map[string]string{}
	for _, s := range updateApplied.GetSettings() {
		updateSettings[s.GetSettingId()] = s.GetValue()
	}
	require.Equal(t, "9600", updateSettings["baudrate"], "applied_settings after update must reflect the new baudrate")

	// Close the monitor to allow cleanup of the env (otherwise on Windows the
	// tmp monitor.exe cannot be deleted because it's still open).
	cancel()
	_, err = mon.Recv()
	require.EqualError(t, err, "rpc error: code = Canceled desc = context canceled")
	// the mocked serial-monitor is designed to delay 2 seconds before closing,
	// let's wait a bit to ensure the monitor process has exited before the test ends and the env is cleaned up.
	time.Sleep(3 * time.Second)
}
