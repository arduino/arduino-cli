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

package board_test

import (
	"runtime"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestHardwareLoading(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// install two cores, boards must be ordered by package name and platform name
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	// _, _, err = cli.Run("core", "install", "arduino:sam")
	// require.NoError(t, err)

	localTxt, err := paths.New("testdata", "custom_local_txts").Abs()
	require.NoError(t, err)
	downloadedHardwareAvr := cli.DataDir().Join("packages", "arduino", "hardware", "avr", "1.8.6")
	localTxt.Join("boards.local.txt").CopyTo(downloadedHardwareAvr.Join("boards.local.txt"))
	localTxt.Join("platform.local.txt").CopyTo(downloadedHardwareAvr.Join("platform.local.txt"))

	t.Run("Simple", func(t *testing.T) {
		{
			out, _, err := cli.Run("core", "list", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.LengthMustEqualTo(1)
			jsonOut.MustContain(`[
				{
					"id": "arduino:avr",
					"installed_version": "1.8.6",
					"releases": {
						"1.8.6": {
							"name": "Arduino AVR Boards",
							"boards": [
								{
									"name": "Arduino Uno",
									"fqbn": "arduino:avr:uno"
								},
								{
									"name": "Arduino Yún",
									"fqbn": "arduino:avr:yun"
								}
							]
						}
					}
				}
			]`)
		}

		{
			// Also test local platform.txt properties override
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:uno", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "uno",
				"build_properties": [
					"_id=uno",
					"tools.avrdude.bootloader.params.verbose=-v",
					"tools.avrdude.cmd.path=/my/personal/avrdude"
				],
				"programmers": [
					{
					"platform": "Arduino AVR Boards",
					"id": "usbasp",
					"name": "USBasp"
					},
					{
					"platform": "Arduino AVR Boards",
					"id": "avrispmkii",
					"name": "AVRISP mkII"
					}
				]
			}`)
		}

		{
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:yun", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "yun",
				"build_properties": [
					"_id=yun",
					"upload.wait_for_upload_port=true"
				]
			}`)
		}

		{
			// Check un-expansion of board_properties
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:robotMotor", "--show-properties=unexpanded", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "robotMotor",
				"build_properties": [
					"_id=robotMotor",
					"build.extra_flags={build.usb_flags}",
					"upload.wait_for_upload_port=true"
				]
			}`)
		}

		{
			// Also test local boards.txt properties override
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:diecimila", "--show-properties=unexpanded", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "diecimila",
				"build_properties": [
					"_id=diecimila",
					"menu.cpu.atmega123=ATmega123"
				]
			}`)
		}
	})

	t.Run("MixingUserHardware", func(t *testing.T) {
		// Install custom hardware required for tests
		customHwDir, err := paths.New("..", "testdata", "user_hardware").Abs()
		require.NoError(t, err)
		require.NoError(t, customHwDir.CopyDirTo(cli.SketchbookDir().Join("hardware")))

		{
			out, _, err := cli.Run("core", "list", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			if runtime.GOOS == "windows" {
				// a package is a symlink, and windows does not support them
				jsonOut.LengthMustEqualTo(2)
			} else {
				jsonOut.LengthMustEqualTo(3)
			}
			jsonOut.MustContain(`[
				{
					"id": "arduino:avr",
					"installed_version": "1.8.6",
					"releases": {
						"1.8.6": {
							"name": "Arduino AVR Boards",
							"boards": [
								{
									"name": "Arduino Uno",
									"fqbn": "arduino:avr:uno"
								},
								{
									"name": "Arduino Yún",
									"fqbn": "arduino:avr:yun"
								}
							]
						}
					}
				}
			]`)
			jsonOut.MustContain(`[
				{
					"id": "my_avr_platform:avr",
					"installed_version": "9.9.9",
					"releases": {
						"9.9.9": {
							"name": "My AVR Boards",
							"missing_metadata": true,
							"boards": [
								{
									"name": "Arduino Yún",
									"fqbn": "my_avr_platform:avr:custom_yun"
								}
							]
						}
					},
					"manually_installed": true
				}
			]`)

			//		require.False(t, myAVRPlatformAvrArch.Properties.ContainsKey("preproc.includes.flags"))

			if runtime.GOOS != "windows" {
				jsonOut.MustContain(`[
					{
						"id": "my_symlinked_avr_platform:avr",
						"manually_installed": true,
						"releases": {
							"9.9.9": {
								"missing_metadata": true
							}
						}
					}
				]`)
			}
		}

		{
			// Also test local platform.txt properties override
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:uno", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "uno",
				"build_properties": [
					"_id=uno",
					"tools.avrdude.bootloader.params.verbose=-v",
					"tools.avrdude.cmd.path=/my/personal/avrdude"
				],
				"programmers": [
					{
						"platform": "Arduino AVR Boards",
						"id": "usbasp",
						"name": "USBasp"
					},
					{
						"platform": "Arduino AVR Boards",
						"id": "avrispmkii",
						"name": "AVRISP mkII"
					}
				]
			}`)
		}

		{
			out, _, err := cli.Run("board", "details", "-b", "arduino:avr:yun", "--show-properties=unexpanded", "--format", "json")
			require.NoError(t, err)
			jsonOut := requirejson.Parse(t, out)
			jsonOut.MustContain(`{
				"version": "1.8.6",
				"properties_id": "yun",
				"build_properties": [
					"_id=yun",
					"upload.wait_for_upload_port=true",
					"preproc.includes.flags=-w -x c++ -M -MG -MP",
					"preproc.macros.flags=-w -x c++ -E -CC",
					"recipe.preproc.includes=\"{compiler.path}{compiler.cpp.cmd}\" {compiler.cpp.flags} {preproc.includes.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.cpp.extra_flags} {build.extra_flags} {includes} \"{source_file}\""
				]
			}`)
			jsonOut.Query(`isempty( .build_properties[] | select(startswith("preproc.macros.compatibility_flags")) )`).MustEqual("true")
		}
	})
}
