package iohelpers

import (
	"io"
)

// RuneWriter abstracts the WriteRune method, which is defined on both
// bytes.Buffer and bufio.Writer, but curiously not reified as an interface.
type RuneWriter interface {
	WriteRune(r rune) (n int, err error)
}

// CopyNRunes copies n runes (or until an error) from src to dst, using the
// UTF-8 encoding for both.  It returns the count of bytes and runes copied, and
// the first error (including io.EOF on read) encountered while copying, if any.
// To make index arithmetic convenient, n < 0 is treated identically to n == 0.
//
// runesWritten gives the count of successfully written whole runes.  On error,
// the last few bytes written to dst may form an incomplete rune, which is not
// counted in runesWritten.
//
// This function transparently replaces invalid UTF-8 encodings in src with the
// Unicode replacement character U+FFFD in the output.  This follows the
// behavior of most Go standard Unicode libraries.
func CopyNRunes(dst RuneWriter, src io.RuneReader, n int64) (
	bytesWritten int64, runesWritten int64, err error,
) {
	for {
		if runesWritten >= n {
			break
		}
		r, _, readErr := src.ReadRune()
		if readErr == io.EOF { // indicates no bytes available
			err = readErr
			break
		} else if readErr != nil {
			return bytesWritten, runesWritten, readErr
		}
		n, writeErr := dst.WriteRune(r)
		bytesWritten += int64(n)
		if writeErr != nil {
			return bytesWritten, runesWritten, writeErr
		}
		runesWritten++
	}
	return bytesWritten, runesWritten, err
}
