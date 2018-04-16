package grab

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// isCanceled returns a non-nil error if the given context has been canceled.
func isCanceled(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// setLastModified sets the last modified timestamp of a local file according to
// the Last-Modified header returned by a remote server.
func setLastModified(resp *http.Response, filename string) error {
	// https://tools.ietf.org/html/rfc7232#section-2.2
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified
	header := resp.Header.Get("Last-Modified")
	if header == "" {
		return nil
	}
	lastmod, err := time.Parse(http.TimeFormat, header)
	if err != nil {
		return nil
	}
	return os.Chtimes(filename, lastmod, lastmod)
}

// mkdirp creates all missing parent directories for the destination file path.
func mkdirp(path string) error {
	dir := filepath.Dir(path)
	if fi, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error checking destination directory: %v", err)
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating destination directory: %v", err)
		}
	} else if !fi.IsDir() {
		panic("destination path is not directory")
	}
	return nil
}
