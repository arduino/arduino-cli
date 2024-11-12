// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
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
	"crypto/md5"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/arduino/go-paths-helper"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestCompilePart4(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Manually installs a library
	gitUrl := "https://github.com/arduino-libraries/WiFi101.git"
	libPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")
	_, err = git.PlainClone(libPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.16.1"),
	})
	require.NoError(t, err)

	integrationtest.CLISubtests{
		{"WithLibrary", compileWithLibrary},
		{"WithLibraryPriority", compileWithLibraryPriority},
		{"WithDifferentLibrary", recompileWithDifferentLibrary},
	}.Run(t, env, cli)
}

func compileWithLibrary(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchWithWiFi101Dependency"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	libPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")

	// Create new sketch and add library include
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	data, err := sketchFile.ReadFile()
	require.NoError(t, err)
	data = append([]byte("#include <WiFi101.h>\n"), data...)
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--library", libPath.String(), "-v")
	require.NoError(t, err)
	require.Contains(t, string(stdout), "WiFi101")
}

func compileWithLibraryPriority(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "CompileSketchWithLibraryPriority"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	manuallyInstalledLibPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")

	// Install the same library we installed manually
	_, _, err := cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)
	defer cli.SketchbookDir().Join("libraries").RemoveAll()

	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	lines, err := sketchFile.ReadFileAsLines()
	require.NoError(t, err)
	lines = append([]string{"#include <WiFi101.h>\n"}, lines...)
	var data []byte
	for _, l := range lines {
		data = append(data, []byte(l)...)
	}
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--library", manuallyInstalledLibPath.String(), "-v")
	require.NoError(t, err)
	cliInstalledLibPath := cli.SketchbookDir().Join("libraries", "WiFi101")
	expectedOutput := [3]string{
		"Multiple libraries were found for \"WiFi101.h\"",
		"  Used: " + manuallyInstalledLibPath.String(),
		"  Not used: " + cliInstalledLibPath.String(),
	}
	require.Contains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func recompileWithDifferentLibrary(t *testing.T, env *integrationtest.Environment, cli *integrationtest.ArduinoCLI) {
	sketchName := "RecompileCompileSketchWithDifferentLibrary"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	defer sketchPath.RemoveAll()
	fqbn := "arduino:avr:uno"

	// Install library
	_, _, err := cli.Run("lib", "install", "WiFi101")
	require.NoError(t, err)
	defer cli.SketchbookDir().Join("libraries").RemoveAll()

	manuallyInstalledLibPath := cli.SketchbookDir().Join("my-libraries", "WiFi101")

	// Create new sketch and add library include
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)
	sketchFile := sketchPath.Join(sketchName + ".ino")
	lines, err := sketchFile.ReadFileAsLines()
	require.NoError(t, err)
	lines = append([]string{"#include <WiFi101.h>\n"}, lines...)
	var data []byte
	for _, l := range lines {
		data = append(data, []byte(l)...)
	}
	err = sketchFile.WriteFile(data)
	require.NoError(t, err)

	md5 := md5.Sum(([]byte(sketchPath.String())))
	sketchPathMd5 := strings.ToUpper(hex.EncodeToString(md5[:]))
	require.NotEmpty(t, sketchPathMd5)
	buildDir := paths.TempDir().Join("arduino", "sketches", sketchPathMd5)

	// Compile sketch using library not managed by CLI
	stdout, _, err := cli.Run("compile", "-b", fqbn, "--library", manuallyInstalledLibPath.String(), sketchPath.String(), "-v")
	require.NoError(t, err)
	objPath := buildDir.Join("libraries", "WiFi101", "WiFi.cpp.o")
	require.NotContains(t, string(stdout), "Using previously compiled file: "+objPath.String())

	// Compile again using library installed from CLI
	stdout, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "-v")
	require.NoError(t, err)
	require.NotContains(t, string(stdout), "Using previously compiled file: "+objPath.String())
}

func TestCompileManuallyInstalledPlatformUsingBoardsLocalTxt(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	sketchName := "CompileSketchManuallyInstalledPlatformUsingBoardsLocalTxt"
	sketchPath := cli.SketchbookDir().Join(sketchName)
	fqbn := "arduino-beta-development:avr:nessuno"
	_, _, err := cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Manually installs a core in sketchbooks hardware folder
	gitUrl := "https://github.com/arduino/ArduinoCore-avr.git"
	repoDir := cli.SketchbookDir().Join("hardware", "arduino-beta-development", "avr")
	_, err = git.PlainClone(repoDir.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("1.8.3"),
	})
	require.NoError(t, err)

	// Installs also the same core via CLI so all the necessary tools are installed
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Verifies compilation fails because board doesn't exist
	_, stderr, err := cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error during build: Invalid FQBN: board arduino-beta-development:avr:nessuno not found")

	// Use custom boards.local.txt with made arduino:avr:nessuno board
	boardsLocalTxt := repoDir.Join("boards.local.txt")
	err = paths.New("..", "testdata", "boards.local.txt").CopyTo(boardsLocalTxt)
	require.NoError(t, err)

	_, _, err = cli.Run("compile", "--clean", "-b", fqbn, sketchPath.String())
	require.NoError(t, err)
}

