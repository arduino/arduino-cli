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

package daemon_test

import (
	"testing"

	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// TaskProgressAnalyzer analyzes TaskProgress messages for consistency
type TaskProgressAnalyzer struct {
	t       *testing.T
	Results map[string][]*commands.TaskProgress
}

// NewTaskProgressAnalyzer creates a new TaskProgressAnalyzer
func NewTaskProgressAnalyzer(t *testing.T) *TaskProgressAnalyzer {
	return &TaskProgressAnalyzer{
		t:       t,
		Results: map[string][]*commands.TaskProgress{},
	}
}

// Process the given TaskProgress message.
func (a *TaskProgressAnalyzer) Process(progress *commands.TaskProgress) {
	if progress == nil {
		return
	}

	taskName := progress.GetName()
	a.Results[taskName] = append(a.Results[taskName], progress)
}
