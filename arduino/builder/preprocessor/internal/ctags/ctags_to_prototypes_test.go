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

	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func producePrototypes(t *testing.T, filename string, mainFile string) ([]*Prototype, int) {
	bytes, err := os.ReadFile(filepath.Join("testdata", filename))
	require.NoError(t, err)

	parser := &Parser{}
	return parser.Parse(bytes, paths.New(mainFile))
}

func TestCTagsToPrototypesShouldListPrototypes(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldListPrototypes.txt", "/tmp/sketch7210316334309249705.cpp")
	require.Equal(t, 5, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/sketch7210316334309249705.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "void digitalCommand(YunClient client);", prototypes[2].Prototype)
	require.Equal(t, "void analogCommand(YunClient client);", prototypes[3].Prototype)
	require.Equal(t, "void modeCommand(YunClient client);", prototypes[4].Prototype)

	require.Equal(t, 33, line)
}

func TestCTagsToPrototypesShouldListTemplates(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldListTemplates.txt", "/tmp/sketch8398023134925534708.cpp")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "template <typename T> T minimum (T a, T b);", prototypes[0].Prototype)
	require.Equal(t, "/tmp/sketch8398023134925534708.cpp", prototypes[0].File)
	require.Equal(t, "void setup();", prototypes[1].Prototype)
	require.Equal(t, "void loop();", prototypes[2].Prototype)

	require.Equal(t, 2, line)
}

func TestCTagsToPrototypesShouldListTemplates2(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldListTemplates2.txt", "/tmp/sketch463160524247569568.cpp")

	require.Equal(t, 4, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/sketch463160524247569568.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "template <class T> int SRAM_writeAnything(int ee, const T& value);", prototypes[2].Prototype)
	require.Equal(t, "template <class T> int SRAM_readAnything(int ee, T& value);", prototypes[3].Prototype)

	require.Equal(t, 1, line)
}

func TestCTagsToPrototypesShouldDealWithClasses(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldDealWithClasses.txt", "/tmp/sketch9043227824785312266.cpp")

	require.Equal(t, 0, len(prototypes))

	require.Equal(t, 8, line)
}

func TestCTagsToPrototypesShouldDealWithStructs(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldDealWithStructs.txt", "/tmp/sketch8930345717354294915.cpp")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/sketch8930345717354294915.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "void dostuff(A_NEW_TYPE * bar);", prototypes[2].Prototype)

	require.Equal(t, 9, line)
}

func TestCTagsToPrototypesShouldDealWithMacros(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldDealWithMacros.txt", "/tmp/sketch5976699731718729500.cpp")

	require.Equal(t, 5, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/sketch5976699731718729500.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "void debug();", prototypes[2].Prototype)
	require.Equal(t, "void disabledIsDefined();", prototypes[3].Prototype)
	require.Equal(t, "int useMyType(MyType type);", prototypes[4].Prototype)

	require.Equal(t, 18, line)
}

func TestCTagsToPrototypesShouldDealFunctionWithDifferentSignatures(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserShouldDealFunctionWithDifferentSignatures.txt", "/tmp/test260613593/preproc/ctags_target.cpp")

	require.Equal(t, 1, len(prototypes))
	require.Equal(t, "boolean getBytes( byte addr, int amount );", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test260613593/preproc/ctags_target.cpp", prototypes[0].File)

	require.Equal(t, 5031, line)
}

func TestCTagsToPrototypesClassMembersAreFilteredOut(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserClassMembersAreFilteredOut.txt", "/tmp/test834438754/preproc/ctags_target.cpp")

	require.Equal(t, 2, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test834438754/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)

	require.Equal(t, 14, line)
}

func TestCTagsToPrototypesStructWithFunctions(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserStructWithFunctions.txt", "/tmp/build7315640391316178285.tmp/preproc/ctags_target.cpp")

	require.Equal(t, 2, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/build7315640391316178285.tmp/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)

	require.Equal(t, 16, line)
}

func TestCTagsToPrototypesDefaultArguments(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserDefaultArguments.txt", "/tmp/test179252494/preproc/ctags_target.cpp")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "void test(int x = 1);", prototypes[0].Prototype)
	require.Equal(t, "void setup();", prototypes[1].Prototype)
	require.Equal(t, "/tmp/test179252494/preproc/ctags_target.cpp", prototypes[1].File)
	require.Equal(t, "void loop();", prototypes[2].Prototype)

	require.Equal(t, 2, line)
}

func TestCTagsToPrototypesNamespace(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserNamespace.txt", "/tmp/test030883150/preproc/ctags_target.cpp")

	require.Equal(t, 2, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test030883150/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)

	require.Equal(t, 8, line)
}

func TestCTagsToPrototypesStatic(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserStatic.txt", "/tmp/test542833488/preproc/ctags_target.cpp")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test542833488/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "void doStuff();", prototypes[2].Prototype)
	require.Equal(t, "static", prototypes[2].Modifiers)

	require.Equal(t, 2, line)
}

func TestCTagsToPrototypesFunctionPointer(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserFunctionPointer.txt", "/tmp/test547238273/preproc/ctags_target.cpp")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "void t1Callback();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test547238273/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void setup();", prototypes[1].Prototype)
	require.Equal(t, "void loop();", prototypes[2].Prototype)

	require.Equal(t, 2, line)
}

func TestCTagsToPrototypesFunctionPointers(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserFunctionPointers.txt", "/tmp/test907446433/preproc/ctags_target.cpp")
	require.Equal(t, 2, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test907446433/preproc/ctags_target.cpp", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)

	require.Equal(t, 2, line)
}

func TestCTagsToPrototypesFunctionPointersNoIndirect(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsParserFunctionPointersNoIndirect.txt", "/tmp/test547238273/preproc/bug_callback.ino")
	require.Equal(t, 5, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "/tmp/test547238273/preproc/bug_callback.ino", prototypes[0].File)
	require.Equal(t, "void loop();", prototypes[1].Prototype)

	require.Equal(t, 10, line)
}

func TestCTagsRunnerSketchWithClassFunction(t *testing.T) {
	prototypes, _ := producePrototypes(t, "TestCTagsRunnerSketchWithClassFunction.txt", "/home/megabug/Workspace/arduino-builder/src/github.com/arduino/arduino-builder/test/sketch_class_function/sketch_class_function.ino")

	require.Equal(t, 3, len(prototypes))
	require.Equal(t, "void setup();", prototypes[0].Prototype)
	require.Equal(t, "void loop();", prototypes[1].Prototype)
	require.Equal(t, "void asdf();", prototypes[2].Prototype)
}

func TestCTagsRunnerSketchWithMultiFile(t *testing.T) {
	prototypes, line := producePrototypes(t, "TestCTagsRunnerSketchWithMultifile.txt", "/tmp/apUNI8a/main.ino")

	require.Equal(t, 0, line)
	require.Equal(t, "void A7105_Setup();", prototypes[0].Prototype)
	require.Equal(t, "void A7105_Reset();", prototypes[1].Prototype)
	require.Equal(t, "int A7105_calibrate_VCB(uint8_t channel);", prototypes[2].Prototype)
}
