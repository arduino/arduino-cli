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

package completion_test

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

// test if the completion command behaves correctly
func TestCompletionNoArgs(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion")
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error: accepts 1 arg(s), received 0")
	require.Empty(t, stdout)
}

func TestCompletionBash(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "bash")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "# bash completion V2 for arduino-cli")
	require.Contains(t, string(stdout), "__start_arduino-cli()")
}

func TestCompletionZsh(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "zsh")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "#compdef arduino-cli")
	require.Contains(t, string(stdout), "_arduino-cli()")
}

func TestCompletionFish(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "fish")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "# fish completion for arduino-cli")
	require.Contains(t, string(stdout), "function __arduino_cli_perform_completion")
}

func TestCompletionPowershell(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "powershell")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "# powershell completion for arduino-cli")
	require.Contains(t, string(stdout), "Register-ArgumentCompleter -CommandName 'arduino-cli' -ScriptBlock")
}

func TestCompletionBashNoDesc(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "bash", "--no-descriptions")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "# bash completion V2 for arduino-cli")
	require.Contains(t, string(stdout), "__start_arduino-cli()")
	require.Contains(t, string(stdout), "__completeNoDesc")
}

func TestCompletionZshNoDesc(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "zsh", "--no-descriptions")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "#compdef arduino-cli")
	require.Contains(t, string(stdout), "_arduino-cli()")
	require.Contains(t, string(stdout), "__completeNoDesc")
}

func TestCompletionFishNoDesc(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "fish", "--no-descriptions")
	require.NoError(t, err)
	require.Empty(t, stderr)
	require.Contains(t, string(stdout), "# fish completion for arduino-cli")
	require.Contains(t, string(stdout), "function __arduino_cli_perform_completion")
	require.Contains(t, string(stdout), "__completeNoDesc")
}

func TestCompletionPowershellNoDesc(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, stderr, err := cli.Run("completion", "powershell", "--no-descriptions")
	require.Error(t, err)
	require.Empty(t, stdout)
	require.Contains(t, string(stderr), "Error: command description is not supported by powershell")
}

// test if the completion suggestions returned are meaningful
// we use the __complete hidden command
// https://github.com/spf13/cobra/blob/master/shell_completions.md#debugging

// test static completions
func TestStaticCompletions(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, _ := cli.Run("__complete", "--format", "")
	require.Contains(t, string(stdout), "json")

	stdout, _, _ = cli.Run("__complete", "--log-format", "")
	require.Contains(t, string(stdout), "json")

	stdout, _, _ = cli.Run("__complete", "--log-level", "")
	require.Contains(t, string(stdout), "trace")
}

// here we test if the completions coming from the core are working
func TestConfigCompletion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	stdout, _, _ := cli.Run("__complete", "config", "add", "")
	require.Contains(t, string(stdout), "board_manager.additional_urls")
	stdout, _, _ = cli.Run("__complete", "config", "remove", "")
	require.Contains(t, string(stdout), "board_manager.additional_urls")
	stdout, _, _ = cli.Run("__complete", "config", "delete", "")
	require.Contains(t, string(stdout), "board_manager.additional_urls")
	stdout, _, _ = cli.Run("__complete", "config", "set", "")
	require.Contains(t, string(stdout), "board_manager.additional_urls")
}

// here we test if the completions coming from the libs are working
func TestLibCompletion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("lib", "update-index")
	require.NoError(t, err)
	stdout, _, _ := cli.Run("__complete", "lib", "install", "")
	require.Contains(t, string(stdout), "WiFi101")
	stdout, _, _ = cli.Run("__complete", "lib", "download", "")
	require.Contains(t, string(stdout), "WiFi101")
	stdout, _, _ = cli.Run("__complete", "lib", "uninstall", "")
	require.NotContains(t, string(stdout), "WiFi101") // not yet installed

	_, _, err = cli.Run("lib", "install", "Wifi101")
	require.NoError(t, err)
	stdout, _, _ = cli.Run("__complete", "lib", "uninstall", "")
	require.Contains(t, string(stdout), "WiFi101")
	stdout, _, _ = cli.Run("__complete", "lib", "examples", "")
	require.Contains(t, string(stdout), "WiFi101")
	stdout, _, _ = cli.Run("__complete", "lib", "deps", "")
	require.Contains(t, string(stdout), "WiFi101")
}

// here we test if the completions coming from the core are working
func TestCoreCompletion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	stdout, _, _ := cli.Run("__complete", "core", "install", "")
	require.Contains(t, string(stdout), "arduino:avr")
	stdout, _, _ = cli.Run("__complete", "core", "download", "")
	require.Contains(t, string(stdout), "arduino:avr")
	stdout, _, _ = cli.Run("__complete", "core", "uninstall", "")
	require.NotContains(t, string(stdout), "arduino:avr")

	// we install a core because the provided completions comes from it
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, _ = cli.Run("__complete", "core", "uninstall", "")
	require.Contains(t, string(stdout), "arduino:avr")

	stdout, _, _ = cli.Run("__complete", "board", "details", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "burn-bootloader", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "compile", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "debug", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "lib", "examples", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "upload", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")
	stdout, _, _ = cli.Run("__complete", "monitor", "-b", "")
	require.Contains(t, string(stdout), "arduino:avr:uno")

	// -l/--protocol and -p/--port cannot be tested because there are
	// no board connected.

	stdout, _, _ = cli.Run("__complete", "burn-bootloader", "-P", "")
	require.Contains(t, string(stdout), "atmel_ice")
	stdout, _, _ = cli.Run("__complete", "compile", "-P", "")
	require.Contains(t, string(stdout), "atmel_ice")
	stdout, _, _ = cli.Run("__complete", "debug", "-P", "")
	require.Contains(t, string(stdout), "atmel_ice")
	stdout, _, _ = cli.Run("__complete", "upload", "-P", "")
	require.Contains(t, string(stdout), "atmel_ice")
}

func TestProfileCompletion(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Create test sketches
	sketchWithProfilesPath, err := paths.New("testdata", "SketchWithProfiles").Abs()
	require.NoError(t, err)
	require.True(t, sketchWithProfilesPath.IsDir())

	stdout, _, _ := cli.Run("__complete", "compile", sketchWithProfilesPath.String(), "--profile", "")
	require.Contains(t, string(stdout), "profile1")
	stdout, _, _ = cli.Run("__complete", "monitor", sketchWithProfilesPath.String(), "--profile", "")
	require.Contains(t, string(stdout), "profile1")
	stdout, _, _ = cli.Run("__complete", "upload", sketchWithProfilesPath.String(), "--profile", "")
	require.Contains(t, string(stdout), "profile1")

	// The cli is running in the sketch folder, so need the explicitly specify the path in the cli
	cli.SetWorkingDir(sketchWithProfilesPath)
	stdout, _, _ = cli.Run("__complete", "compile", "--profile", "")
	require.Contains(t, string(stdout), "profile1")
	stdout, _, _ = cli.Run("__complete", "monitor", "--profile", "")
	require.Contains(t, string(stdout), "profile1")
	stdout, _, _ = cli.Run("__complete", "upload", "--profile", "")
	require.Contains(t, string(stdout), "profile1")

}
