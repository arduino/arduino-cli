/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package ctags

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func producePrototypes(t *testing.T, filename string, mainFile string) ([]*types.Prototype, int) {
	bytes, err := ioutil.ReadFile(filepath.Join("test_data", filename))
	require.NoError(t, err)

	parser := &CTagsParser{}
	parser.Parse(string(bytes), paths.New(mainFile))
	return parser.GeneratePrototypes()
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
