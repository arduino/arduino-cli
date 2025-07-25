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

package runner

import (
	"context"
	"runtime"
	"sync"
)

// Runner is a helper to run commands in a queue, the commands are immediately exectuded
// in a goroutine as they are enqueued. The runner can be stopped by calling Cancel.
type Runner struct {
	lock      sync.Mutex
	queue     chan<- *enqueuedCommand
	results   map[string]<-chan *Result
	ctx       context.Context
	ctxCancel func()
	wg        sync.WaitGroup
}

type enqueuedCommand struct {
	task   *Task
	accept func(*Result)
}

func (cmd *enqueuedCommand) String() string {
	return cmd.task.String()
}

// New creates a new Runner with the given number of workers.
// If workers is 0, the number of workers will be the number of available CPUs.
func New(inCtx context.Context, workers int) *Runner {
	ctx, cancel := context.WithCancel(inCtx)
	queue := make(chan *enqueuedCommand, 1000)
	r := &Runner{
		ctx:       ctx,
		ctxCancel: cancel,
		queue:     queue,
		results:   map[string]<-chan *Result{},
	}

	// Spawn workers
	if workers == 0 {
		workers = runtime.NumCPU()
	}
	for i := 0; i < workers; i++ {
		r.wg.Add(1)
		go func() {
			worker(ctx, queue)
			r.wg.Done()
		}()
	}

	return r
}

func worker(ctx context.Context, queue <-chan *enqueuedCommand) {
	done := ctx.Done()
	for {
		select {
		case <-done:
			return
		default:
		}

		select {
		case <-done:
			return
		case cmd := <-queue:
			result := cmd.task.Run(ctx)
			cmd.accept(result)
		}
	}
}

func (r *Runner) Enqueue(task *Task) {
	r.lock.Lock()
	defer r.lock.Unlock()

	result := make(chan *Result, 1)
	r.results[task.String()] = result
	r.queue <- &enqueuedCommand{
		task: task,
		accept: func(res *Result) {
			result <- res
		},
	}
}

func (r *Runner) Results(task *Task) *Result {
	r.lock.Lock()
	result, ok := r.results[task.String()]
	r.lock.Unlock()
	if !ok {
		return nil
	}
	return <-result
}

func (r *Runner) Cancel() {
	r.ctxCancel()
	r.wg.Wait()
}
