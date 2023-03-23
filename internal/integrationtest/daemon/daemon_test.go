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
	"time"

	"github.com/arduino/arduino-cli/arduino"
	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/go-paths-helper"

	"github.com/stretchr/testify/require"
)

func TestArduinoCliDaemon(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/pull/1804

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Run a one-shot board list
	boardListResp, err := grpcInst.BoardList(time.Second)
	require.NoError(t, err)
	fmt.Printf("Got boardlist response with %d ports\n", len(boardListResp.GetPorts()))

	// Run a one-shot board list again (should not fail)
	boardListResp, err = grpcInst.BoardList(time.Second)
	require.NoError(t, err)
	fmt.Printf("Got boardlist response with %d ports\n", len(boardListResp.GetPorts()))

	testWatcher := func() {
		// Run watcher
		watcher, err := grpcInst.BoardListWatch()
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()
			for {
				msg, err := watcher.Recv()
				if err == io.EOF {
					fmt.Println("Watcher EOF")
					return
				}
				require.Empty(t, msg.Error, "Board list watcher returned an error")
				require.NoError(t, err, "BoardListWatch grpc call returned an error")
				fmt.Printf("WATCH> %v\n", msg)
			}
		}()
		time.Sleep(time.Second)
		require.NoError(t, watcher.CloseSend())
		select {
		case <-ctx.Done():
			// all right!
		case <-time.After(time.Second):
			require.Fail(t, "BoardListWatch didn't close")
		}
	}

	testWatcher()
	testWatcher()
}

func TestDaemonAutoUpdateIndexOnFirstInit(t *testing.T) {
	// https://github.com/arduino/arduino-cli/issues/1529

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	_, err := grpcInst.PlatformList(context.Background())
	require.NoError(t, err)

	require.FileExists(t, cli.DataDir().Join("package_index.json").String())
}

// createEnvForDaemon performs the minimum required operations to start the arduino-cli daemon.
// It returns a testsuite.Environment and an ArduinoCLI client to perform the integration tests.
// The Environment must be disposed by calling the CleanUp method via defer.
func createEnvForDaemon(t *testing.T) (*integrationtest.Environment, *integrationtest.ArduinoCLI) {
	env := integrationtest.NewEnvironment(t)

	cli := integrationtest.NewArduinoCliWithinEnvironment(env, &integrationtest.ArduinoCLIConfig{
		ArduinoCLIPath:         integrationtest.FindRepositoryRootPath(t).Join("arduino-cli"),
		UseSharedStagingFolder: true,
	})

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)

	_ = cli.StartDaemon(false)
	return env, cli
}

func TestDaemonCompileOptions(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1614
	// See: https://github.com/arduino/arduino-cli/pull/1820

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	plInst, err := grpcInst.PlatformInstall(context.Background(), "arduino", "avr", "1.8.5", true)
	require.NoError(t, err)
	for {
		msg, err := plInst.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		fmt.Printf("INSTALL> %v\n", msg)
	}

	// Install boards.local.txt to trigger bug
	platformLocalTxt := paths.New("testdata", "boards.local.txt-issue1614")
	err = platformLocalTxt.CopyTo(cli.DataDir().
		Join("packages", "arduino", "hardware", "avr", "1.8.5", "boards.local.txt"))
	require.NoError(t, err)

	// Re-init instance to update changes
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Build sketch (with errors)
	sk := paths.New("testdata", "bare_minimum")
	compile, err := grpcInst.Compile(context.Background(), "arduino:avr:uno:some_menu=bad", sk.String(), "")
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if err == io.EOF {
			require.FailNow(t, "Expected compilation failure", "compilation succeeded")
			break
		}
		if err != nil {
			fmt.Println("COMPILE ERROR>", err)
			break
		}
		if msg.ErrStream != nil {
			fmt.Printf("COMPILE> %v\n", string(msg.GetErrStream()))
		}
	}

	// Build sketch (without errors)
	compile, err = grpcInst.Compile(context.Background(), "arduino:avr:uno:some_menu=good", sk.String(), "")
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if msg.ErrStream != nil {
			fmt.Printf("COMPILE> %v\n", string(msg.GetErrStream()))
		}
	}
}

