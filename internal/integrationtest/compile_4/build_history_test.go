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

package compile_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

const (
	historyHitMsg   = "Using a previous identical build from the build history, compilation skipped."
	historyMissMsg  = "No matching build found in the build history."
	historySavedMsg = "Build saved in the build history."
)

type historyTestEnv struct {
	t          *testing.T
	cli        *integrationtest.ArduinoCLI
	env        map[string]string
	cacheDir   *paths.Path
	historyDir *paths.Path
}

// newHistoryTestEnv creates a test environment with arduino:avr installed and
// the build history enabled (if maxEntries > 0) in a private build cache.
func newHistoryTestEnv(t *testing.T, maxEntries int) *historyTestEnv {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	cacheDir := env.RootDir().Join("build-cache")
	cliEnv := cli.GetDefaultEnv()
	cliEnv["ARDUINO_BUILD_CACHE_PATH"] = cacheDir.String()
	if maxEntries > 0 {
		cliEnv["ARDUINO_BUILD_CACHE_HISTORY_ENABLED"] = "true"
		cliEnv["ARDUINO_BUILD_CACHE_HISTORY_MAX_ENTRIES"] = fmt.Sprint(maxEntries)
	}
	return &historyTestEnv{
		t:          t,
		cli:        cli,
		env:        cliEnv,
		cacheDir:   cacheDir,
		historyDir: cacheDir.Join("history"),
	}
}

// compile runs a verbose compilation and returns the output.
func (h *historyTestEnv) compile(args ...string) (string, error) {
	stdout, stderr, err := h.cli.RunWithCustomEnv(h.env, append([]string{"compile", "-v"}, args...)...)
	return string(stdout) + string(stderr), err
}

func (h *historyTestEnv) requireMiss(args ...string) string {
	out, err := h.compile(args...)
	require.NoError(h.t, err, out)
	require.Contains(h.t, out, historyMissMsg)
	require.Contains(h.t, out, "Compiling sketch...")
	require.NotContains(h.t, out, historyHitMsg)
	return out
}

func (h *historyTestEnv) requireHit(args ...string) string {
	out, err := h.compile(args...)
	require.NoError(h.t, err, out)
	require.Contains(h.t, out, historyHitMsg)
	require.NotContains(h.t, out, "Compiling sketch...")
	require.NotContains(h.t, out, "Linking everything together...")
	require.Contains(h.t, out, "Sketch uses", "the program size must be reported also for restored builds")
	return out
}

// entries returns the number of builds stored in the history.
func (h *historyTestEnv) entries() int {
	dirs, err := h.historyDir.ReadDir()
	if err != nil {
		return 0
	}
	dirs.FilterDirs()
	dirs.Filter(func(p *paths.Path) bool {
		if strings.HasPrefix(p.Base(), ".") {
			return false
		}
		data, err := p.Join("metadata.json").ReadFile()
		var md map[string]any
		return err == nil && json.Unmarshal(data, &md) == nil
	})
	return len(dirs)
}

func (h *historyTestEnv) buildPath(args ...string) *paths.Path {
	stdout, _, err := h.cli.RunWithCustomEnv(h.env, append([]string{"compile", "--format", "json"}, args...)...)
	require.NoError(h.t, err)
	var res struct {
		BuilderResult struct {
			BuildPath string `json:"build_path"`
		} `json:"builder_result"`
	}
	require.NoError(h.t, json.Unmarshal(stdout, &res))
	require.NotEmpty(h.t, res.BuilderResult.BuildPath)
	return paths.New(res.BuilderResult.BuildPath)
}

func writeBlink(t *testing.T, sketch *paths.Path, delay string) {
	require.NoError(t, sketch.MkdirAll())
	src := "void setup() { pinMode(13, OUTPUT); }\n" +
		"void loop() { digitalWrite(13, HIGH); delay(" + delay + "); digitalWrite(13, LOW); delay(" + delay + "); }\n"
	require.NoError(t, sketch.Join(sketch.Base()+".ino").WriteFile([]byte(src)))
}

func readFile(t *testing.T, p *paths.Path) string {
	data, err := p.ReadFile()
	require.NoError(t, err)
	return string(data)
}

