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

package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/commands/internal/instances"
	"github.com/arduino/arduino-cli/internal/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	paths "github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/stretchr/testify/require"
	"go.bug.st/downloader/v3"
)

// A real child process exercises tool execution without a shell or hardware.
func TestReadMemoryHelperProcess(t *testing.T) {
	for i, arg := range os.Args {
		if arg != "read-memory-helper" || i+2 >= len(os.Args) {
			continue
		}
		output, mode := os.Args[i+1], os.Args[i+2]
		if mode == "missing" {
			os.Exit(0)
		}
		if err := os.WriteFile(output, []byte(mode), 0o600); err != nil {
			os.Exit(2)
		}
		if mode == "partial" {
			os.Exit(3)
		}
		fmt.Fprint(os.Stdout, "read diagnostic")
		os.Exit(0)
	}
}

func readMemoryTestBuilder(t *testing.T) *packagemanager.Builder {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	root := t.TempDir()
	platform := filepath.Join(root, "reader", "avr")
	require.NoError(t, os.MkdirAll(platform, 0o700))
	command := `"` + filepath.ToSlash(executable) + `" -test.run=^TestReadMemoryHelperProcess$ -- read-memory-helper "{read.output_file}" `
	files := map[string]string{
		"platform.txt": "name=Read test platform\nversion=1.0.0\n" +
			"tools.reader.read.flash.bin.pattern=" + command + "flash\n" +
			"tools.reader.read.eeprom.ihex.pattern=" + command + "eeprom\n" +
			"tools.reader.read_program.flash.bin.pattern=" + command + "programmer\n" +
			"tools.reader.read.partial.bin.pattern=" + command + "partial\n" +
			"tools.reader.read.missing.bin.pattern=" + command + "missing\n" +
			"tools.reader.upload.pattern=" + command + "WRONG-upload\n" +
			"tools.reader.program.pattern=" + command + "WRONG-program\n" +
			"tools.reader.erase.pattern=" + command + "WRONG-erase\n",
		"boards.txt":      "board.name=Read test board\nboard.read.tool.default=reader\n",
		"programmers.txt": "probe.name=Read test programmer\nprobe.read_program.tool.default=reader\n",
	}
	for name, data := range files {
		require.NoError(t, os.WriteFile(filepath.Join(platform, name), []byte(data), 0o600))
	}
	pmb := packagemanager.NewBuilder(nil, nil, nil, nil, nil, "test", downloader.GetDefaultConfig())
	require.Empty(t, pmb.LoadHardwareFromDirectory(paths.New(root)))
	return pmb
}

func TestReadMemoryRecipes(t *testing.T) {
	pme, release := readMemoryTestBuilder(t).Build().NewExplorer()
	defer release()
	for _, test := range []struct {
		memory, format, programmer, want string
		dryRun, fail                     bool
	}{
		{"flash", "bin", "", "flash", false, false},
		{"eeprom", "ihex", "", "eeprom", false, false},
		{"flash", "bin", "probe", "programmer", false, false},
		{"flash", "bin", "", "", true, false},
		{"eeprom", "bin", "", "", false, true},
		{"eeprom", "ihex", "probe", "", false, true},
		{"partial", "bin", "", "", false, true},
		{"missing", "bin", "", "", false, true},
	} {
		t.Run(fmt.Sprintf("%s-%s-%s-dry%v", test.memory, test.format, test.programmer, test.dryRun), func(t *testing.T) {
			root := t.TempDir()
			output := filepath.Join(root, "backup with spaces.bin")
			o, err := newReadMemoryOptions(&rpc.ReadMemoryRequest{Memory: test.memory, Format: test.format, OutputFile: output, DryRun: test.dryRun})
			require.NoError(t, err)
			defer o.cleanup()
			var stdout bytes.Buffer
			srv := NewArduinoCoreServer().(*arduinoCoreServerImpl)
			_, err = srv.runProgramAction(context.Background(), pme, nil, "", "", "reader:avr:board", nil, test.programmer, false, false, false, &stdout, io.Discard, test.dryRun, nil, nil, o)
			if err == nil {
				_, err = o.publish()
			}
			if test.fail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if test.fail || test.dryRun {
				_, statErr := os.Stat(output)
				require.ErrorIs(t, statErr, os.ErrNotExist)
			} else {
				data, err := os.ReadFile(output)
				require.NoError(t, err)
				require.Equal(t, test.want, string(data))
			}
			if test.dryRun {
				require.Contains(t, stdout.String(), "read-memory-helper")
			}
			o.cleanup()
			entries, err := os.ReadDir(root)
			require.NoError(t, err)
			for _, entry := range entries {
				require.False(t, strings.HasPrefix(entry.Name(), ".arduino-read-"))
			}
		})
	}
}

