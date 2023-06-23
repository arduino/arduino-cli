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
