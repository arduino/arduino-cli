package line

import (
	"bytes"
	"io"
	"regexp"
)

// CompileOrDie converts a list of regexp strings to the corresponding regexps
// or panics.
func CompileOrDie(filters ...string) []*regexp.Regexp {
	var compiled []*regexp.Regexp
	for _, re := range filters {
		compiled = append(compiled, regexp.MustCompile(re))
	}
	return compiled
}

// SquelchingWriter is an io.Writer that filters each distinct line written to
// it against an array of regexps.  If any match, the line is suppressed.
type SquelchingWriter struct {
	filters []*regexp.Regexp
	dest    io.Writer
	linebuf []byte
}

// Create a SquelchingWriter with the specified list of regexps.
func NewSquelchingWriter(dest io.Writer, filters []*regexp.Regexp) *SquelchingWriter {
	return &SquelchingWriter{
		filters: filters,
		dest:    dest,
	}
}

func (s *SquelchingWriter) Write(p []byte) (n int, err error) {
	combinedOutput := append(s.linebuf, p...)
	lines := bytes.SplitAfter(combinedOutput, []byte("\n"))
	n -= len(s.linebuf)
	for _, line := range lines {
		isComplete := len(line) > 0 && line[len(line)-1] == byte('\n')
		if !isComplete {
			s.linebuf = line
			n += len(line)
		} else if s.filtered(line) {
			n += len(line) // skip it!  pretend we wrote it.
		} else {
			written, err := s.dest.Write(line)
			n += written
			if err != nil {
				return n, err
			}
		}
	}
	return n, err
}

func (s *SquelchingWriter) filtered(line []byte) bool {
	for i := range s.filters {
		if s.filters[i].Match(line) {
			return true
		}
	}
	return false
}
