// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package f

import "sync"

// DiscardCh consume all incoming messages from the given channel until its closed.
func DiscardCh[T any](ch <-chan T) {
	for range ch {
	}
}

// Future is an object that holds a result value. The value may be read and
// written asynchronously.
type Future[T any] interface {
	Send(T)
	Await() T
}

type future[T any] struct {
	wg    sync.WaitGroup
	value T
}

// NewFuture creates a new Future[T]
func NewFuture[T any]() Future[T] {
	res := &future[T]{}
	res.wg.Add(1)
	return res
}

// Send a result in the Future. Threads waiting for result will be unlocked.
func (f *future[T]) Send(value T) {
	f.value = value
	f.wg.Done()
}

// Await for a result from the Future, blocks until a result is available.
func (f *future[T]) Await() T {
	f.wg.Wait()
	return f.value
}
