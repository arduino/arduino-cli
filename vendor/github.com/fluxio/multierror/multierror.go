// Package multierror defines an error Accumulator to contain multiple errors.
package multierror

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/fluxio/iohelpers/line"
)

// Accumulator is an error accumulator.
//
// Usage:
//
//     var errors multierror.Accumulator
//     errors.Push(returnsErrOrNil())
//     errors.Push(returnsErrOrNil())
//     errors.Push(returnsErrOrNil())
//     return errors.Error()
type Accumulator []error

// Push adds an error to the Accumulator.  If err is nil, then Accumulator is
// not affected.
func (m *Accumulator) Push(err error) {
	if err == nil {
		return
	}
	// Check for a Accumulator
	if e, ok := err.(_error); ok {
		*m = append(*m, e...)
		return
	}

	*m = append(*m, err)
}

// Pushf adds a formatted error string to the Accumulator.  It is a shortcut
// for Push(fmt.Error(...)).
func (m *Accumulator) Pushf(fmtstr string, args ...interface{}) {
	*m = append(*m, fmt.Errorf(fmtstr, args...))
}

// PushWithf adds a formatted error string to m if err is non-nil.  err is
// passed as the first argument to fmtstr, and any additional arguments in args
// are passed as the remaining arguments.
func (m *Accumulator) PushWithf(fmtstr string, err error, args ...interface{}) {
	if err != nil {
		m.Pushf(fmtstr, append([]interface{}{err}, args...)...)
	}
}

// Error returns the accumulated errors.  If no errors have been pushed onto the
// accumulator, then nil will be returned.  If only a single error has been
// pushed onto the accumulator, it is returned directly.
//
// Note that Accumulator can't implement the error interface directly because
// of http://golang.org/doc/faq#nil_error.  That is, return a nil Accumulator
// as an error interface would result in a non-nil error to a nil Accumulator.
func (m *Accumulator) Error() error {
	if len(*m) == 0 {
		return nil
	} else if len(*m) == 1 {
		return (*m)[0]
	}
	return _error(*m)
}

// String prints the accumulated errors or "nil" if no errors have been pushed.
func (m Accumulator) String() string {
	return _error(m).Error()
}

// This type implements the actual error interface.  This is separate from
// Accumulator so to avoid accidentally returning an interface to a nil pointer.
// http://golang.org/doc/faq#nil_error
type _error []error

func (m _error) Error() string {
	if len(m) == 0 {
		return "nil"
	}
	if len(m) == 1 {
		return m[0].Error()
	}

	buf := &bytes.Buffer{}
	w := &line.PrefixWriter{buf, []byte(`:   `), true}
	fmt.Fprintf(w, "%d errors:", len(m))
	for _, err := range m {
		fmt.Fprintf(w, "\n%v", err)
	}
	return buf.String()
}

// ConcurrentAccumulator is a thread-safe accumulator that can be used across
// many goroutines.
type ConcurrentAccumulator struct {
	errs  Accumulator
	mutex sync.Mutex
}

func (c *ConcurrentAccumulator) Push(err error) {
	if err == nil {
		return
	}
	c.mutex.Lock()
	c.errs.Push(err)
	c.mutex.Unlock()
}

func (c *ConcurrentAccumulator) Pushf(fmtstr string, args ...interface{}) {
	c.mutex.Lock()
	c.errs.Pushf(fmtstr, args...)
	c.mutex.Unlock()
}

func (c *ConcurrentAccumulator) Error() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.errs.Error()
}

// FilterAccumulator allows
type FilterAccumulator struct {
	// Filter is a client-provided function that, given a (possibly-nil)
	// error, returns a transformed error.  Filter is allowed to suppress an
	// error by returning nil.
	Filter func(err error) error
	Accumulator
}

func (a *FilterAccumulator) Push(err error) {
	err = a.Filter(err)
	a.Accumulator.Push(err)
}

func (a *FilterAccumulator) Pushf(fmtstr string, args ...interface{}) {
	a.Accumulator.Push(a.Filter(fmt.Errorf(fmtstr, args...)))
}

// ReplaceNewlines collapses an error's message into a single line.  Using this
// as a FilterAccumulator.Filter helps when accumulating errors for systems that
// split log reports on line boundaries (like loggly).
func ReplaceNewlines(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(strings.Replace(err.Error(), "\n", " ", -1))
}
