package diagnostics

import (
	"github.com/sirupsen/logrus"
)

type Store struct {
	results Diagnostics
}

func NewStore() *Store {
	return &Store{}
}

func (m *Store) Parse(cmdline []string, out []byte) {
	compiler := DetectCompilerFromCommandLine(
		cmdline,
		false, // at the moment compiler-probing is not required
	)
	if compiler == nil {
		logrus.Warnf("Could not detect compiler from: %s", cmdline)
		return
	}
	diags, err := ParseCompilerOutput(compiler, out)
	if err != nil {
		logrus.Warnf("Error parsing compiler output: %s", err)
		return
	}
	m.results = append(m.results, diags...)
}

func (m *Store) Diagnostics() Diagnostics {
	return m.results
}