func TestBuildHistoryDisabledByDefault(t *testing.T) {
	h := newHistoryTestEnv(t, 0)
	sketch := h.cli.SketchbookDir().Join("Blink")
	writeBlink(t, sketch, "100")

	for range 2 {
		out, err := h.compile("-b", "arduino:avr:uno", sketch.String())
		require.NoError(t, err, out)
		require.NotContains(t, out, "build history")
	}
	require.False(t, h.historyDir.Exist())

	// The incremental build is unchanged
	out, err := h.compile("-b", "arduino:avr:uno", sketch.String())
	require.NoError(t, err, out)
	require.Contains(t, out, "Using previously compiled file")
	require.Contains(t, out, "Using precompiled core")
}

func TestBuildHistory(t *testing.T) {
	h := newHistoryTestEnv(t, 10)
	sketch := h.cli.SketchbookDir().Join("Blink")
	uno := []string{"-b", "arduino:avr:uno", sketch.String()}

	// First compilation: miss, the build is stored
	writeBlink(t, sketch, "100")
	out := h.requireMiss(uno...)
	require.Contains(t, out, historySavedMsg)
	require.Equal(t, 1, h.entries())
	buildPath := h.buildPath(uno...)
	hexA := readFile(t, buildPath.Join("Blink.ino.hex"))

	// Identical compilation: hit
	h.requireHit(uno...)
	require.Equal(t, 1, h.entries())

	// Versions B and C: miss
	writeBlink(t, sketch, "200")
	h.requireMiss(uno...)
	writeBlink(t, sketch, "300")
	h.requireMiss(uno...)
	hexC := readFile(t, buildPath.Join("Blink.ino.hex"))
	require.NotEqual(t, hexA, hexC)
	require.Equal(t, 3, h.entries())

	// Revert to A: hit, the firmware of A is restored in the build path
	writeBlink(t, sketch, "100")
	h.requireHit(uno...)
	require.Equal(t, hexA, readFile(t, buildPath.Join("Blink.ino.hex")))
	require.Equal(t, 3, h.entries())

	// The restored build can be uploaded with the normal upload workflow
	stdout, stderr, err := h.cli.RunWithCustomEnv(h.env, "upload", "-b", "arduino:avr:uno", "-p", "/dev/ttyACM0", "--dry-run", "-v", sketch.String())
	require.NoError(t, err, string(stderr))
	require.Contains(t, string(stdout), buildPath.Join("Blink.ino.hex").String())

	// The binaries of a restored build can be exported
	exportDir := sketch.Join("exported")
	h.requireHit("-b", "arduino:avr:uno", "--output-dir", exportDir.String(), sketch.String())
	require.Equal(t, hexA, readFile(t, exportDir.Join("Blink.ino.hex")))
	require.NoError(t, exportDir.RemoveAll())

	// Different board: miss
	h.requireMiss("-b", "arduino:avr:nano", sketch.String())
	h.requireHit("-b", "arduino:avr:nano", sketch.String())
	// Different board options: miss
	h.requireMiss("-b", "arduino:avr:nano:cpu=atmega328old", sketch.String())
	h.requireHit("-b", "arduino:avr:nano:cpu=atmega328old", sketch.String())

	// Different compiler flags: miss
	flags := []string{"-b", "arduino:avr:uno", "--build-property", "compiler.cpp.extra_flags=-DEXTRA_FLAG", sketch.String()}
	h.requireMiss(flags...)
	h.requireHit(flags...)
	// arduino:avr does not define compiler.optimization_flags.debug: optimizing
	// for debug produces exactly the same build properties and firmware.
	h.requireHit("-b", "arduino:avr:uno", "--optimize-for-debug", sketch.String())
	h.requireMiss("-b", "arduino:avr:uno", "--warnings", "all", sketch.String())

	// The original configuration is still a hit
	h.requireHit(uno...)
	entries := h.entries()

	// Failed compilations are not stored
	require.NoError(t, sketch.Join("Blink.ino").WriteFile([]byte("void setup() { syntax error }\nvoid loop() {}\n")))
	out, err = h.compile(uno...)
	require.Error(t, err, out)
	require.NotContains(t, out, historySavedMsg)
	require.Equal(t, entries, h.entries())

	// A clean build is always performed and does not store duplicates
	writeBlink(t, sketch, "100")
	out, err = h.compile(append([]string{"--clean"}, uno...)...)
	require.NoError(t, err, out)
	require.NotContains(t, out, historyHitMsg)
	require.Contains(t, out, "Compiling sketch...")
	require.NotContains(t, out, historySavedMsg)
	require.Equal(t, entries, h.entries())
	require.Equal(t, hexA, readFile(t, buildPath.Join("Blink.ino.hex")))

	// After a hit the incremental build keeps working
	h.requireHit(uno...)
	writeBlink(t, sketch, "400")
	out = h.requireMiss(uno...)
	require.Contains(t, out, "Using precompiled core")

	// Partial builds do not use the build history
	out, err = h.compile("-b", "arduino:avr:uno", "--only-compilation-database", sketch.String())
	require.NoError(t, err, out)
	require.NotContains(t, out, "build history")

	// Sketches embedding the build time are never reused
	require.NoError(t, sketch.Join("Blink.ino").WriteFile([]byte(
		"const char *built = __DATE__ \" \" __TIME__;\nvoid setup() { Serial.begin(9600); Serial.print(built); }\nvoid loop() {}\n")))
	before := h.entries()
	for range 2 {
		out, err = h.compile(uno...)
		require.NoError(t, err, out)
		require.Contains(t, out, "Build history not used: the sketch uses the __DATE__ macro")
		require.NotContains(t, out, historyHitMsg)
	}
	require.Equal(t, before, h.entries())
}

