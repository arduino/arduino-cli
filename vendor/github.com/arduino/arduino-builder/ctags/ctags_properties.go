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
 * Copyright 2015-2018 Arduino LLC (http://www.arduino.cc/)
 */

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
