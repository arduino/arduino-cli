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
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/arduino/arduino-cli/arduino/builder/cpp"
	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestCompileOfProblematicSketches(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	// Install Libraries required for tests
	_, _, err = cli.Run("lib", "install", "Bridge@1.6.1")
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"SketchWithInlineFunction", tryBuildAvrLeonardo},
		{"SketchWithConst", tryBuildAvrLeonardo},
		{"SketchWithFunctionSignatureInsideIfdef", tryBuildAvrLeonardo},
		{"SketchWithOldLibrary", tryBuildAvrLeonardo},
		{"SketchWithoutFunctions", tryPreprocessAvrLeonardo},
		{"SketchWithConfig", testBuilderSketchWithConfig},
		{"SketchWithUsbcon", tryBuildAvrLeonardo},
		//{"SketchWithTypename", tryBuildAvrLeonardo}, // XXX: Failing sketch, typename not supported
		{"SketchWithMacosxGarbage", tryBuildAvrLeonardo},
		{"SketchWithNamespace", tryBuildAvrLeonardo},
		{"SketchWithDefaultArgs", tryBuildAvrLeonardo},
		{"SketchWithClass", tryBuildAvrLeonardo},
		{"SketchWithBackupFiles", tryBuildAvrLeonardo},
		{"SketchWithSubfolders", tryBuildAvrLeonardo},
	}.Run(t, env, cli)
}

func testBuilderSketchWithConfig(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	// Compile
	buildPath := tryBuildAvrLeonardoAndGetBuildPath(t, env, cli)
	require.True(t, buildPath.Join("core", "HardwareSerial.cpp.o").Exist())
	require.True(t, buildPath.Join("sketch", "SketchWithConfig.ino.cpp.o").Exist())
	require.True(t, buildPath.Join("SketchWithConfig.ino.elf").Exist())
	require.True(t, buildPath.Join("SketchWithConfig.ino.hex").Exist())
	require.True(t, buildPath.Join("libraries", "Bridge", "Mailbox.cpp.o").Exist())

	// Preprocessing
	sketchPath, preprocessedSketchData := tryPreprocessAvrLeonardoAndGetResult(t, env, cli)
	preprocessedSketch := string(preprocessedSketchData)
	quotedSketchMainFile := cpp.QuoteString(sketchPath.Join(sketchPath.Base() + ".ino").String())
	require.Contains(t, preprocessedSketch, "#include <Arduino.h>\n#line 1 "+quotedSketchMainFile+"\n")
	require.Contains(t, preprocessedSketch, "#line 13 "+quotedSketchMainFile+"\nvoid setup();\n#line 17 "+quotedSketchMainFile+"\nvoid loop();\n#line 13 "+quotedSketchMainFile+"\n")
	preprocessed := loadPreprocessedGoldenFileAndInterpolate(t, sketchPath)
	require.Equal(t, preprocessed, strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}

func tryBuildAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_ = tryBuildAvrLeonardoAndGetBuildPath(t, env, cli)
}

func tryBuildAvrLeonardoAndGetBuildPath(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) *paths.Path {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	libsPath, err := paths.New("testdata", "libraries").Abs()
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("compile", "-b", "arduino:avr:leonardo", "--libraries", libsPath.String(), "--format", "json", sketchPath.String())
	require.NoError(t, err)
	var builderOutput struct {
		BuilderResult struct {
			BuildPath *paths.Path `json:"build_path"`
		} `json:"builder_result"`
	}
	require.NotNil(t, builderOutput.BuilderResult.BuildPath)
	require.NoError(t, json.Unmarshal(jsonOut, &builderOutput))
	return builderOutput.BuilderResult.BuildPath
}

func tryPreprocessAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, _ = tryPreprocessAvrLeonardoAndGetResult(t, env, cli)
}

func tryPreprocessAvrLeonardoAndGetResult(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) (*paths.Path, []byte) {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	out, _, err := cli.Run("compile", "-b", "arduino:avr:leonardo", "--preprocess", sketchPath.String())
	require.NoError(t, err)
	return sketchPath, out
}

func loadPreprocessedGoldenFileAndInterpolate(t *testing.T, sketchDir *paths.Path) string {
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

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	require.NoError(t, err)

	return buf.String()
}
