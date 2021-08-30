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

import (
	"github.com/arduino/arduino-cli/i18n"
)

var tr = i18n.Tr

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

var MSG_ARCH_FOLDER_NOT_SUPPORTED = tr("%[1]s folder is no longer supported! See %[2]s for more information", "'arch'", "http://goo.gl/gfFJzU")
var MSG_ARCHIVING_CORE_CACHE = tr("Archiving built core (caching) in: {0}")
var MSG_ERROR_ARCHIVING_CORE_CACHE = tr("Error archiving built core (caching) in {0}: {1}")
var MSG_CORE_CACHE_UNAVAILABLE = tr("Unable to cache built core, please tell {0} maintainers to follow %s", "https://arduino.github.io/arduino-cli/latest/platform-specification/#recipes-to-build-the-corea-archive-file")
var MSG_BOARD_UNKNOWN = tr("Board {0} (platform {1}, package {2}) is unknown")
var MSG_BOOTLOADER_FILE_MISSING = tr("Bootloader file specified but missing: {0}")
var MSG_REBUILD_ALL = tr(", rebuilding all")
var MSG_BUILD_OPTIONS_CHANGED = tr("Build options changed")
var MSG_BUILD_OPTIONS_INVALID = tr("{0} invalid")
var MSG_CANT_FIND_SKETCH_IN_PATH = tr("Unable to find {0} in {1}")
var MSG_FQBN_INVALID = tr("{0} is not a valid fully qualified board name. Required format is targetPackageName:targetPlatformName:targetBoardName.")
var MSG_SKIP_PRECOMPILED_LIBRARY = tr("Skipping dependencies detection for precompiled library {0}")
var MSG_FIND_INCLUDES_FAILED = tr("Error while detecting libraries included by {0}")
var MSG_LIB_LEGACY = tr("(legacy)")
var MSG_LIBRARIES_MULTIPLE_LIBS_FOUND_FOR = tr("Multiple libraries were found for \"{0}\"")
var MSG_LIBRARIES_NOT_USED = tr(" Not used: {0}")
var MSG_LIBRARIES_USED = tr(" Used: {0}")
var MSG_LIBRARY_CAN_USE_SRC_AND_UTILITY_FOLDERS = tr("Library can't use both '%[1]s' and '%[2]s' folders. Double check {0}", "src", "utility")
var MSG_LIBRARY_INCOMPATIBLE_ARCH = tr("WARNING: library {0} claims to run on {1} architecture(s) and may be incompatible with your current board which runs on {2} architecture(s).")
var MSG_LOOKING_FOR_RECIPES = tr("Looking for recipes like {0}*{1}")
var MSG_MISSING_BUILD_BOARD = tr("Warning: Board {0}:{1}:{2} doesn''t define a %s preference. Auto-set to: {3}", "''build.board''")
var MSG_MISSING_CORE_FOR_BOARD = tr("Selected board depends on '{0}' core (not installed).")
var MSG_PACKAGE_UNKNOWN = tr("{0}: Unknown package")
var MSG_PLATFORM_UNKNOWN = tr("Platform {0} (package {1}) is unknown")
var MSG_PROGRESS = tr("Progress {0}")
var MSG_PROP_IN_LIBRARY = tr("Missing '{0}' from library in {1}")
var MSG_RUNNING_COMMAND = tr("Ts: {0} - Running: {1}")
var MSG_RUNNING_RECIPE = tr("Running recipe: {0}")
var MSG_SETTING_BUILD_PATH = tr("Setting build path to {0}")
var MSG_SIZER_TEXT_FULL = tr("Sketch uses {0} bytes ({2}%%) of program storage space. Maximum is {1} bytes.")
var MSG_SIZER_DATA_FULL = tr("Global variables use {0} bytes ({2}%%) of dynamic memory, leaving {3} bytes for local variables. Maximum is {1} bytes.")
var MSG_SIZER_DATA = tr("Global variables use {0} bytes of dynamic memory.")
var MSG_SIZER_TEXT_TOO_BIG = tr("Sketch too big; see %s for tips on reducing it.", "https://support.arduino.cc/hc/en-us/articles/360013825179")
var MSG_SIZER_DATA_TOO_BIG = tr("Not enough memory; see %s for tips on reducing your footprint.", "https://support.arduino.cc/hc/en-us/articles/360013825179")
var MSG_SIZER_LOW_MEMORY = tr("Low memory available, stability problems may occur.")
var MSG_SIZER_ERROR_NO_RULE = tr("Couldn't determine program size")
var MSG_SKETCH_CANT_BE_IN_BUILDPATH = tr("Sketch cannot be located in build path. Please specify a different build path")
var MSG_UNKNOWN_SKETCH_EXT = tr("Unknown sketch file extension: {0}")
var MSG_USING_LIBRARY_AT_VERSION = tr("Using library {0} at version {1} in folder: {2} {3}")
var MSG_USING_LIBRARY = tr("Using library {0} in folder: {1} {2}")
var MSG_USING_BOARD = tr("Using board '{0}' from platform in folder: {1}")
var MSG_USING_CORE = tr("Using core '{0}' from platform in folder: {1}")
var MSG_USING_PREVIOUS_COMPILED_FILE = tr("Using previously compiled file: {0}")
var MSG_USING_CACHED_INCLUDES = tr("Using cached library dependencies for file: {0}")
var MSG_WARNING_LIB_INVALID_CATEGORY = tr("WARNING: Category '{0}' in library {1} is not valid. Setting to '{2}'")
var MSG_WARNING_PLATFORM_OLD_VALUES = tr("Warning: platform.txt from core '{0}' contains deprecated {1}, automatically converted to {2}. Consider upgrading this core.")
var MSG_WARNING_SPURIOUS_FILE_IN_LIB = tr("WARNING: Spurious {0} folder in '{1}' library")

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