func TestBuildHistoryLibraryChanges(t *testing.T) {
	h := newHistoryTestEnv(t, 10)
	sketch := h.cli.SketchbookDir().Join("UsesMyLib")
	require.NoError(t, sketch.MkdirAll())
	require.NoError(t, sketch.Join("UsesMyLib.ino").WriteFile([]byte(
		"#include <MyLib.h>\nvoid setup() { Serial.begin(MYLIB_SPEED); }\nvoid loop() {}\n")))
	lib := h.cli.SketchbookDir().Join("libraries", "MyLib")
	require.NoError(t, lib.Join("src").MkdirAll())
	writeLib := func(version, speed string) {
		require.NoError(t, lib.Join("library.properties").WriteFile([]byte(
			"name=MyLib\nversion="+version+"\nauthor=test\nmaintainer=test\nsentence=test\nparagraph=test\ncategory=Other\nurl=http://example.com\narchitectures=*\n")))
		require.NoError(t, lib.Join("src", "MyLib.h").WriteFile([]byte("#define MYLIB_SPEED "+speed+"\n")))
	}
	args := []string{"-b", "arduino:avr:uno", sketch.String()}

	writeLib("1.0.0", "9600")
	out := h.requireMiss(args...)
	require.Contains(t, out, historySavedMsg)
	out = h.requireHit(args...)
	require.Contains(t, out, "Using library MyLib", "the used libraries must be reported also for restored builds")

	// Changing the content of a library header: miss
	writeLib("1.0.0", "115200")
	h.requireMiss(args...)
	// Reverting the content: hit
	writeLib("1.0.0", "9600")
	h.requireHit(args...)

	// Changing the library version: miss
	writeLib("1.0.1", "9600")
	h.requireMiss(args...)
	h.requireHit(args...)

	// Adding a source file to the library: miss
	require.NoError(t, lib.Join("src", "extra.cpp").WriteFile([]byte("int mylib_extra = 1;\n")))
	h.requireMiss(args...)
	h.requireHit(args...)

	// Installing a library that may change the library resolution: miss
	other := h.cli.SketchbookDir().Join("libraries", "OtherLib")
	require.NoError(t, other.MkdirAll())
	require.NoError(t, other.Join("OtherLib.h").WriteFile([]byte("#define OTHER 1\n")))
	h.requireMiss(args...)
}

func TestBuildHistoryEviction(t *testing.T) {
	h := newHistoryTestEnv(t, 2)
	sketch := h.cli.SketchbookDir().Join("Blink")
	args := []string{"-b", "arduino:avr:uno", sketch.String()}

	writeBlink(t, sketch, "100")
	h.requireMiss(args...)
	writeBlink(t, sketch, "200")
	h.requireMiss(args...)
	require.Equal(t, 2, h.entries())

	// Using "100" makes "200" the least recently used build
	writeBlink(t, sketch, "100")
	h.requireHit(args...)

	writeBlink(t, sketch, "300")
	h.requireMiss(args...)
	require.Equal(t, 2, h.entries())

	writeBlink(t, sketch, "100")
	h.requireHit(args...)
	writeBlink(t, sketch, "300")
	h.requireHit(args...)
	writeBlink(t, sketch, "200")
	h.requireMiss(args...) // evicted
	require.Equal(t, 2, h.entries())
}

