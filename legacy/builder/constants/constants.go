// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
// Copyright 2015 Matthijs Kooijman
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

package constants

const BUILD_OPTIONS_FILE = "build.options.json"
const BUILD_PROPERTIES_ARCHIVE_FILE = "archive_file"
const BUILD_PROPERTIES_ARCHIVE_FILE_PATH = "archive_file_path"
const BUILD_PROPERTIES_ARCH_OVERRIDE_CHECK = "architecture.override_check"
const BUILD_PROPERTIES_BOOTLOADER_FILE = "bootloader.file"
const BUILD_PROPERTIES_BOOTLOADER_NOBLINK = "bootloader.noblink"
const BUILD_PROPERTIES_BUILD_BOARD = "build.board"
const BUILD_PROPERTIES_BUILD_MCU = "build.mcu"
const BUILD_PROPERTIES_COMPILER_C_ELF_FLAGS = "compiler.c.elf.flags"
const BUILD_PROPERTIES_COMPILER_LDFLAGS = "compiler.ldflags"
const BUILD_PROPERTIES_COMPILER_CPP_FLAGS = "compiler.cpp.flags"
const BUILD_PROPERTIES_COMPILER_WARNING_FLAGS = "compiler.warning_flags"
const BUILD_PROPERTIES_FQBN = "build.fqbn"
const BUILD_PROPERTIES_INCLUDES = "includes"
const BUILD_PROPERTIES_OBJECT_FILE = "object_file"
const BUILD_PROPERTIES_PATTERN = "pattern"
const BUILD_PROPERTIES_PID = "pid"
const BUILD_PROPERTIES_PREPROCESSED_FILE_PATH = "preprocessed_file_path"
const BUILD_PROPERTIES_RUNTIME_PLATFORM_PATH = "runtime.platform.path"
const BUILD_PROPERTIES_SOURCE_FILE = "source_file"
const BUILD_PROPERTIES_TOOLS_KEY = "tools"
const BUILD_PROPERTIES_VID = "vid"
const CTAGS = "ctags"
const EMPTY_STRING = ""
const FILE_CTAGS_TARGET_FOR_GCC_MINUS_E = "ctags_target_for_gcc_minus_e.cpp"
const FILE_PLATFORM_KEYS_REWRITE_TXT = "platform.keys.rewrite.txt"
const FOLDER_BOOTLOADERS = "bootloaders"
const FOLDER_CORE = "core"
const FOLDER_PREPROC = "preproc"
const FOLDER_SKETCH = "sketch"
const FOLDER_TOOLS = "tools"
const hooks_core = hooks + ".core"
const HOOKS_CORE_POSTBUILD = hooks_core + hooks_postbuild_suffix
const HOOKS_CORE_PREBUILD = hooks_core + hooks_prebuild_suffix
const hooks_libraries = hooks + ".libraries"
const HOOKS_LIBRARIES_POSTBUILD = hooks_libraries + hooks_postbuild_suffix
const HOOKS_LIBRARIES_PREBUILD = hooks_libraries + hooks_prebuild_suffix
const hooks_linking = hooks + ".linking"
const HOOKS_LINKING_POSTLINK = hooks_linking + hooks_postlink_suffix
const HOOKS_LINKING_PRELINK = hooks_linking + hooks_prelink_suffix
const hooks_objcopy = hooks + ".objcopy"
const HOOKS_OBJCOPY_POSTOBJCOPY = hooks_objcopy + hooks_postobjcopy_suffix
const HOOKS_OBJCOPY_PREOBJCOPY = hooks_objcopy + hooks_preobjcopy_suffix
const HOOKS_PATTERN_SUFFIX = ".pattern"
const HOOKS_POSTBUILD = hooks + hooks_postbuild_suffix
const hooks_postbuild_suffix = ".postbuild"
const hooks_postlink_suffix = ".postlink"
const hooks_postobjcopy_suffix = ".postobjcopy"
const HOOKS_PREBUILD = hooks + hooks_prebuild_suffix
const hooks_prebuild_suffix = ".prebuild"
const hooks_prelink_suffix = ".prelink"
const hooks_preobjcopy_suffix = ".preobjcopy"
const hooks = "recipe.hooks"
const hooks_sketch = hooks + ".sketch"
const HOOKS_SKETCH_POSTBUILD = hooks_sketch + hooks_postbuild_suffix
const HOOKS_SKETCH_PREBUILD = hooks_sketch + hooks_prebuild_suffix
const LIBRARY_ALL_ARCHS = "*"
const LIBRARY_EMAIL = "email"
const LIBRARY_FOLDER_ARCH = "arch"
const LIBRARY_FOLDER_SRC = "src"
const LOG_LEVEL_DEBUG = "debug"
const LOG_LEVEL_ERROR = "error"
const LOG_LEVEL_INFO = "info"
const LOG_LEVEL_WARN = "warn"

const PACKAGE_NAME = "name"
const PACKAGE_TOOLS = "tools"
const PLATFORM_ARCHITECTURE = "architecture"
const PLATFORM_NAME = "name"
const PLATFORM_REWRITE_NEW = "new"
const PLATFORM_REWRITE_OLD = "old"
const PLATFORM_URL = "url"
const PLATFORM_VERSION = "version"
const PROPERTY_WARN_DATA_PERCENT = "build.warn_data_percentage"
const PROPERTY_UPLOAD_MAX_SIZE = "upload.maximum_size"
const PROPERTY_UPLOAD_MAX_DATA_SIZE = "upload.maximum_data_size"
const RECIPE_AR_PATTERN = "recipe.ar.pattern"
const RECIPE_C_COMBINE_PATTERN = "recipe.c.combine.pattern"
const RECIPE_C_PATTERN = "recipe.c.o.pattern"
const RECIPE_CPP_PATTERN = "recipe.cpp.o.pattern"
const RECIPE_SIZE_PATTERN = "recipe.size.pattern"
const RECIPE_PREPROC_MACROS = "recipe.preproc.macros"
const RECIPE_S_PATTERN = "recipe.S.o.pattern"
const RECIPE_SIZE_REGEXP = "recipe.size.regex"
const RECIPE_SIZE_REGEXP_DATA = "recipe.size.regex.data"
const RECIPE_SIZE_REGEXP_EEPROM = "recipe.size.regex.eeprom"
const REWRITING_DISABLED = "disabled"
const REWRITING = "rewriting"
const SPACE = " "
const TOOL_NAME = "name"
const TOOL_URL = "url"
const TOOL_VERSION = "version"
