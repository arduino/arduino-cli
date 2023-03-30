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

package ctags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/types"

	"github.com/stretchr/testify/require"
)

func produceTags(t *testing.T, filename string) []*types.CTag {
	bytes, err := os.ReadFile(filepath.Join("test_data", filename))
	require.NoError(t, err)

	parser := CTagsParser{}
	return parser.Parse(bytes, nil)
}

func TestCTagsParserShouldListPrototypes(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldListPrototypes.txt")

	require.Equal(t, 8, len(tags))
	idx := 0
	require.Equal(t, "server", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "process", tags[idx].FunctionName)
	require.Equal(t, "prototype", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "process", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "digitalCommand", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "analogCommand", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
	idx++
	require.Equal(t, "modeCommand", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", tags[idx].Filename)
}

func TestCTagsParserShouldListTemplates(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldListTemplates.txt")

	require.Equal(t, 3, len(tags))
	idx := 0
	require.Equal(t, "minimum", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "template <typename T> T minimum (T a, T b);", tags[idx].Prototype)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "void setup();", tags[idx].Prototype)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "void loop();", tags[idx].Prototype)
}

func TestCTagsParserShouldListTemplates2(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldListTemplates2.txt")

	require.Equal(t, 4, len(tags))
	idx := 0
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "SRAM_writeAnything", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "template <class T> int SRAM_writeAnything(int ee, const T& value);", tags[idx].Prototype)
	idx++
	require.Equal(t, "SRAM_readAnything", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "template <class T> int SRAM_readAnything(int ee, T& value);", tags[idx].Prototype)
}

func TestCTagsParserShouldDealWithClasses(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldDealWithClasses.txt")

	require.Equal(t, 2, len(tags))
	idx := 0
	require.Equal(t, "SleepCycle", tags[idx].FunctionName)
	require.Equal(t, "prototype", tags[idx].Kind)
	idx++
	require.Equal(t, "SleepCycle", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserShouldDealWithStructs(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldDealWithStructs.txt")

	require.Equal(t, 5, len(tags))
	idx := 0
	require.Equal(t, "A_NEW_TYPE", tags[idx].FunctionName)
	require.Equal(t, "struct", tags[idx].Kind)
	idx++
	require.Equal(t, "foo", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	require.Equal(t, "struct:A_NEW_TYPE", tags[idx].Typeref)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "dostuff", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserShouldDealWithMacros(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldDealWithMacros.txt")

	require.Equal(t, 8, len(tags))
	idx := 0
	require.Equal(t, "DEBUG", tags[idx].FunctionName)
	require.Equal(t, "macro", tags[idx].Kind)
	idx++
	require.Equal(t, "DISABLED", tags[idx].FunctionName)
	require.Equal(t, "macro", tags[idx].Kind)
	idx++
	require.Equal(t, "hello", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "debug", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "disabledIsDefined", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "useMyType", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserShouldDealFunctionWithDifferentSignatures(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserShouldDealFunctionWithDifferentSignatures.txt")

	require.Equal(t, 3, len(tags))
	idx := 0
	require.Equal(t, "getBytes", tags[idx].FunctionName)
	require.Equal(t, "prototype", tags[idx].Kind)
	idx++
	require.Equal(t, "getBytes", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "getBytes", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserClassMembersAreFilteredOut(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserClassMembersAreFilteredOut.txt")

	require.Equal(t, 5, len(tags))
	idx := 0
	require.Equal(t, "set_values", tags[idx].FunctionName)
	require.Equal(t, "prototype", tags[idx].Kind)
	require.Equal(t, "Rectangle", tags[idx].Class)
	idx++
	require.Equal(t, "area", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "Rectangle", tags[idx].Class)
	idx++
	require.Equal(t, "set_values", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "Rectangle", tags[idx].Class)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserStructWithFunctions(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserStructWithFunctions.txt")

	require.Equal(t, 8, len(tags))
	idx := 0
	require.Equal(t, "sensorData", tags[idx].FunctionName)
	require.Equal(t, "struct", tags[idx].Kind)
	idx++
	require.Equal(t, "sensorData", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "sensorData", tags[idx].Struct)
	idx++
	require.Equal(t, "sensorData", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "sensorData", tags[idx].Struct)
	idx++
	require.Equal(t, "sensors", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	idx++
	require.Equal(t, "sensor1", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	idx++
	require.Equal(t, "sensor2", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserDefaultArguments(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserDefaultArguments.txt")

	require.Equal(t, 3, len(tags))
	idx := 0
	require.Equal(t, "test", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "void test(int x = 1);", tags[idx].Prototype)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserNamespace(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserNamespace.txt")

	require.Equal(t, 3, len(tags))
	idx := 0
	require.Equal(t, "value", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "Test", tags[idx].Namespace)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserStatic(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserStatic.txt")

	require.Equal(t, 3, len(tags))
	idx := 0
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "doStuff", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserFunctionPointer(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserFunctionPointer.txt")

	require.Equal(t, 4, len(tags))
	idx := 0
	require.Equal(t, "t1Callback", tags[idx].FunctionName)
	require.Equal(t, "variable", tags[idx].Kind)
	idx++
	require.Equal(t, "t1Callback", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
}

func TestCTagsParserFunctionPointers(t *testing.T) {
	tags := produceTags(t, "TestCTagsParserFunctionPointers.txt")

	require.Equal(t, 5, len(tags))
	idx := 0
	require.Equal(t, "setup", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "loop", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "func", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	idx++
	require.Equal(t, "funcArr", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "int funcArr();", tags[idx].Prototype)
	idx++
	require.Equal(t, "funcCombo", tags[idx].FunctionName)
	require.Equal(t, "function", tags[idx].Kind)
	require.Equal(t, "void funcCombo(void (*(&in)[5])(int));", tags[idx].Prototype)
}