func TestCompileWithConflictingLibrariesInclude(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	// Installs conflicting libraries
	gitUrl := "https://github.com/pstolarz/OneWireNg.git"
	oneWireNgLibPath := cli.SketchbookDir().Join("libraries", "onewireng_0_8_1")
	_, err = git.PlainClone(oneWireNgLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("0.8.1"),
	})
	require.NoError(t, err)

	gitUrl = "https://github.com/PaulStoffregen/OneWire.git"
	oneWireLibPath := cli.SketchbookDir().Join("libraries", "onewire_2_3_5")
	_, err = git.PlainClone(oneWireLibPath.String(), false, &git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewTagReferenceName("v2.3.5"),
	})
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_conflicting_libraries_include")
	fqbn := "arduino:avr:uno"

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
	expectedOutput := [3]string{
		"Multiple libraries were found for \"OneWire.h\"",
		"  Used: " + oneWireLibPath.String(),
		"  Not used: " + oneWireNgLibPath.String(),
	}
	require.Contains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func TestCompileWithEsp32BundledLibraries(t *testing.T) {
	// Some esp cores have bundled libraries that are optimize for that architecture,
	// it might happen that if the user has a library with the same name installed conflicts
	// can ensue and the wrong library is used for compilation, thus it fails.
	// This happens because for "historical" reasons these platform have their "name" key
	// in the "library.properties" flag suffixed with "(esp32)" or similar even though that
	// doesn't respect the libraries specification.
	// https://arduino.github.io/arduino-cli/latest/library-specification/#libraryproperties-file-format
	//
	// The reason those libraries have these suffixes is to avoid an annoying bug in the Java IDE
	// that would have caused the libraries that are both bundled with the core and the Java IDE to be
	// always marked as updatable. For more info see: https://github.com/arduino/Arduino/issues/4189
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Update index with esp32 core and install it
	url := "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
	coreVersion := "1.0.6"
	_, _, err = cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "esp32:esp32@"+coreVersion, "--additional-urls="+url)
	require.NoError(t, err)

	// Install a library with the same name as one bundled with the core
	_, _, err = cli.Run("lib", "install", "SD")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_sd_library")
	fqbn := "esp32:esp32:esp32"

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.Error(t, err)

	coreBundledLibPath := cli.DataDir().Join("packages", "esp32", "hardware", "esp32", coreVersion, "libraries", "SD")
	cliInstalledLibPath := cli.SketchbookDir().Join("libraries", "SD")
	expectedOutput := [3]string{
		"Multiple libraries were found for \"OneWire.h\"",
		"  Used: " + coreBundledLibPath.String(),
		"  Not used: " + cliInstalledLibPath.String(),
	}
	require.NotContains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func TestCompileWithEsp8266BundledLibraries(t *testing.T) {
	// Some esp cores have bundled libraries that are optimize for that architecture,
	// it might happen that if the user has a library with the same name installed conflicts
	// can ensue and the wrong library is used for compilation, thus it fails.
	// This happens because for "historical" reasons these platform have their "name" key
	// in the "library.properties" flag suffixed with "(esp32)" or similar even though that
	// doesn't respect the libraries specification.
	// https://arduino.github.io/arduino-cli/latest/library-specification/#libraryproperties-file-format
	//
	// The reason those libraries have these suffixes is to avoid an annoying bug in the Java IDE
	// that would have caused the libraries that are both bundled with the core and the Java IDE to be
	// always marked as updatable. For more info see: https://github.com/arduino/Arduino/issues/4189
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Update index with esp8266 core and install it
	url := "http://arduino.esp8266.com/stable/package_esp8266com_index.json"
	coreVersion := "2.7.4"
	_, _, err = cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "esp8266:esp8266@"+coreVersion, "--additional-urls="+url)
	require.NoError(t, err)

	// Install a library with the same name as one bundled with the core
	_, _, err = cli.Run("lib", "install", "SD")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_sd_library")
	fqbn := "esp8266:esp8266:generic"

	stdout, _, err := cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.Error(t, err)

	coreBundledLibPath := cli.DataDir().Join("packages", "esp8266", "hardware", "esp8266", coreVersion, "libraries", "SD")
	cliInstalledLibPath := cli.SketchbookDir().Join("libraries", "SD")
	expectedOutput := [3]string{
		"Multiple libraries were found for \"OneWire.h\"",
		"  Used: " + coreBundledLibPath.String(),
		"  Not used: " + cliInstalledLibPath.String(),
	}
	require.NotContains(t, string(stdout), expectedOutput[0]+"\n"+expectedOutput[1]+"\n"+expectedOutput[2]+"\n")
}

