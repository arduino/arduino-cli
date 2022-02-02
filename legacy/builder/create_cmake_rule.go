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

package builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	properties "github.com/arduino/go-properties-orderedmap"

	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
)

var VALID_EXPORT_EXTENSIONS = map[string]bool{".h": true, ".c": true, ".hpp": true, ".hh": true, ".cpp": true, ".S": true, ".a": true, ".properties": true}
var DOTHEXTENSION = map[string]bool{".h": true, ".hh": true, ".hpp": true}
var DOTAEXTENSION = map[string]bool{".a": true}

type ExportProjectCMake struct {
	// Was there an error while compiling the sketch?
	SketchError bool
}

var lineMatcher = regexp.MustCompile(`^#line\s\d+\s"`)

func (s *ExportProjectCMake) Run(ctx *types.Context) error {
	if s.SketchError || !canExportCmakeProject(ctx) {
		return nil
	}

	// Create new cmake subFolder - clean if the folder is already there
	cmakeFolder := ctx.BuildPath.Join("_cmake")
	if _, err := cmakeFolder.Stat(); err == nil {
		cmakeFolder.RemoveAll()
	}
	cmakeFolder.MkdirAll()

	// Create lib and build subfolders
	libBaseFolder := cmakeFolder.Join("lib")
	libBaseFolder.MkdirAll()
	buildBaseFolder := cmakeFolder.Join("build")
	buildBaseFolder.MkdirAll()

	// Create core subfolder path (don't create it yet)
	coreFolder := cmakeFolder.Join("core")
	cmakeFile := cmakeFolder.Join("CMakeLists.txt")

	dynamicLibsFromPkgConfig := map[string]bool{}
	extensions := func(ext string) bool { return VALID_EXPORT_EXTENSIONS[ext] }
	staticLibsExtensions := func(ext string) bool { return DOTAEXTENSION[ext] }
	for _, library := range ctx.ImportedLibraries {
		// Copy used libraries in the correct folder
		libDir := libBaseFolder.Join(library.Name)
		mcu := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
		utils.CopyDir(library.InstallDir.String(), libDir.String(), extensions)

		// Read cmake options if available
		isStaticLib := true
		if cmakeOptions, err := properties.LoadFromPath(libDir.Join("src", mcu, "arduino_builder.properties")); err == nil {
			// If the library can be linked dynamically do not copy the library folder
			if pkgs, ok := cmakeOptions.GetOk("cmake.pkg_config"); ok {
				isStaticLib = false
				for _, pkg := range strings.Split(pkgs, " ") {
					dynamicLibsFromPkgConfig[pkg] = true
				}
			}
		}

		// Remove examples folder
		if _, err := libBaseFolder.Join("examples").Stat(); err == nil {
			libDir.Join("examples").RemoveAll()
		}

		// Remove stray folders contining incompatible or not needed libraries archives
		var files []string
		utils.FindFilesInFolder(&files, libDir.Join("src").String(), staticLibsExtensions, true)
		for _, file := range files {
			staticLibDir := filepath.Dir(file)
			if !isStaticLib || !strings.Contains(staticLibDir, mcu) {
				os.RemoveAll(staticLibDir)
			}
		}
	}

	// Copy core + variant in use + preprocessed sketch in the correct folders
	err := utils.CopyDir(ctx.BuildProperties.Get("build.core.path"), coreFolder.String(), extensions)
	if err != nil {
		fmt.Println(err)
	}
	err = utils.CopyDir(ctx.BuildProperties.Get("build.variant.path"), coreFolder.Join("variant").String(), extensions)
	if err != nil {
		fmt.Println(err)
	}

	// Use old ctags method to generate export file
	commands := []types.Command{
		//&ContainerMergeCopySketchFiles{},
		&ContainerAddPrototypes{},
		&FilterSketchSource{Source: &ctx.Source, RemoveLineMarkers: true},
	}

	for _, command := range commands {
		command.Run(ctx)
	}

	err = utils.CopyDir(ctx.SketchBuildPath.String(), cmakeFolder.Join("sketch").String(), extensions)
	if err != nil {
		fmt.Println(err)
	}

	// remove "#line 1 ..." from exported c_make folder sketch
	var sketchFiles []string
	utils.FindFilesInFolder(&sketchFiles, cmakeFolder.Join("sketch").String(), extensions, false)

	for _, file := range sketchFiles {
		input, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		lines := strings.Split(string(input), "\n")

		for i, line := range lines {
			if lineMatcher.MatchString(line) {
				lines[i] = ""
			}
		}
		output := strings.Join(lines, "\n")
		err = ioutil.WriteFile(file, []byte(output), 0644)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Extract CFLAGS, CPPFLAGS and LDFLAGS
	var defines []string
	var linkerflags []string
	var dynamicLibsFromGccMinusL []string
	var linkDirectories []string

	extractCompileFlags(ctx, constants.RECIPE_C_COMBINE_PATTERN, &defines, &dynamicLibsFromGccMinusL, &linkerflags, &linkDirectories)
	extractCompileFlags(ctx, "recipe.c.o.pattern", &defines, &dynamicLibsFromGccMinusL, &linkerflags, &linkDirectories)
	extractCompileFlags(ctx, "recipe.cpp.o.pattern", &defines, &dynamicLibsFromGccMinusL, &linkerflags, &linkDirectories)

	// Extract folders with .h in them for adding in include list
	var headerFiles []string
	isHeader := func(ext string) bool { return DOTHEXTENSION[ext] }
	utils.FindFilesInFolder(&headerFiles, cmakeFolder.String(), isHeader, true)
	foldersContainingDotH := findUniqueFoldersRelative(headerFiles, cmakeFolder.String())

	// Extract folders with .a in them for adding in static libs paths list
	var staticLibs []string
	utils.FindFilesInFolder(&staticLibs, cmakeFolder.String(), staticLibsExtensions, true)

	// Generate the CMakeLists global file

	projectName := ctx.Sketch.Name

	cmakelist := "cmake_minimum_required(VERSION 3.5.0)\n"
	cmakelist += "INCLUDE(FindPkgConfig)\n"
	cmakelist += "project (" + projectName + " C CXX)\n"
	cmakelist += "add_definitions (" + strings.Join(defines, " ") + " " + strings.Join(linkerflags, " ") + ")\n"
	cmakelist += "include_directories (" + foldersContainingDotH + ")\n"

	// Make link directories relative
	// We can totally discard them since they mostly are outside the core folder
	// If they are inside the core they are not getting copied :)
	var relLinkDirectories []string
	for _, dir := range linkDirectories {
		if strings.Contains(dir, cmakeFolder.String()) {
			relLinkDirectories = append(relLinkDirectories, strings.TrimPrefix(dir, cmakeFolder.String()))
		}
	}

	// Add SO_PATHS option for libraries not getting found by pkg_config
	cmakelist += "set(EXTRA_LIBS_DIRS \"\" CACHE STRING \"Additional paths for dynamic libraries\")\n"

	linkGroup := ""
	for _, lib := range dynamicLibsFromGccMinusL {
		// Dynamic libraries should be discovered by pkg_config
		cmakelist += "pkg_search_module (" + strings.ToUpper(lib) + " " + lib + ")\n"
		relLinkDirectories = append(relLinkDirectories, "${"+strings.ToUpper(lib)+"_LIBRARY_DIRS}")
		linkGroup += " " + lib
	}
	for lib := range dynamicLibsFromPkgConfig {
		cmakelist += "pkg_search_module (" + strings.ToUpper(lib) + " " + lib + ")\n"
		relLinkDirectories = append(relLinkDirectories, "${"+strings.ToUpper(lib)+"_LIBRARY_DIRS}")
		linkGroup += " ${" + strings.ToUpper(lib) + "_LIBRARIES}"
	}
	cmakelist += "link_directories (" + strings.Join(relLinkDirectories, " ") + " ${EXTRA_LIBS_DIRS})\n"
	for _, staticLib := range staticLibs {
		// Static libraries are fully configured
		lib := filepath.Base(staticLib)
		lib = strings.TrimPrefix(lib, "lib")
		lib = strings.TrimSuffix(lib, ".a")
		if !utils.SliceContains(dynamicLibsFromGccMinusL, lib) {
			linkGroup += " " + lib
			cmakelist += "add_library (" + lib + " STATIC IMPORTED)\n"
			location := strings.TrimPrefix(staticLib, cmakeFolder.String())
			cmakelist += "set_property(TARGET " + lib + " PROPERTY IMPORTED_LOCATION " + "${PROJECT_SOURCE_DIR}" + location + " )\n"
		}
	}

	// Include source files
	// TODO: remove .cpp and .h from libraries example folders
	cmakelist += "file (GLOB_RECURSE SOURCES core/*.c* lib/*.c* sketch/*.c*)\n"

	// Compile and link project
	cmakelist += "add_executable (" + projectName + " ${SOURCES} ${SOURCES_LIBS})\n"
	cmakelist += "target_link_libraries( " + projectName + " -Wl,--as-needed -Wl,--start-group " + linkGroup + " -Wl,--end-group)\n"

	cmakeFile.WriteFile([]byte(cmakelist))

	return nil
}

func canExportCmakeProject(ctx *types.Context) bool {
	return ctx.BuildProperties.Get("compiler.export_cmake") != ""
}

func extractCompileFlags(ctx *types.Context, recipe string, defines, dynamicLibs, linkerflags, linkDirectories *[]string) {
	command, _ := builder_utils.PrepareCommandForRecipe(ctx.BuildProperties, recipe, true, ctx.PackageManager.GetEnvVarsForSpawnedProcess())

	for _, arg := range command.Args {
		if strings.HasPrefix(arg, "-D") {
			*defines = utils.AppendIfNotPresent(*defines, arg)
			continue
		}
		if strings.HasPrefix(arg, "-l") {
			*dynamicLibs = utils.AppendIfNotPresent(*dynamicLibs, arg[2:])
			continue
		}
		if strings.HasPrefix(arg, "-L") {
			*linkDirectories = utils.AppendIfNotPresent(*linkDirectories, arg[2:])
			continue
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "-I") && !strings.HasPrefix(arg, "-o") {
			// HACK : from linkerflags remove MMD (no cache is produced)
			if !strings.HasPrefix(arg, "-MMD") {
				*linkerflags = utils.AppendIfNotPresent(*linkerflags, arg)
			}
		}
	}
}

func findUniqueFoldersRelative(slice []string, base string) string {
	var out []string
	for _, element := range slice {
		path := filepath.Dir(element)
		path = strings.TrimPrefix(path, base+"/")
		if !utils.SliceContains(out, path) {
			out = append(out, path)
		}
	}
	return strings.Join(out, " ")
}
