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

package compile_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
	"text/template"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestCompileOfProblematicSketches(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:sam@1.6.12")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:samd@1.8.13")
	require.NoError(t, err)

	// Install REDBear Lad platform
	_, _, err = cli.Run("config", "init")
	require.NoError(t, err)
	_, _, err = cli.Run("config", "add", "board_manager.additional_urls", "https://redbearlab.github.io/arduino/package_redbearlab_index.json")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "RedBear:avr@1.0.0")
	require.NoError(t, err)

	// XXX: This compiler gives an error "sorry - this program has been built without plugin support"
	//      for avr-gcc/4.8.1-arduino5/bin/avr-gcc-ar. Maybe it's a problem of the very old avr-gcc...
	//      Removing it will enforce the builder to use the more recent avr-gcc from arduino:avr platform.
	require.NoError(t, cli.DataDir().Join("packages", "arduino", "tools", "avr-gcc", "4.8.1-arduino5").RemoveAll())

	// Install Libraries required for tests
	_, _, err = cli.Run("lib", "install", "Bridge@1.6.1")
	require.NoError(t, err)
	_, _, err = cli.Run("lib", "install", "CapacitiveSensor@0.5")
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"SketchWithInlineFunction", testBuilderSketchWithInlineFunction},
		{"SketchWithConst", testBuilderSketchWithConst},
		{"SketchWithFunctionSignatureInsideIfdef", testBuilderSketchWithFunctionSignatureInsideIfdef},
		{"SketchWithOldLibrary", testBuilderSketchWithOldLibrary},
		{"SketchWithoutFunctions", testBuilderSketchWithoutFunctions},
		{"SketchWithConfig", testBuilderSketchWithConfig},
		{"SketchWithUsbcon", testBuilderSketchWithUsbcon},
		{"SketchWithTypename", testBuilderSketchWithTypename},
		{"SketchWithMacosxGarbage", tryBuildAvrLeonardo},
		{"SketchWithNamespace", testBuilderSketchWithNamespace},
		{"SketchWithDefaultArgs", testBuilderSketchWithDefaultArgs},
		{"SketchWithClass", testBuilderSketchWithClass},
		{"SketchWithBackupFiles", testBuilderSketchWithBackupFiles},
		{"SketchWithSubfolders", testBuilderSketchWithSubfolders},
		{"SketchWithTemplatesAndShift", testBuilderSketchWithTemplatesAndShift},
		{"SketchRequiringEOLProcessing", tryBuildAvrLeonardo},
		{"SketchWithIfDef", testBuilderSketchWithIfDef},
		{"SketchWithIfDef2", testBuilderSketchWithIfDef2},
		{"SketchWithIfDef3", testBuilderSketchWithIfDef3},
		{"BridgeExample", testBuilderBridgeExample},
		{"Baladuino", testBuilderBaladuino},
		{"SketchWithEscapedDoubleQuote", testBuilderSketchWithEscapedDoubleQuote},
		{"SketchWithIncludeBetweenMultilineComment", testBuilderSketchWithIncludeBetweenMultilineComment},
		{"SketchWithLineContinuations", testBuilderSketchWithLineContinuations},
		{"SketchWithStringWithComment", testBuilderSketchWithStringWithComment},
		{"SketchWithStruct", testBuilderSketchWithStruct},
		{"SketchNoFunctionsTwoFiles", testBuilderSketchNoFunctionsTwoFiles},
		{"SketchWithClassAndMethodSubstring", testBuilderSketchWithClassAndMethodSubstring},
		{"SketchThatChecksIfSPIHasTransactions", tryBuildAvrLeonardo},
		{"SketchWithDependendLibraries", tryBuildAvrLeonardo},
		{"SketchWithFunctionPointer", tryBuildAvrLeonardo},
		{"USBHostExample", testBuilderUSBHostExample},
		{"SketchWithConflictingLibraries", testBuilderSketchWithConflictingLibraries},
		{"SketchLibraryProvidesAllIncludes", testBuilderSketchLibraryProvidesAllIncludes},
	}.Run(t, env, cli)
}

