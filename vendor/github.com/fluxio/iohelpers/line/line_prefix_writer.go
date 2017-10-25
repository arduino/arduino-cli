package line

import (
	"bytes"
	"io"
)

// PrefixWriter is an io.Writer that prepends a prefix to each distinct line
// of the output.  It automatically all writes and inserts prefixes at newlines.
type PrefixWriter struct {
	Dest   io.Writer
	Prefix []byte

	// SkipOnePrefix may be set to true to skip prefixing the next write. This
	// is particular useful when creating a PrefixWriter if you want it to
	// prefix all but the first line.
	SkipOnePrefix bool
}

func (l *PrefixWriter) Write(p []byte) (n int, err error) {
	lines := bytes.SplitAfter(p, []byte("\n"))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		if !l.SkipOnePrefix {
			_, err = l.Dest.Write(l.Prefix)
			if err != nil {
				return n, err
			}
		}

		l.SkipOnePrefix = line[len(line)-1] != byte('\n')

		written, err := l.Dest.Write(line)
		n += written
		if err != nil {
			return n, err
		}
	}

	return n, err
}