func TestReadMemoryService(t *testing.T) {
	root := paths.New(t.TempDir())
	inst, err := instances.Create(root, root.Join("packages"), root.Join("hardware"), root.Join("downloads"), "test", downloader.GetDefaultConfig())
	require.NoError(t, err)
	defer instances.Delete(inst)
	// Register only the synthetic platform; do not initialize/download real cores
	// or start discovery tools against attached hardware.
	pm, err := instances.GetPackageManager(inst)
	require.NoError(t, err)
	readMemoryTestBuilder(t).BuildIntoExistingPackageManager(pm)
	for _, memory := range []string{"flash", "partial", "missing", "unsupported"} {
		t.Run(memory, func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "backup.bin")
			var stdout bytes.Buffer
			stream, result := ReadMemoryToServerStreams(context.Background(), &stdout, io.Discard)
			err := NewArduinoCoreServer().ReadMemory(&rpc.ReadMemoryRequest{
				Instance: inst, Fqbn: "reader:avr:board", Memory: memory, Format: "bin", OutputFile: output,
			}, stream)
			entries, readErr := os.ReadDir(dir)
			require.NoError(t, readErr)
			if memory == "flash" {
				require.NoError(t, err)
				require.Len(t, entries, 1)
				require.Equal(t, int64(5), result().GetSize())
				require.Equal(t, output, result().GetOutputFile())
				require.Equal(t, "read diagnostic", stdout.String())
			} else {
				require.Error(t, err)
				require.Nil(t, result())
				require.Empty(t, entries)
			}
		})
	}
}

func TestReadMemoryOutputProtection(t *testing.T) {
	root := t.TempDir()
	output := filepath.Join(root, "backup.bin")
	req := &rpc.ReadMemoryRequest{Memory: "flash", Format: "bin", OutputFile: output}
	o, err := newReadMemoryOptions(req)
	require.NoError(t, err)
	defer o.cleanup()
	props := properties.NewMap()
	props.Set("read.flash.bin.pattern", "unused")
	_, err = o.prepareRecipe("read", props)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(o.toolOutput, []byte("new data"), 0o600))
	// A competing writer after preflight must not be overwritten.
	require.NoError(t, os.WriteFile(output, []byte("existing backup"), 0o600))
	_, err = o.publish()
	require.Error(t, err)
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, "existing backup", string(data))
	_, err = newReadMemoryOptions(req)
	require.Error(t, err)
	// Empty output cannot masquerade as a successful backup.
	require.NoError(t, os.WriteFile(o.toolOutput, nil, 0o600))
	_, err = o.publish()
	require.ErrorContains(t, err, "non-empty regular file")
}

func TestReadMemoryRejectsInvalidAndUnsupportedRequests(t *testing.T) {
	for _, memory := range []string{"", "../flash", "flash.hex", "flash\" -e", "flash\n"} {
		_, err := newReadMemoryOptions(&rpc.ReadMemoryRequest{Memory: memory, Format: "bin", OutputFile: filepath.Join(t.TempDir(), "out")})
		require.Error(t, err)
	}
	for _, output := range []string{"", "-", "a\"b", "a\nb"} {
		_, err := newReadMemoryOptions(&rpc.ReadMemoryRequest{Memory: "flash", Format: "bin", OutputFile: output})
		require.Error(t, err)
	}
	root := t.TempDir()
	o, err := newReadMemoryOptions(&rpc.ReadMemoryRequest{Memory: "flash", Format: "bin", OutputFile: filepath.Join(root, "out")})
	require.NoError(t, err)
	_, err = o.prepareRecipe("read", properties.NewMap())
	require.ErrorContains(t, err, "not supported")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	props := properties.NewMap()
	props.Set("read.flash.bin.pattern", "must-not-run")
	require.ErrorIs(t, runTool(ctx, "read.flash.bin.pattern", props, io.Discard, io.Discard, false, false, nil), context.Canceled)
}

func TestReadMemoryStreamAdapter(t *testing.T) {
	var stdout, stderr bytes.Buffer
	stream, result := ReadMemoryToServerStreams(context.Background(), &stdout, &stderr)
	require.NoError(t, stream.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_OutStream{OutStream: []byte("out")}}))
	require.NoError(t, stream.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_ErrStream{ErrStream: []byte("err")}}))
	want := &rpc.ReadMemoryResult{OutputFile: "backup.bin", Size: 42}
	require.NoError(t, stream.Send(&rpc.ReadMemoryResponse{Message: &rpc.ReadMemoryResponse_Result{Result: want}}))
	require.Equal(t, "out", stdout.String())
	require.Equal(t, "err", stderr.String())
	require.Equal(t, want, result())
}
