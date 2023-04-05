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
	"bytes"
	"strings"

	"github.com/arduino/arduino-cli/arduino/sketch"
	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/arduino-cli/legacy/builder/ctags"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func RunCTags(sketch *sketch.Sketch, source string, targetFileName string, buildProperties *properties.Map, preprocPath *paths.Path,
) (ctagsStdout, ctagsStderr []byte, prototypesLineWhereToInsert int, prototypes []*ctags.Prototype, err error) {
	if err = preprocPath.MkdirAll(); err != nil {
		return
	}

	ctagsTargetFilePath := preprocPath.Join(targetFileName)
	if err = ctagsTargetFilePath.WriteFile([]byte(source)); err != nil {
		return
	}

	ctagsBuildProperties := properties.NewMap()
	ctagsBuildProperties.Set("tools.ctags.path", "{runtime.tools.ctags.path}")
	ctagsBuildProperties.Set("tools.ctags.cmd.path", "{path}/ctags")
	ctagsBuildProperties.Set("tools.ctags.pattern", `"{cmd.path}" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives "{source_file}"`)
	ctagsBuildProperties.Merge(buildProperties)
	ctagsBuildProperties.Merge(ctagsBuildProperties.SubTree("tools").SubTree("ctags"))
	ctagsBuildProperties.SetPath("source_file", ctagsTargetFilePath)

	pattern := ctagsBuildProperties.Get("pattern")
	if pattern == "" {
		err = errors.Errorf(tr("%s pattern is missing"), "ctags")
		return
	}

	commandLine := ctagsBuildProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return
	}
	proc, err := executils.NewProcess(nil, parts...)
	if err != nil {
		return
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	proc.RedirectStdoutTo(stdout)
	proc.RedirectStderrTo(stderr)
	if err = proc.Run(); err != nil {
		return
	}
	stderr.WriteString(strings.Join(parts, " "))
	ctagsStdout = stdout.Bytes()
	ctagsStderr = stderr.Bytes()
	if err != nil {
		return
	}

	parser := &ctags.CTagsParser{}
	parser.Parse(ctagsStdout, sketch.MainFile)
	parser.FixCLinkageTagsDeclarations()

	prototypes, line := parser.GeneratePrototypes()
	if line != -1 {
		prototypesLineWhereToInsert = line
	}
	return
}