func testBuilderSketchWithConfig(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Compile
	out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	buildPath := out.BuilderResult.BuildPath
	require.NotNil(t, buildPath)
	require.True(t, buildPath.Join("core", "HardwareSerial.cpp.o").Exist())
	require.True(t, buildPath.Join("sketch", "SketchWithConfig.ino.cpp.o").Exist())
	require.True(t, buildPath.Join("SketchWithConfig.ino.elf").Exist())
	require.True(t, buildPath.Join("SketchWithConfig.ino.hex").Exist())
	require.True(t, buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())

	// Preprocessing
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithoutFunctions(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.Error(t, err)
	_, err = tryBuild(t, env, cli, "RedBear:avr:blend")
	require.Error(t, err)

	// Preprocess
	sketchPath, preprocessedSketchData, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	preprocessedSketch := string(preprocessedSketchData)
	quotedSketchMainFile := cpp.QuoteString(sketchPath.Join(sketchPath.Base() + ".ino").String())
	require.Contains(t, preprocessedSketch, "#include <Arduino.h>\n#line 1 "+quotedSketchMainFile+"\n")
}

func testBuilderSketchWithBackupFiles(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	_, err = tryBuild(t, env, cli, "arduino:avr:uno")
	require.NoError(t, err)
}

func testBuilderSketchWithOldLibrary(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	_, err = tryBuild(t, env, cli, "arduino:avr:uno")
	require.NoError(t, err)
}

func testBuilderSketchWithSubfolders(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	importedLibraries := out.BuilderResult.UsedLibraries
	slices.SortFunc(importedLibraries, func(x, y *builderLibrary) bool { return x.Name < y.Name })
	require.NoError(t, err)
	require.Equal(t, 3, len(importedLibraries))
	require.Equal(t, "testlib1", importedLibraries[0].Name)
	require.Equal(t, "testlib2", importedLibraries[1].Name)
	require.Equal(t, "testlib3", importedLibraries[2].Name)

	_, err = tryBuild(t, env, cli, "arduino:avr:uno")
	require.NoError(t, err)
}

func testBuilderSketchWithClass(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	_, err = tryBuild(t, env, cli, "arduino:avr:uno")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:uno")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithTypename(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// XXX: Failing sketch, typename not supported.
	//      This test will be skipped until a better C++ parser is adopted
	t.SkipNow()

	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithNamespace(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithDefaultArgs(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithInlineFunction(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithFunctionSignatureInsideIfdef(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithUsbcon(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithConst(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithTemplatesAndShift(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// XXX: Failing sketch, template with shift not supported.
	//      This test will be skipped until a better C++ parser is adopted
	t.SkipNow()

	// Build
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithIfDef(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	output, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	require.Empty(t, output.BuilderResult.UsedLibraries)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithIfDef2(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	output, err := tryBuild(t, env, cli, "arduino:avr:yun")
	require.NoError(t, err)
	require.Empty(t, output.BuilderResult.UsedLibraries)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:yun")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderSketchWithIfDef3(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Build
	output, err := tryBuild(t, env, cli, "arduino:sam:arduino_due_x_dbg")
	require.NoError(t, err)
	require.Empty(t, output.BuilderResult.UsedLibraries)

	// Preprocess
	sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:sam:arduino_due_x_dbg")
	require.NoError(t, err)
	comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
}

func testBuilderBridgeExample(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("BuildForAVR", func(t *testing.T) {
		// Build
		out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)

		buildPath := out.BuilderResult.BuildPath
		require.True(t, buildPath.Join("core", "HardwareSerial.cpp.o").Exist())
		require.True(t, buildPath.Join("sketch", "BridgeExample.ino.cpp.o").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.elf").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.hex").Exist())
		require.True(t, buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())

		libs := out.BuilderResult.UsedLibraries
		require.Len(t, libs, 1)
		require.Equal(t, "Bridge", libs[0].Name)

		// Build again...
		out2, err2 := tryBuild(t, env, cli, "arduino:avr:leonardo", "no-clean")
		require.NoError(t, err2)
		buildPath2 := out2.BuilderResult.BuildPath
		require.True(t, buildPath2.Join("core", "HardwareSerial.cpp.o").Exist())
		require.True(t, buildPath2.Join("sketch", "BridgeExample.ino.cpp.o").Exist())
		require.True(t, buildPath2.Join("BridgeExample.ino.elf").Exist())
		require.True(t, buildPath2.Join("BridgeExample.ino.hex").Exist())
		require.True(t, buildPath2.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())
	})

	t.Run("BuildForSAM", func(t *testing.T) {
		// Build again for SAM...
		out, err := tryBuild(t, env, cli, "arduino:sam:arduino_due_x_dbg", "all-warnings")
		require.NoError(t, err)

		buildPath := out.BuilderResult.BuildPath
		require.True(t, buildPath.Join("core", "syscalls_sam3.c.o").Exist())
		require.True(t, buildPath.Join("core", "USB", "PluggableUSB.cpp.o").Exist())
		require.True(t, buildPath.Join("core", "avr", "dtostrf.c.d").Exist())
		require.True(t, buildPath.Join("sketch", "BridgeExample.ino.cpp.o").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.elf").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.bin").Exist())
		require.True(t, buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())

		objdump := cli.DataDir().Join("packages", "arduino", "tools", "arm-none-eabi-gcc", "4.8.3-2014q1", "bin", "arm-none-eabi-objdump")
		cmd := exec.Command(
			objdump.String(),
			"-f", buildPath.Join("core", "core.a").String())
		bytes, err := cmd.CombinedOutput()
		require.NoError(t, err)
		require.NotContains(t, string(bytes), "variant.cpp.o")
	})

	t.Run("BuildForRedBearAVR", func(t *testing.T) {
		// Build again for RedBearLab...
		out, err := tryBuild(t, env, cli, "RedBear:avr:blend", "verbose")
		require.NoError(t, err)
		buildPath := out.BuilderResult.BuildPath
		require.True(t, buildPath.Join("core", "HardwareSerial.cpp.o").Exist())
		require.True(t, buildPath.Join("sketch", "BridgeExample.ino.cpp.o").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.elf").Exist())
		require.True(t, buildPath.Join("BridgeExample.ino.hex").Exist())
		require.True(t, buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())
	})

	t.Run("BuildPathMustNotContainsUnusedPreviouslyCompiledLibrary", func(t *testing.T) {
		out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		buildPath := out.BuilderResult.BuildPath

		// Simulate a library use in libraries build path
		require.NoError(t, buildPath.Join("libraries", "SPI").MkdirAll())

		// Build again...
		_, err = tryBuild(t, env, cli, "arduino:avr:leonardo", "no-clean")
		require.NoError(t, err)

		require.False(t, buildPath.Join("libraries", "SPI").Exist())
		require.True(t, buildPath.Join("libraries", "Bridge").Exist())
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderBaladuino(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		t.Skip("The sketch is missing required libraries to build")

		// Build
		output, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		require.Empty(t, output.BuilderResult.UsedLibraries)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithEscapedDoubleQuote(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithIncludeBetweenMultilineComment(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithLineContinuations(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithStringWithComment(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithStruct(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchNoFunctionsTwoFiles(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.Error(t, err)
		require.Contains(t, out.CompilerErr, "undefined reference to `loop'")
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderSketchWithClassAndMethodSubstring(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		_, err := tryBuild(t, env, cli, "arduino:avr:uno")
		require.NoError(t, err)
	})

	t.Run("Preprocess", func(t *testing.T) {
		// Preprocess
		sketchPath, preprocessedSketch, err := tryPreprocess(t, env, cli, "arduino:avr:uno")
		require.NoError(t, err)
		comparePreprocessGoldenFile(t, sketchPath, preprocessedSketch)
	})
}

func testBuilderUSBHostExample(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		out, err := tryBuild(t, env, cli, "arduino:samd:arduino_zero_native")
		require.NoError(t, err)

		libs := out.BuilderResult.UsedLibraries
		require.Len(t, libs, 1)
		require.Equal(t, "USBHost", libs[0].Name)
		usbHostLibDir, err := paths.New("testdata", "libraries", "USBHost", "src").Abs()
		require.NoError(t, err)
		require.True(t, libs[0].SourceDir.EquivalentTo(usbHostLibDir))
	})
}

func testBuilderSketchWithConflictingLibraries(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// This library has a conflicting IRremote.h
		_, _, err := cli.Run("lib", "install", "Robot IR Remote@1.0.2")
		require.NoError(t, err)

		// Build
		out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		libs := out.BuilderResult.UsedLibraries
		slices.SortFunc(libs, func(x, y *builderLibrary) bool { return x.Name < y.Name })
		require.Len(t, libs, 2)
		require.Equal(t, "Bridge", libs[0].Name)
		require.Equal(t, "IRremote", libs[1].Name)
	})
}

func testBuilderSketchLibraryProvidesAllIncludes(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	t.Run("Build", func(t *testing.T) {
		// Build
		out, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
		require.NoError(t, err)
		libs := out.BuilderResult.UsedLibraries
		slices.SortFunc(libs, func(x, y *builderLibrary) bool { return x.Name < y.Name })
		require.Len(t, libs, 2)
		require.Equal(t, "ANewLibrary-master", libs[0].Name)
		require.Equal(t, "IRremote", libs[1].Name)
	})
}

func tryBuildAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
}

type builderOutput struct {
	CompilerOut   string `json:"compiler_out"`
	CompilerErr   string `json:"compiler_err"`
	BuilderResult struct {
		BuildPath     *paths.Path       `json:"build_path"`
		UsedLibraries []*builderLibrary `json:"used_libraries"`
	} `json:"builder_result"`
}

type builderLibrary struct {
	Name       string      `json:"Name"`
	InstallDir *paths.Path `json:"install_dir"`
	SourceDir  *paths.Path `json:"source_dir"`
}

func tryBuild(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI, fqbn string, options ...string) (*builderOutput, error) {
	subTestName := strings.Split(t.Name(), "/")[1]
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	libsPath, err := paths.New("testdata", "libraries").Abs()
	require.NoError(t, err)
	args := []string{
		"compile",
		"-b", fqbn,
		"--libraries", libsPath.String(),
		"--format", "json",
		sketchPath.String()}
	if !slices.Contains(options, "no-clean") {
		args = append(args, "--clean")
	}
	if slices.Contains(options, "all-warnings") {
		args = append(args, "--warnings", "all")
	}
	if slices.Contains(options, "verbose") {
		args = append(args, "-v")
	}
	jsonOut, _, err := cli.Run(args...)
	var out builderOutput
	require.NoError(t, json.Unmarshal(jsonOut, &out))
	return &out, err
}

func tryPreprocess(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI, fqbn string) (*paths.Path, []byte, error) {
	subTestName := strings.Split(t.Name(), "/")[1]
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	libsPath, err := paths.New("testdata", "libraries").Abs()
	require.NoError(t, err)
	out, _, err := cli.Run("compile", "-b", fqbn, "--preprocess", "--libraries", libsPath.String(), sketchPath.String())
	return sketchPath, out, err
}

func comparePreprocessGoldenFile(t *testing.T, sketchDir *paths.Path, preprocessedSketchData []byte) {
	preprocessedSketch := string(preprocessedSketchData)

	sketchName := sketchDir.Base()
	sketchMainFile := sketchDir.Join(sketchName + ".ino")
	sketchTemplate := sketchDir.Join(sketchName + ".preprocessed.txt")

	funcsMap := template.FuncMap{
		"QuoteCppString": func(p *paths.Path) string { return cpp.QuoteString(p.String()) },
	}
	tpl, err := template.New(sketchTemplate.Base()).
		Funcs(funcsMap).
		ParseFiles(sketchTemplate.String())
	require.NoError(t, err)

	data := make(map[string]interface{})
	data["sketchMainFile"] = sketchMainFile
	data["sketchDir"] = sketchDir
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	require.NoError(t, err)

	require.Equal(t, buf.String(), strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}
