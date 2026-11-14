// This file is part of arduino-cli.
//
// Copyright 2026 ARDUINO SA (http://www.arduino.cc/)
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

package read

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestReadFlags(t *testing.T) {
	cmd := NewCommand(nil)
	require.ErrorContains(t, cmd.ValidateRequiredFlags(), "fqbn")
	require.NoError(t, cmd.ParseFlags([]string{"--fqbn", "vendor:avr:board"}))
	require.ErrorContains(t, cmd.ValidateRequiredFlags(), "output")
	require.NoError(t, cmd.ParseFlags([]string{"--output", "backup.bin"}))
	require.NoError(t, cmd.ValidateRequiredFlags())
	require.Equal(t, "flash", cmd.Flag("memory").Value.String())
	require.Equal(t, "bin", cmd.Flag("file-format").Value.String())
	require.Error(t, cmd.Args(cmd, []string{"sketch"}))
	// File encoding must not shadow the CLI-wide diagnostic format flag.
	root := &cobra.Command{Use: "arduino-cli"}
	root.PersistentFlags().String("format", "text", "")
	root.AddCommand(cmd)
	require.NoError(t, cmd.ParseFlags([]string{"--format", "json", "--file-format", "ihex"}))
	require.Equal(t, "json", root.Flag("format").Value.String())
	require.Equal(t, "ihex", cmd.Flag("file-format").Value.String())
	require.Equal(t, "bin", NewCommand(nil).Flag("file-format").Value.String())
}

func TestReadResult(t *testing.T) {
	r := &readResult{OutputFile: "backup.bin", Size: 1024, Stdout: "diagnostic"}
	require.Contains(t, r.String(), "1024")
	require.Contains(t, r.String(), "backup.bin")
	data, err := json.Marshal(r.Data())
	require.NoError(t, err)
	require.JSONEq(t, `{"output_file":"backup.bin","size":1024,"dry_run":false,"stdout":"diagnostic","stderr":""}`, string(data))
	r.DryRun = true
	require.Contains(t, r.String(), "no output file created")
}
