package iohelpers

// Converts an error to a read/write/closer that returns the error on any
// operation.
//
// Why?
//
// Because this allows us to handle errors more cleanly:  On load or save
// operations, we ALWAYS return non-nil object.  Since we have to do error
// checking on read/write operations anyways, we eliminate an unnecessary round
// of error checking.
//
// For example, instead of:
//   writer, err := view.Save(...)
//   if err != nil { ... }
//   err = writer.write(...)
//   if err != nil { ... }
//
// We can now do:
//   writer := view.Save(...)  // always valid
//   err := writer.write(...)
//   if err != nil { ... }
//
// ...but with the same error robustness.
//
type ErrorReadWriteCloser struct{ Error error }

func (e ErrorReadWriteCloser) Read(p []byte) (n int, err error)  { return 0, e.Error }
func (e ErrorReadWriteCloser) Write(p []byte) (n int, err error) { return 0, e.Error }
func (e ErrorReadWriteCloser) Close() error                      { return e.Error }
