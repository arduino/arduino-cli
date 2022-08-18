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
	require.Contains(t, string(stdout), "#compdef _arduino-cli arduino-cli")
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
	require.Contains(t, string(stdout), "#compdef _arduino-cli arduino-cli")
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
