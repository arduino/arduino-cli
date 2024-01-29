package diagnosticmanager

import (
	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/diagnostics"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	results         diagnostics.Diagnostics
	lastInsertIndex int
}

func New() *Manager {
	return &Manager{lastInsertIndex: -1}
}

func (m *Manager) Parse(cmdline []string, out []byte) {
	compiler := diagnostics.DetectCompilerFromCommandLine(
		cmdline,
		false, // at the moment compiler-probing is not required
	)
	if compiler == nil {
		logrus.Warnf("Could not detect compiler from: %s", cmdline)
		return
	}
	diags, err := diagnostics.ParseCompilerOutput(compiler, out)
	if err != nil {
		logrus.Warnf("Error parsing compiler output: %s", err)
		return
	}
	m.lastInsertIndex += len(diags)
	m.results = append(m.results, diags...)
}

func (m *Manager) CompilerDiagnostics() diagnostics.Diagnostics {
	return m.results
}
