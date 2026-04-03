// This file is part of arduino-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
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

package compile

import (
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestArduinoPreprocessCache(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	t.Cleanup(env.CleanUp)

	// Install Arduino AVR Boards
	_, _, err := cli.Run("core", "install", "arduino:avr@1.8.6")
	require.NoError(t, err)

	t.Run("ChangesToSketchInvalidateCache", func(t *testing.T) {
		// Create a tmp sketch
		tmp, err := paths.MkTempDir("", "")
		require.NoError(t, err)
		t.Cleanup(func() { _ = tmp.RemoveAll() })
		sketch := tmp.Join("sketch")
		require.NoError(t, sketch.MkdirAll())
		sketchFile := sketch.Join("sketch.ino")
		require.NoError(t, sketchFile.WriteFile([]byte(`
void setup() {}
void loop() {}
`)))

		// Run compile two times in a row and check that the second time the cache is used
		out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")

		// Touch the sketch file to invalidate the cache
		require.NoError(t, sketchFile.WriteFile([]byte(`
void setup() {}

void loop() {}
`)))

		// Run compile two times in a row and check that the first time the cache is not used and the second time the cache is used
		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")
	})

	t.Run("SketchWithInoAndCppFiles", func(t *testing.T) {
		// Create a tmp sketch
		tmp, err := paths.MkTempDir("", "")
		require.NoError(t, err)
		t.Cleanup(func() { _ = tmp.RemoveAll() })
		sketch := tmp.Join("sketch")
		require.NoError(t, sketch.MkdirAll())
		sketchFile := sketch.Join("sketch.ino")
		require.NoError(t, sketchFile.WriteFile([]byte(`
void myFunction();

void setup() {}
void loop() {
	myFunction();
}
`)))

		cppFile := sketch.Join("myfile.cpp")
		require.NoError(t, cppFile.WriteFile([]byte(`
#include <SPI.h>

void myFunction() {
	SPI.begin();
}
`)))

		// Run compile two times in a row and check that the second time the cache is used
		out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")

		// Updates to the cpp file should not invalidate the sketch preprocessing cache
		require.NoError(t, cppFile.WriteFile([]byte(`
#include <SPI.h>
void myFunction() {
	SPI.begin();
}
`)))
		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")

		// Updates to the ino file should invalidate the sketch preprocessing cache
		require.NoError(t, sketchFile.WriteFile([]byte(`
void myFunction();
void setup() {}
void loop() {
	myFunction();
}
`)))
		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")
	})

	t.Run("ChangesToLibraryInvalidateCache", func(t *testing.T) {
		// Create a tmp sketch
		tmp, err := paths.MkTempDir("", "")
		require.NoError(t, err)
		t.Cleanup(func() { _ = tmp.RemoveAll() })
		sketch := tmp.Join("sketch")
		require.NoError(t, sketch.MkdirAll())
		sketchFile := sketch.Join("sketch.ino")
		require.NoError(t, sketchFile.WriteFile([]byte(`
#include <MyLibrary.h>

void setup() {}
void loop() {
	#ifdef HAS_NEW_FUNCTION
	myNewFunction();
	#else
	myFunction();
	#endif
}
`)))

		library := cli.SketchbookDir().Join("libraries").Join("MyLibrary")
		t.Cleanup(func() { _ = library.RemoveAll() })
		require.NoError(t, library.MkdirAll())
		header := library.Join("MyLibrary.h")
		require.NoError(t, header.WriteFile([]byte(`
#ifndef MY_LIBRARY_H
#define MY_LIBRARY_H

void myFunction();

#endif
`)))

		source := library.Join("MyLibrary.cpp")
		require.NoError(t, source.WriteFile([]byte(`
#include "MyLibrary.h"

void myFunction() {}
`)))

		// Run compile two times in a row and check that the second time the cache is used
		out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")

		// Touch the library header file to invalidate the cache
		require.NoError(t, header.WriteFile([]byte(`
#ifndef MY_LIBRARY_H
#define MY_LIBRARY_H

#define HAS_NEW_FUNCTION

void myNewFunction();

#endif
`)))
		require.NoError(t, source.WriteFile([]byte(`
#include "MyLibrary.h"

void myNewFunction() {}
`)))

		// Run compile two times in a row and check that the first time the cache is not used and the second time the cache is used
		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String())
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.NotContains(t, string(out), "Using cached sketch with function prototypes.")

		out, _, err = cli.Run("compile", "-b", "arduino:avr:uno", "-v", sketch.String(), "--log", "--log-level", "trace")
		require.NoError(t, err)
		require.Contains(t, string(out), "Generating function prototypes...")
		require.Contains(t, string(out), "Using cached sketch with function prototypes.")
	})
}
