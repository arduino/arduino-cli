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
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	properties "github.com/arduino/go-properties-orderedmap"
	"golang.org/x/exp/slices"

	"github.com/arduino/arduino-cli/arduino/builder/utils"
	"github.com/arduino/arduino-cli/arduino/globals"
	"github.com/arduino/arduino-cli/legacy/builder/builder_utils"
	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/types"
)

type ExportProjectCMake struct {
	// Was there an error while compiling the sketch?
	SketchError bool
}

var lineMatcher = regexp.MustCompile(`^#line\s\d+\s"`)

func (s *ExportProjectCMake) Run(ctx *types.Context) error {
	// copies the contents of the file named src to the file named
	// by dst. The file will be created if it does not already exist. If the
	// destination file exists, all it's contents will be replaced by the contents
	// of the source file. The file mode will be copied from the source and
	// the copied data is synced/flushed to stable storage.
	// TODO: Replace with call to go-paths-helper...
	copyFile := func(src, dst string) (err error) {
		in, err := os.Open(src)
		if err != nil {
			return
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			return
		}
		defer func() {
			if e := out.Close(); e != nil {
				err = e
			}
		}()

		_, err = io.Copy(out, in)
		if err != nil {
			return
		}

		err = out.Sync()
		if err != nil {
			return
		}

		si, err := os.Stat(src)
		if err != nil {
			return
		}
		err = os.Chmod(dst, si.Mode())
		if err != nil {
			return
		}

		return
	}

	// recursively copies a directory tree, attempting to preserve permissions.
	// Source directory must exist, destination directory must *not* exist.
	// Symlinks are ignored and skipped.
	// TODO: Replace with call to go-paths-helper...
	var copyDir func(src string, dst string, extensions []string) (err error)
	copyDir = func(src string, dst string, extensions []string) (err error) {
		isAcceptedExtension := func(ext string) bool {
			ext = strings.ToLower(ext)
			for _, valid := range extensions {
				if ext == valid {
					return true
				}
			}
			return false
		}

		src = filepath.Clean(src)
		dst = filepath.Clean(dst)

		si, err := os.Stat(src)
		if err != nil {
			return err
		}
		if !si.IsDir() {
			return fmt.Errorf(tr("source is not a directory"))
		}

		_, err = os.Stat(dst)
		if err != nil && !os.IsNotExist(err) {
			return
		}
		if err == nil {
			return fmt.Errorf(tr("destination already exists"))
		}

		err = os.MkdirAll(dst, si.Mode())
		if err != nil {
			return
		}

		entries, err := os.ReadDir(src)
		if err != nil {
			return
		}

		for _, dirEntry := range entries {
			entry, scopeErr := dirEntry.Info()
			if scopeErr != nil {
				return
			}

			srcPath := filepath.Join(src, entry.Name())
			dstPath := filepath.Join(dst, entry.Name())

			if entry.IsDir() {
				err = copyDir(srcPath, dstPath, extensions)
				if err != nil {
					return
				}
			} else {
				// Skip symlinks.
				if entry.Mode()&os.ModeSymlink != 0 {
					continue
				}

				if !isAcceptedExtension(filepath.Ext(srcPath)) {
					continue
				}

				err = copyFile(srcPath, dstPath)
				if err != nil {
					return
				}
			}
		}

		return
	}

	var validExportExtensions = []string{".a", ".properties"}
	for ext := range globals.SourceFilesValidExtensions {
		validExportExtensions = append(validExportExtensions, ext)
	}
	var validHeaderExtensions = []string{}
	for ext := range globals.HeaderFilesValidExtensions {
		validExportExtensions = append(validExportExtensions, ext)
		validHeaderExtensions = append(validHeaderExtensions, ext)
	}
	var validStaticLibExtensions = []string{".a"}

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
	for _, library := range ctx.SketchLibrariesDetector.ImportedLibraries() {
		// Copy used libraries in the correct folder
		libDir := libBaseFolder.Join(library.DirName)
		mcu := ctx.BuildProperties.Get(constants.BUILD_PROPERTIES_BUILD_MCU)
		copyDir(library.InstallDir.String(), libDir.String(), validExportExtensions)

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
		files, _ := utils.FindFilesInFolder(libDir.Join("src"), true, validStaticLibExtensions...)
		for _, file := range files {
			staticLibDir := file.Parent()
			if !isStaticLib || !strings.Contains(staticLibDir.String(), mcu) {
				staticLibDir.RemoveAll()
			}
		}
	}

	// Copy core + variant in use + preprocessed sketch in the correct folders
	err := copyDir(ctx.BuildProperties.Get("build.core.path"), coreFolder.String(), validExportExtensions)
	if err != nil {
		fmt.Println(err)
	}
	err = copyDir(ctx.BuildProperties.Get("build.variant.path"), coreFolder.Join("variant").String(), validExportExtensions)
	if err != nil {
		fmt.Println(err)
	}

	if err := PreprocessSketch(ctx); err != nil {
		return err
	}

	err = copyDir(ctx.SketchBuildPath.String(), cmakeFolder.Join("sketch").String(), validExportExtensions)
	if err != nil {
		fmt.Println(err)
	}

	// remove "#line 1 ..." from exported c_make folder sketch
	sketchFiles, _ := utils.FindFilesInFolder(cmakeFolder.Join("sketch"), false, validExportExtensions...)

	for _, file := range sketchFiles {
		input, err := file.ReadFile()
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
		err = file.WriteFile([]byte(output))
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
	headerFiles, _ := utils.FindFilesInFolder(cmakeFolder, true, validHeaderExtensions...)
	foldersContainingHeaders := findUniqueFoldersRelative(headerFiles.AsStrings(), cmakeFolder.String())

	// Extract folders with .a in them for adding in static libs paths list
	staticLibs, _ := utils.FindFilesInFolder(cmakeFolder, true, validStaticLibExtensions...)

	// Generate the CMakeLists global file

	projectName := ctx.Sketch.Name

	cmakelist := "cmake_minimum_required(VERSION 3.5.0)\n"
	cmakelist += "INCLUDE(FindPkgConfig)\n"
	cmakelist += "project (" + projectName + " C CXX)\n"
	cmakelist += "add_definitions (" + strings.Join(defines, " ") + " " + strings.Join(linkerflags, " ") + ")\n"
	cmakelist += "include_directories (" + foldersContainingHeaders + ")\n"

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
		lib := staticLib.Base()
		lib = strings.TrimPrefix(lib, "lib")
		lib = strings.TrimSuffix(lib, ".a")
		if !slices.Contains(dynamicLibsFromGccMinusL, lib) {
			linkGroup += " " + lib
			cmakelist += "add_library (" + lib + " STATIC IMPORTED)\n"
			location := strings.TrimPrefix(staticLib.String(), cmakeFolder.String())
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
	appendIfNotPresent := func(target []string, elements ...string) []string {
		for _, element := range elements {
			if !slices.Contains(target, element) {
				target = append(target, element)
			}
		}
		return target
	}

	command, _ := builder_utils.PrepareCommandForRecipe(ctx.BuildProperties, recipe, true)

	for _, arg := range command.GetArgs() {
		if strings.HasPrefix(arg, "-D") {
			*defines = appendIfNotPresent(*defines, arg)
			continue
		}
		if strings.HasPrefix(arg, "-l") {
			*dynamicLibs = appendIfNotPresent(*dynamicLibs, arg[2:])
			continue
		}
		if strings.HasPrefix(arg, "-L") {
			*linkDirectories = appendIfNotPresent(*linkDirectories, arg[2:])
			continue
		}
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "-I") && !strings.HasPrefix(arg, "-o") {
			// HACK : from linkerflags remove MMD (no cache is produced)
			if !strings.HasPrefix(arg, "-MMD") {
				*linkerflags = appendIfNotPresent(*linkerflags, arg)
			}
		}
	}
}

func findUniqueFoldersRelative(slice []string, base string) string {
	var out []string
	for _, element := range slice {
		path := filepath.Dir(element)
		path = strings.TrimPrefix(path, base+"/")
		if !slices.Contains(out, path) {
			out = append(out, path)
		}
	}
	return strings.Join(out, " ")
}