func TestDaemonCompileAfterFailedLibInstall(t *testing.T) {
	// See: https://github.com/arduino/arduino-cli/issues/1812

	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))

	// Build sketch (with errors)
	sk := paths.New("testdata", "bare_minimum")
	compile, err := grpcInst.Compile(context.Background(), "", sk.String(), "")
	require.NoError(t, err)
	for {
		msg, err := compile.Recv()
		if err == io.EOF {
			require.FailNow(t, "Expected compilation failure", "compilation succeeded")
			break
		}
		if err != nil {
			fmt.Println("COMPILE ERROR>", err)
			require.Contains(t, err.Error(), "Missing FQBN")
			break
		}
		if msg.ErrStream != nil {
			fmt.Printf("COMPILE> %v\n", string(msg.GetErrStream()))
		}
	}
}

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
		res, err := analyzeUpdateIndexClient(t, cl)
		require.NoError(t, err)
		require.Len(t, res, 1)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Success)
	}
	{
		cl, err := grpcInst.UpdateIndex(context.Background(), false)
		require.NoError(t, err)
		res, err := analyzeUpdateIndexClient(t, cl)
		require.Error(t, err)
		require.Len(t, res, 3)
		require.True(t, res["https://downloads.arduino.cc/packages/package_index.tar.bz2"].Success)
		require.True(t, res["http://arduino.esp8266.com/stable/package_esp8266com_index.json"].Success)
		require.False(t, res["http://downloads.arduino.cc/package_inexistent_index.json"].Success)
	}
}

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
	installedEthernetVersion := ""
	{
		resp, err := grpcInst.LibraryList(context.Background(), "", "", true, false)
		require.NoError(t, err)
		libsAndLocation := map[string]commands.LibraryLocation{}
		for _, lib := range resp.GetInstalledLibraries() {
			libsAndLocation[lib.Library.Name] = lib.Library.Location
			if lib.Library.Name == "Ethernet" {
				installedEthernetVersion = lib.Library.Version
			}
		}
		require.Contains(t, libsAndLocation, "Ethernet")
		require.Contains(t, libsAndLocation, "SD")
		require.Contains(t, libsAndLocation, "Firmata")
		require.Equal(t, libsAndLocation["Ethernet"], commands.LibraryLocation_LIBRARY_LOCATION_USER)
		require.Equal(t, libsAndLocation["SD"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
		require.Equal(t, libsAndLocation["Firmata"], commands.LibraryLocation_LIBRARY_LOCATION_BUILTIN)
	}

	// Remove library from sketchbook
	{
		uninstCl, err := grpcInst.LibraryUninstall(context.Background(), "Ethernet", installedEthernetVersion)
		require.NoError(t, err)
		for {
			msg, err := uninstCl.Recv()
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

func TestDaemonLibrariesRescanOnInstall(t *testing.T) {
	/*
		Ensures that the libraries are rescanned prior to installing a new one,
		to avoid clashes with libraries installed after the daemon initialization.
		To perform the check:
		 - the daemon is run and a gprc instance initialized
		 - a library is installed through the cli
		 - an attempt to install a new version of the library is done
		   with the gprc instance
		The last attempt is expected to not raise an error
	*/
	env, cli := createEnvForDaemon(t)
	defer env.CleanUp()

	grpcInst := cli.Create()
	require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {
		fmt.Printf("INIT> %v\n", ir.GetMessage())
	}))
	cli.Run("lib", "install", "SD@1.2.3")

	instCl, err := grpcInst.LibraryInstall(context.Background(), "SD", "1.2.4", false, false, true)

	require.NoError(t, err)
	for {
		_, err := instCl.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}

}

func TestDaemonCoreUpgradePlatform(t *testing.T) {
	refreshInstance := func(t *testing.T, grpcInst *integrationtest.ArduinoCLIInstance) {
		require.NoError(t, grpcInst.Init("", "", func(ir *commands.InitResponse) {}))
	}
	updateIndexAndInstallPlatform := func(cli *integrationtest.ArduinoCLI, grpcInst *integrationtest.ArduinoCLIInstance, version string) {
		refreshInstance(t, grpcInst)

		// adding the additional urls
		err := cli.SetValue("board_manager.additional_urls", `["https://arduino.esp8266.com/stable/package_esp8266com_index.json"]`)
		require.NoError(t, err)

		cl, err := grpcInst.UpdateIndex(context.Background(), false)
		require.NoError(t, err)
		res, err := analyzeUpdateIndexClient(t, cl)
		require.NoError(t, err)
		require.Len(t, res, 2)
		require.True(t, res["https://arduino.esp8266.com/stable/package_esp8266com_index.json"].Success)

		refreshInstance(t, grpcInst)

		// installing outdated version
		plInst, err := grpcInst.PlatformInstall(context.Background(), "esp8266", "esp8266", version, true)
		require.NoError(t, err)
		for {
			_, err := plInst.Recv()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
		}
	}

	t.Run("upgraded successfully with additional urls", func(t *testing.T) {
		t.Run("and install.json is present", func(t *testing.T) {
			env, cli := createEnvForDaemon(t)
			defer env.CleanUp()

			grpcInst := cli.Create()
			updateIndexAndInstallPlatform(cli, grpcInst, "3.1.0")

			plUpgrade, err := grpcInst.PlatformUpgrade(context.Background(), "esp8266", "esp8266", true)
			require.NoError(t, err)

			platform, upgradeError := analyzePlatformUpgradeClient(plUpgrade)
			require.NoError(t, upgradeError)
			require.NotNil(t, platform)
			require.True(t, platform.Indexed)          // the esp866 is present in the additional-urls
			require.False(t, platform.MissingMetadata) // install.json is present
		})
		t.Run("and install.json is missing", func(t *testing.T) {
			env, cli := createEnvForDaemon(t)
			defer env.CleanUp()

			grpcInst := cli.Create()
			updateIndexAndInstallPlatform(cli, grpcInst, "3.1.0")

			// remove installed.json
			x := env.RootDir().Join("A/packages/esp8266/hardware/esp8266/3.1.0/installed.json")
			require.NoError(t, x.Remove())

			plUpgrade, err := grpcInst.PlatformUpgrade(context.Background(), "esp8266", "esp8266", true)
			require.NoError(t, err)

			platform, upgradeError := analyzePlatformUpgradeClient(plUpgrade)
			require.NoError(t, upgradeError)
			require.NotNil(t, platform)
			require.True(t, platform.Indexed)          // the esp866 is not present in the additional-urls
			require.False(t, platform.MissingMetadata) // install.json is present because the old version got upgraded

		})
	})

	t.Run("upgrade failed", func(t *testing.T) {
		t.Run("without additional URLs", func(t *testing.T) {
			env, cli := createEnvForDaemon(t)
			defer env.CleanUp()

			grpcInst := cli.Create()
			updateIndexAndInstallPlatform(cli, grpcInst, "3.1.0")

			// remove esp8266 from the additional-urls
			require.NoError(t, cli.SetValue("board_manager.additional_urls", `[]`))
			refreshInstance(t, grpcInst)

			plUpgrade, err := grpcInst.PlatformUpgrade(context.Background(), "esp8266", "esp8266", true)
			require.NoError(t, err)

			platform, upgradeError := analyzePlatformUpgradeClient(plUpgrade)
			require.ErrorIs(t, upgradeError, (&arduino.PlatformAlreadyAtTheLatestVersionError{Platform: "esp8266:esp8266"}).ToRPCStatus().Err())
			require.NotNil(t, platform)
			require.False(t, platform.Indexed)         // the esp866 is not present in the additional-urls
			require.False(t, platform.MissingMetadata) // install.json is present
		})
		t.Run("missing both additional URLs and install.json", func(t *testing.T) {
			env, cli := createEnvForDaemon(t)
			defer env.CleanUp()

			grpcInst := cli.Create()
			updateIndexAndInstallPlatform(cli, grpcInst, "3.1.0")

			// remove additional urls and installed.json
			{
				require.NoError(t, cli.SetValue("board_manager.additional_urls", `[]`))
				refreshInstance(t, grpcInst)

				x := env.RootDir().Join("A/packages/esp8266/hardware/esp8266/3.1.0/installed.json")
				require.NoError(t, x.Remove())
			}

			plUpgrade, err := grpcInst.PlatformUpgrade(context.Background(), "esp8266", "esp8266", true)
			require.NoError(t, err)

			platform, upgradeError := analyzePlatformUpgradeClient(plUpgrade)
			require.ErrorIs(t, upgradeError, (&arduino.PlatformAlreadyAtTheLatestVersionError{Platform: "esp8266:esp8266"}).ToRPCStatus().Err())
			require.NotNil(t, platform)
			require.False(t, platform.Indexed)        // the esp866 is not present in the additional-urls
			require.True(t, platform.MissingMetadata) // install.json is present
		})
	})
}

func analyzeUpdateIndexClient(t *testing.T, cl commands.ArduinoCoreService_UpdateIndexClient) (map[string]*commands.DownloadProgressEnd, error) {
	analyzer := NewDownloadProgressAnalyzer(t)
	for {
		msg, err := cl.Recv()
		if err == io.EOF {
			return analyzer.Results, nil
		}
		if err != nil {
			return analyzer.Results, err
		}
		require.NoError(t, err)
		analyzer.Process(msg.GetDownloadProgress())
	}
}

func analyzePlatformUpgradeClient(cl commands.ArduinoCoreService_PlatformUpgradeClient) (*commands.Platform, error) {
	var platform *commands.Platform
	var upgradeError error
	for {
		msg, err := cl.Recv()
		if err == io.EOF {
			break
		}
		if msg.GetPlatform() != nil {
			platform = msg.GetPlatform()
		}
		if err != nil {
			upgradeError = err
			break
		}
	}
	return platform, upgradeError
}
