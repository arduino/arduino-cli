// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package test

import (
	"strings"
	"testing"

	bldr "github.com/arduino/arduino-cli/arduino/builder"
	"github.com/arduino/arduino-cli/arduino/builder/preprocessor"
	"github.com/arduino/arduino-cli/legacy/builder"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func ctagsRunnerTestTemplate(t *testing.T, sketchLocation *paths.Path) []byte {
	ctx := prepareBuilderTestContext(t, nil, sketchLocation, "arduino:avr:leonardo")
	defer cleanUpBuilderTestContext(t, ctx)
	ctx.Verbose = true

	err := (&builder.ContainerSetupHardwareToolsLibsSketchAndProps{}).Run(ctx)
	NoError(t, err)

	_, err = bldr.PrepareSketchBuildPath(ctx.Sketch, nil, ctx.SketchBuildPath)
	NoError(t, err)

	source := loadPreprocessedSketch(t, ctx)
	target := ctx.BuildPath.Join("ctags_target.cpp")
	NoError(t, target.WriteFile([]byte(source)))

	ctagsOutput, _, err := preprocessor.RunCTags(target, ctx.BuildProperties)
	NoError(t, err)

	return ctagsOutput
}

func TestCTagsRunner(t *testing.T) {
	sketchLocation := Abs(t, paths.New("downloaded_libraries", "Bridge", "examples", "Bridge", "Bridge.ino"))
	ctagsOutput := ctagsRunnerTestTemplate(t, sketchLocation)

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "server	" + quotedSketchLocation + "	/^BridgeServer server;$/;\"	kind:variable	line:31\n" +
		"setup	" + quotedSketchLocation + "	/^void setup() {$/;\"	kind:function	line:33	signature:()	returntype:void\n" +
		"loop	" + quotedSketchLocation + "	/^void loop() {$/;\"	kind:function	line:46	signature:()	returntype:void\n" +
		"process	" + quotedSketchLocation + "	/^void process(BridgeClient client) {$/;\"	kind:function	line:62	signature:(BridgeClient client)	returntype:void\n" +
		"digitalCommand	" + quotedSketchLocation + "	/^void digitalCommand(BridgeClient client) {$/;\"	kind:function	line:82	signature:(BridgeClient client)	returntype:void\n" +
		"analogCommand	" + quotedSketchLocation + "	/^void analogCommand(BridgeClient client) {$/;\"	kind:function	line:109	signature:(BridgeClient client)	returntype:void\n" +
		"modeCommand	" + quotedSketchLocation + "	/^void modeCommand(BridgeClient client) {$/;\"	kind:function	line:149	signature:(BridgeClient client)	returntype:void\n"
	require.Equal(t, expectedOutput, strings.Replace(string(ctagsOutput), "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithNamespace(t *testing.T) {
	sketchLocation := Abs(t, paths.New("sketch_with_namespace", "sketch_with_namespace.ino"))
	ctagsOutput := ctagsRunnerTestTemplate(t, sketchLocation)

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "value\t" + quotedSketchLocation + "\t/^\tint value() {$/;\"\tkind:function\tline:2\tnamespace:Test\tsignature:()\treturntype:int\n" +
		"setup\t" + quotedSketchLocation + "\t/^void setup() {}$/;\"\tkind:function\tline:7\tsignature:()\treturntype:void\n" +
		"loop\t" + quotedSketchLocation + "\t/^void loop() {}$/;\"\tkind:function\tline:8\tsignature:()\treturntype:void\n"
	require.Equal(t, expectedOutput, strings.Replace(string(ctagsOutput), "\r\n", "\n", -1))
}

func TestCTagsRunnerSketchWithTemplates(t *testing.T) {
	sketchLocation := Abs(t, paths.New("sketch_with_templates_and_shift", "sketch_with_templates_and_shift.ino"))
	ctagsOutput := ctagsRunnerTestTemplate(t, sketchLocation)

	quotedSketchLocation := strings.Replace(sketchLocation.String(), "\\", "\\\\", -1)
	expectedOutput := "printGyro\t" + quotedSketchLocation + "\t/^void printGyro()$/;\"\tkind:function\tline:10\tsignature:()\treturntype:void\n" +
		"bVar\t" + quotedSketchLocation + "\t/^c< 8 > bVar;$/;\"\tkind:variable\tline:15\n" +
		"aVar\t" + quotedSketchLocation + "\t/^c< 1<<8 > aVar;$/;\"\tkind:variable\tline:16\n" +
		"func\t" + quotedSketchLocation + "\t/^template<int X> func( c< 1<<X> & aParam) {$/;\"\tkind:function\tline:18\tsignature:( c< 1<<X> & aParam)\treturntype:template\n"
	require.Equal(t, expectedOutput, strings.Replace(string(ctagsOutput), "\r\n", "\n", -1))
}
