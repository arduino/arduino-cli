//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build linux darwin freebsd openbsd

package unixutils // "go.bug.st/serial.v1/unixutils"

import (
	"time"

	"github.com/creack/goselect"
)

// FDSet is a set of file descriptors suitable for a select call
type FDSet struct {
	set goselect.FDSet
	max uintptr
}

// NewFDSet creates a set of file descriptors suitable for a Select call.
func NewFDSet(fds ...int) *FDSet {
	s := &FDSet{}
	s.Add(fds...)
	return s
}

// Add adds the file descriptors passed as parameter to the FDSet.
func (s *FDSet) Add(fds ...int) {
	for _, fd := range fds {
		f := uintptr(fd)
		s.set.Set(f)
		if f > s.max {
			s.max = f
		}
	}
}

// FDResultSets contains the result of a Select operation.
type FDResultSets struct {
	readable  *goselect.FDSet
	writeable *goselect.FDSet
	errors    *goselect.FDSet
}

// IsReadable test if a file descriptor is ready to be read.
func (r *FDResultSets) IsReadable(fd int) bool {
	return r.readable.IsSet(uintptr(fd))
}

// IsWritable test if a file descriptor is ready to be written.
func (r *FDResultSets) IsWritable(fd int) bool {
	return r.writeable.IsSet(uintptr(fd))
}

// IsError test if a file descriptor is in error state.
func (r *FDResultSets) IsError(fd int) bool {
	return r.errors.IsSet(uintptr(fd))
}

// Select performs a select system call,
// file descriptors in the rd set are tested for read-events,
// file descriptors in the wd set are tested for write-events and
// file descriptors in the er set are tested for error-events.
// The function will block until an event happens or the timeout expires.
// The function return an FDResultSets that contains all the file descriptor
// that have a pending read/write/error event.
func Select(rd, wr, er *FDSet, timeout time.Duration) (*FDResultSets, error) {
	max := uintptr(0)
	res := &FDResultSets{}
	if rd != nil {
		// fdsets are copied so the parameters are left untouched
		copyOfRd := rd.set
		res.readable = &copyOfRd
		// Determine max fd.
		max = rd.max
	}
	if wr != nil {
		// fdsets are copied so the parameters are left untouched
		copyOfWr := wr.set
		res.writeable = &copyOfWr
		// Determine max fd.
		if wr.max > max {
			max = wr.max
		}
	}
	if er != nil {
		// fdsets are copied so the parameters are left untouched
		copyOfEr := er.set
		res.errors = &copyOfEr
		// Determine max fd.
		if er.max > max {
			max = er.max
		}
	}

	err := goselect.Select(int(max+1), res.readable, res.writeable, res.errors, timeout)
	return res, err
}
