/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package task

import (
	"fmt"
	"sync"

	"github.com/bcmi-labs/arduino-cli/cmd/formatter"
)

// resultWithKey values are used by ExecuteParallelFromMap as temporary values.
type resultWithKey struct {
	Result Result
	Key    string
}

// CreateSequence returns a task to execute parameter tasks in sequence.
//
// if abortOnFailure = true then the sequence is aborted with the error,
// otherwise there is just an error logged.
func CreateSequence(taskWrappers []Wrapper, ignoreOnFailure []bool, verbosity int) Sequence {
	results := make([]Result, 0, 10)

	return Sequence(func() []Result {
		for i, taskWrapper := range taskWrappers {
			result := taskWrapper.Execute(verbosity)
			results = append(results, result)
			if result.Error != nil && !ignoreOnFailure[i] {
				formatter.Print(fmt.Sprintf("Warning from task %d: %s", i, result.Error))
			}
		}
		return results
	})
}

// Task creates a Task from a Sequence, by putting it []Result into
// a Result.Result. To access this result it should be done like this
// `result = []Result(ts.Task().Execute(verbosity).Result)`
func (ts Sequence) Task() Task {
	return (func() Result {
		return Result{
			Result: ts(),
			Error:  nil,
		}
	})
}

// ExecuteSequence creates and executes an array of tasks in strict sequence.
func ExecuteSequence(taskWrappers []Wrapper, ignoreOnFailure []bool, verbosity int) []Result {
	return CreateSequence(taskWrappers, ignoreOnFailure, verbosity)()
}

// ExecuteParallelFromMap executes a set of taskwrappers in parallel, taking input from a map[string]Wrapper.
func ExecuteParallelFromMap(taskMap map[string]Wrapper, verbosity int) map[string]Result {
	results := make(chan resultWithKey, len(taskMap))

	var wg sync.WaitGroup
	wg.Add(len(taskMap))

	for key, task := range taskMap {
		go func(key string, task Wrapper) {
			defer wg.Done()
			results <- resultWithKey{
				Key: key,
				Result: func() Result {
					return task.Execute(verbosity)
				}(),
			}
		}(key, task)
	}
	wg.Wait()
	close(results)
	mapResult := make(map[string]Result, len(results))
	for result := range results {
		//logrus.Errorf("results : %v %v\n", result.Key, result.Result)
		mapResult[result.Key] = result.Result
	}
	return mapResult
}

// ExecuteParallel executes a set of Wrappers in parallel, handling concurrency for results.
func ExecuteParallel(taskWrappers []Wrapper, verbosity int) []Result {
	results := make(chan Result, len(taskWrappers))

	var wg sync.WaitGroup
	wg.Add(len(taskWrappers))

	for _, task := range taskWrappers {
		go func(task Wrapper) {
			defer wg.Done()
			results <- func() Result {
				return task.Execute(verbosity)
			}()
		}(task)
	}
	wg.Wait()
	close(results)
	array := make([]Result, len(results))
	for i := range array {
		array[i] = <-results
	}
	return array
}
