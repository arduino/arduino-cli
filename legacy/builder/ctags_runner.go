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

	"github.com/arduino/arduino-cli/executils"
	"github.com/arduino/go-paths-helper"
	properties "github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

func RunCTags(sourceFile *paths.Path, buildProperties *properties.Map) ([]byte, []byte, error) {
	ctagsBuildProperties := properties.NewMap()
	ctagsBuildProperties.Set("tools.ctags.path", "{runtime.tools.ctags.path}")
	ctagsBuildProperties.Set("tools.ctags.cmd.path", "{path}/ctags")
	ctagsBuildProperties.Set("tools.ctags.pattern", `"{cmd.path}" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives "{source_file}"`)
	ctagsBuildProperties.Merge(buildProperties)
	ctagsBuildProperties.Merge(ctagsBuildProperties.SubTree("tools").SubTree("ctags"))
	ctagsBuildProperties.SetPath("source_file", sourceFile)

	pattern := ctagsBuildProperties.Get("pattern")
	if pattern == "" {
		return nil, nil, errors.Errorf(tr("%s pattern is missing"), "ctags")
	}

	commandLine := ctagsBuildProperties.ExpandPropsInString(pattern)
	parts, err := properties.SplitQuotedString(commandLine, `"'`, false)
	if err != nil {
		return nil, nil, err
	}
	proc, err := executils.NewProcess(nil, parts...)
	if err != nil {
		return nil, nil, err
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	proc.RedirectStdoutTo(stdout)
	proc.RedirectStderrTo(stderr)
	if err = proc.Run(); err != nil {
		return nil, nil, err
	}
	_, _ = stderr.WriteString(strings.Join(parts, " "))
	return stdout.Bytes(), stderr.Bytes(), err
}
