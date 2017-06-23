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
 * Copyright 2017 BCMI LABS SA (http://www.arduino.cc/)
 */

package prettyPrints

import (
	"math"

	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

func init() {
	log = logrus.WithFields(logrus.Fields{})
}

// A TaskWrapper wraps a task to be executed to allow
// Useful messages to be print. It is used to pretty
// print operations.
//
// All Message arrays use VERBOSITY as index.
type TaskWrapper struct {
	beforeMessage []string
	task          Task
	afterMessage  []string
	errorMessage  []string
}

// verbosity represents the verbosity level of the message.
//
// Examples:
//
// verbosity 0 Message ""
// verbosity 1 Message "Hi"
// verbosity 2 Message "Hello folk, how are you?"
// type verbosity int

// Task represents a function which can be safely wrapped into a TaskWrapper
type Task func() error

// ExecuteTask executes a task while printing messages to describe what is happening.
func (tw TaskWrapper) ExecuteTask(verb int) error {
	var maxUsableVerb int
	maxUsableVerb = minVerb(verb, tw.beforeMessage)
	log.Info(tw.beforeMessage[maxUsableVerb])
	err := tw.task()
	if err != nil {
		maxUsableVerb = minVerb(verb, tw.errorMessage)
		log.Warn(tw.errorMessage[maxUsableVerb])
	} else {
		maxUsableVerb = minVerb(verb, tw.afterMessage)
		log.Info(tw.afterMessage[maxUsableVerb])
	}
	return err
}

// minVerb tells which is the max level of verbosity for the specified verbosity level (set by another
// function call) and the provided array of strings.
//
// Refer to TaskRunner struct for the usage of the array.
func minVerb(verb1 int, sentences []string) int {
	return int(math.Min(float64(verb1), float64(len(sentences)-1)))
}