func TestGenerateCompileCommandsJsonResilience(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// check it didn't fail with esp32@2.0.1 that has a prebuild hook that must run:
	// https://github.com/arduino/arduino-cli/issues/1547
	url := "https://raw.githubusercontent.com/espressif/arduino-esp32/gh-pages/package_esp32_index.json"
	_, _, err = cli.Run("core", "update-index", "--additional-urls="+url)
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "esp32:esp32@2.0.1", "--additional-urls="+url)
	require.NoError(t, err)
	sketchPath := cli.CopySketch("sketch_simple")
	_, _, err = cli.Run("compile", "-b", "esp32:esp32:featheresp32", "--only-compilation-database", sketchPath.String())
	require.NoError(t, err)

	// check it didn't fail on a sketch with a missing include
	sketchPath = cli.CopySketch("sketch_with_missing_include")
	_, _, err = cli.Run("compile", "-b", "esp32:esp32:featheresp32", "--only-compilation-database", sketchPath.String())
	require.NoError(t, err)
}

func TestCompileSketchSketchWithTppFileInclude(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_tpp_file_include")
	fqbn := "arduino:avr:uno"

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
}

func TestCompileSketchWithIppFileInclude(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Download latest AVR
	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	sketchPath := cli.CopySketch("sketch_with_ipp_file_include")
	fqbn := "arduino:avr:uno"

	_, _, err = cli.Run("compile", "-b", fqbn, sketchPath.String(), "--verbose")
	require.NoError(t, err)
}

func TestCompileWithoutUploadAndFqbn(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Create a sketch
	sketchPath := cli.SketchbookDir().Join("SketchSimple")
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	_, stderr, err := cli.Run("compile", sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Missing FQBN (Fully Qualified Board Name)")
}

func TestCompileNonInstalledPlatformWithWrongPackagerAndArch(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Create a sketch
	sketchPath := cli.SketchbookDir().Join("SketchSimple")
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Compile with wrong packager
	_, stderr, err := cli.Run("compile", "-b", "wrong:avr:uno", sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error during build: Platform 'wrong:avr' not found: platform not installed")
	require.Contains(t, string(stderr), "Platform wrong:avr is not found in any known index")

	// Compile with wrong arch
	_, stderr, err = cli.Run("compile", "-b", "arduino:wrong:uno", sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error during build: Platform 'arduino:wrong' not found: platform not installed")
	require.Contains(t, string(stderr), "Platform arduino:wrong is not found in any known index")
}

func TestCompileWithKnownPlatformNotInstalled(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)

	// Create a sketch
	sketchPath := cli.SketchbookDir().Join("SketchSimple")
	_, _, err = cli.Run("sketch", "new", sketchPath.String())
	require.NoError(t, err)

	// Try to compile using a platform found in the index but not installed
	_, stderr, err := cli.Run("compile", "-b", "arduino:avr:uno", sketchPath.String())
	require.Error(t, err)
	require.Contains(t, string(stderr), "Error during build: Platform 'arduino:avr' not found: platform not installed")
	// Verifies command to fix error is shown to user
	require.Contains(t, string(stderr), "Try running `arduino-cli core install arduino:avr`")
}

func TestSketchWithVendoredLibraries(t *testing.T) {
	sketchBook, err := paths.New("testdata", "sketchbook_1").Abs()
	require.NoError(t, err)

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	cli.SetSketchbookDir(sketchBook)

	_, _, err = cli.Run("core", "install", "arduino:avr")
	require.NoError(t, err)

	{
		sketchWithLibsPath := sketchBook.Join("SketchWithLibraries")
		// Sketch should use sketch bundled "MyLib" with and without profiles
		out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", sketchWithLibsPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].name", `"MyLib"`)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].author", `"user"`)
		out, _, err = cli.Run("compile", "--profile", "uno", sketchWithLibsPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].name", `"MyLib"`)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].author", `"user"`)
	}

	{
		sketchWithoutLibsPath := sketchBook.Join("SketchWithoutLibraries")
		// This sketch should take the user-installed MyLib
		out, _, err := cli.Run("compile", "-b", "arduino:avr:uno", sketchWithoutLibsPath.String(), "--format", "json")
		require.NoError(t, err)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].name", `"MyLib"`)
		requirejson.Query(t, out, ".builder_result.used_libraries[0].author", `"upstream"`)
		// This sketch should fail to compile since profiles will not see the user-installed MyLib
		_, _, err = cli.Run("compile", "--profile", "uno", sketchWithoutLibsPath.String())
		require.Error(t, err)
	}
}
