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
	"golang.org/x/exp/slices"
)

func TestCompileOfProblematicSketches(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
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

	// Install Libraries required for tests
	_, _, err = cli.Run("lib", "install", "Bridge@1.6.1")
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

func tryBuildAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, err := tryBuild(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
}

type builderOutput struct {
	BuilderResult struct {
		BuildPath     *paths.Path       `json:"build_path"`
		UsedLibraries []*builderLibrary `json:"used_libraries"`
	} `json:"builder_result"`
}

type builderLibrary struct {
	Name string `json:"Name"`
}

func tryBuild(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI, fqbn string) (*builderOutput, error) {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	libsPath, err := paths.New("testdata", "libraries").Abs()
	require.NoError(t, err)
	jsonOut, _, err := cli.Run("compile", "-b", fqbn, "--libraries", libsPath.String(), "--clean", "--format", "json", sketchPath.String())
	//require.NoError(t, err)
	var out builderOutput
	require.NoError(t, json.Unmarshal(jsonOut, &out))
	return &out, err
}

func tryPreprocessAvrLeonardo(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	_, _, err := tryPreprocess(t, env, cli, "arduino:avr:leonardo")
	require.NoError(t, err)
}

func tryPreprocess(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI, fqbn string) (*paths.Path, []byte, error) {
	subTestName := filepath.Base(t.Name())
	sketchPath, err := paths.New("testdata", subTestName).Abs()
	require.NoError(t, err)
	out, _, err := cli.Run("compile", "-b", fqbn, "--preprocess", sketchPath.String())
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

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	require.NoError(t, err)

	require.Equal(t, buf.String(), strings.Replace(preprocessedSketch, "\r\n", "\n", -1))
}
