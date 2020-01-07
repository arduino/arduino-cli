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

import properties "github.com/arduino/go-properties-orderedmap"

// CtagsProperties are the platform properties needed to run ctags
var CtagsProperties = properties.NewFromHashmap(map[string]string{
	// Ctags
	"tools.ctags.path":     "{runtime.tools.ctags.path}",
	"tools.ctags.cmd.path": "{path}/ctags",
	"tools.ctags.pattern":  `"{cmd.path}" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives "{source_file}"`,

	// additional entries
	"tools.avrdude.path": "{runtime.tools.avrdude.path}",

	"preproc.macros.flags": "-w -x c++ -E -CC",
	//"preproc.macros.compatibility_flags": `{build.mbed_api_include} {build.nRF51822_api_include} {build.ble_api_include} {compiler.libsam.c.flags} {compiler.arm.cmsis.path} {build.variant_system_include}`,
	//"recipe.preproc.macros":              `"{compiler.path}{compiler.cpp.cmd}" {compiler.cpreprocessor.flags} {compiler.cpp.flags} {preproc.macros.flags} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.cpp.extra_flags} {build.extra_flags} {preproc.macros.compatibility_flags} {includes} "{source_file}" -o "{preprocessed_file_path}"`,
})
