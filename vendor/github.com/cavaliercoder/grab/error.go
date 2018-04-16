package grab

import "errors"

var (
	// ErrBadStatusCode indicates that the server response had a status code that
	// was not in the 200-299 range.
	ErrBadStatusCode = errors.New("server returned a non-2XX status code")

	// ErrBadLength indicates that the server response or an existing file does
	// not match the expected content length.
	ErrBadLength = errors.New("bad content length")

	// ErrBadChecksum indicates that a downloaded file failed to pass checksum
	// validation.
	ErrBadChecksum = errors.New("checksum mismatch")

	// ErrNoFilename indicates that a reasonable filename could not be
	// automatically determined using the URL or response headers from a server.
	ErrNoFilename = errors.New("no filename could be determined")

	// ErrNoTimestamp indicates that a timestamp could not be automatically
	// determined using the reponse headers from the remote server.
	ErrNoTimestamp = errors.New("no timestamp could be determined for the remote file")

	// ErrFileExists indicates that the destination path already exists.
	ErrFileExists = errors.New("file exists")
)