func buildPathArgs(h *historyTestEnv, buildPath *paths.Path) []string {
	return []string{"-b", "arduino:avr:uno", "--build-path", buildPath.String(), h.cli.SketchbookDir().Join("Blink").String()}
}

func TestBuildHistoryDisabledWithBuildPath(t *testing.T) {
	{
		// History disabled: --build-path behaves as usual
		h := newHistoryTestEnv(t, 0)
		buildPath := h.cli.SketchbookDir().Join("custom-build")
		writeBlink(t, h.cli.SketchbookDir().Join("Blink"), "100")
		out, err := h.compile(buildPathArgs(h, buildPath)...)
		require.NoError(t, err, out)
		require.True(t, buildPath.Join("Blink.ino.hex").Exist())
		require.False(t, h.historyDir.Exist())
	}
}

func TestBuildHistoryWithBuildPath(t *testing.T) {
	{
		// History enabled: builds in a custom build path are stored and restored in that path
		h := newHistoryTestEnv(t, 10)
		sketch := h.cli.SketchbookDir().Join("Blink")
		buildPath := h.cli.SketchbookDir().Join("custom-build")
		writeBlink(t, sketch, "100")
		h.requireMiss(buildPathArgs(h, buildPath)...)
		hexA := readFile(t, buildPath.Join("Blink.ino.hex"))
		writeBlink(t, sketch, "200")
		h.requireMiss(buildPathArgs(h, buildPath)...)
		writeBlink(t, sketch, "100")
		h.requireHit(buildPathArgs(h, buildPath)...)
		require.Equal(t, hexA, readFile(t, buildPath.Join("Blink.ino.hex")))

		// Builds are specific to their build path
		h.requireMiss("-b", "arduino:avr:uno", sketch.String())

		// A build path inside the sketch folder is not part of the sketch
		inner := sketch.Join("build-inside")
		h.requireMiss("-b", "arduino:avr:uno", "--build-path", inner.String(), sketch.String())
		h.requireHit("-b", "arduino:avr:uno", "--build-path", inner.String(), sketch.String())
	}
}

func TestBuildHistoryConcurrentCompilations(t *testing.T) {
	h := newHistoryTestEnv(t, 10)

	// Leftovers of interrupted processes must be ignored
	require.NoError(t, h.historyDir.Join(".tmp-interrupted", "artifacts").MkdirAll())
	require.NoError(t, h.historyDir.Join(".tmp-interrupted", "metadata.json").WriteFile([]byte(`{"format_version":1`)))
	require.NoError(t, h.historyDir.Join("truncated").MkdirAll())
	require.NoError(t, h.historyDir.Join("truncated", "metadata.json").WriteFile([]byte(`{"format_version":1,"key":`)))

	var sketches []*paths.Path
	for i := range 3 {
		sketch := h.cli.SketchbookDir().Join(fmt.Sprintf("Sketch%d", i))
		writeBlink(t, sketch, fmt.Sprint(100*(i+1)))
		sketches = append(sketches, sketch)
	}

	compileAll := func() []string {
		outs := make([]string, len(sketches))
		errs := make([]error, len(sketches))
		var wg sync.WaitGroup
		for i, sketch := range sketches {
			wg.Add(1)
			go func() {
				defer wg.Done()
				outs[i], errs[i] = h.compile("-b", "arduino:avr:uno", sketch.String())
			}()
		}
		wg.Wait()
		for i := range sketches {
			require.NoError(t, errs[i], outs[i])
		}
		return outs
	}

	for _, out := range compileAll() {
		require.Contains(t, out, historySavedMsg)
	}
	require.Equal(t, 3, h.entries())
	for _, out := range compileAll() {
		require.Contains(t, out, historyHitMsg)
	}
	require.Equal(t, 3, h.entries())
}
