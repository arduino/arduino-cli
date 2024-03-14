// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package commands

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc/metadata"
)

type streamingResponseProxyToChan[T any] struct {
	ctx      context.Context
	respChan chan<- *T
	respLock sync.Mutex
}

func streamResponseToChan[T any](ctx context.Context) (*streamingResponseProxyToChan[T], <-chan *T) {
	respChan := make(chan *T, 1)
	w := &streamingResponseProxyToChan[T]{
		ctx:      ctx,
		respChan: respChan,
	}
	go func() {
		<-ctx.Done()
		w.respLock.Lock()
		close(w.respChan)
		w.respChan = nil
		w.respLock.Unlock()
	}()
	return w, respChan
}

func (w *streamingResponseProxyToChan[T]) Send(resp *T) error {
	w.respLock.Lock()
	if w.respChan != nil {
		w.respChan <- resp
	}
	w.respLock.Unlock()
	return nil
}

func (w *streamingResponseProxyToChan[T]) Context() context.Context {
	return w.ctx
}

func (w *streamingResponseProxyToChan[T]) RecvMsg(m any) error {
	return errors.New("RecvMsg not implemented")
}

func (w *streamingResponseProxyToChan[T]) SendHeader(metadata.MD) error {
	return errors.New("SendHeader not implemented")
}

func (w *streamingResponseProxyToChan[T]) SendMsg(m any) error {
	return errors.New("SendMsg not implemented")
}

func (w *streamingResponseProxyToChan[T]) SetHeader(metadata.MD) error {
	return errors.New("SetHeader not implemented")
}

func (w *streamingResponseProxyToChan[T]) SetTrailer(tr metadata.MD) {
}
