package line

import (
	"bytes"
	"io"
)

// BoundaryWriter is an io.Writer that passes through only complete lines to
// the underlying writer, buffering partial lines, allowing the underlying
// writer to assume that it never receives a partial line.
type BoundaryWriter struct {
	Target   io.Writer
	buffered string
}

func (e *BoundaryWriter) Write(b []byte) (int, error) {
	n := len(b)
	pos := bytes.LastIndex(b, []byte{'\n'}) + 1
	if pos != 0 {
		msg := e.buffered + string(b[:pos])
		n, err := e.Target.Write([]byte(msg))
		if n != len(msg) || err != nil {
			return n, err
		}
		e.buffered = ""
		b = b[pos:]
	}

	e.buffered += string(b)
	return n, nil
}

// Flush writes any buffered data to the underlying writer.
func (e *BoundaryWriter) Flush() error {
	if e.buffered == "" {
		return nil
	}
	if n, err := e.Target.Write([]byte(e.buffered)); err != nil || n != len(e.buffered) {
		return err
	}
	e.buffered = ""
	return nil
}
