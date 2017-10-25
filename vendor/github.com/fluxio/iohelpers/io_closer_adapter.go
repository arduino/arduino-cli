package iohelpers

// IoCloserAdapter wraps any object with a method
//   Close()
// (note that this does not return an error) in an io.Closer that
// calls the underlying object's Close() method and returns nil.
type IoCloserAdapter struct {
	Closer interface {
		Close()
	}
}

func (c IoCloserAdapter) Close() error {
	c.Closer.Close()
	return nil
}
