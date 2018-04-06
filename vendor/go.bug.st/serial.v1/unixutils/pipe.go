//
// Copyright 2014-2017 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

// +build linux darwin freebsd openbsd

package unixutils // import "go.bug.st/serial.v1/unixutils"

import "syscall"
import "fmt"

// Pipe represents a unix-pipe
type Pipe struct {
	opened bool
	rd     int
	wr     int
}

// Open creates a new pipe
func (p *Pipe) Open() error {
	fds := []int{0, 0}
	if err := syscall.Pipe(fds); err != nil {
		return err
	}
	p.rd = fds[0]
	p.wr = fds[1]
	p.opened = true
	return nil
}

// ReadFD returns the file handle for the read side of the pipe.
func (p *Pipe) ReadFD() int {
	if !p.opened {
		return -1
	}
	return p.rd
}

// WriteFD returns the flie handle for the write side of the pipe.
func (p *Pipe) WriteFD() int {
	if !p.opened {
		return -1
	}
	return p.wr
}

// Write to the pipe the content of data. Returns the numbre of bytes written.
func (p *Pipe) Write(data []byte) (int, error) {
	if !p.opened {
		return 0, fmt.Errorf("Pipe not opened")
	}
	return syscall.Write(p.wr, data)
}

// Read from the pipe into the data array. Returns the number of bytes read.
func (p *Pipe) Read(data []byte) (int, error) {
	if !p.opened {
		return 0, fmt.Errorf("Pipe not opened")
	}
	return syscall.Read(p.rd, data)
}

// Close the pipe
func (p *Pipe) Close() error {
	if !p.opened {
		return fmt.Errorf("Pipe not opened")
	}
	err1 := syscall.Close(p.rd)
	err2 := syscall.Close(p.wr)
	p.opened = false
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
