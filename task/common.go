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

package task

import (
	"math"

	"github.com/sirupsen/logrus"
)

// Task represents a function which can be safely wrapped into a Wrapper.
//
// It may provide a result but always provides an error.
type Task func() Result

// A Wrapper wraps a task to be executed to allow
// Useful messages to be print. It is used to pretty
// print operations.
//
// All Message arrays use VERBOSITY as index.
type Wrapper struct {
	BeforeMessage []string
	Task          Task
	AfterMessage  []string
	ErrorMessage  []string
}

//Result represents a result from a task, or an error.
type Result struct {
	Result interface{}
	Error  error
}

//Sequence represents a sequence of tasks.
type Sequence func() []Result

// Execute executes a task while printing messages to describe what is happening.
func (tw Wrapper) Execute(verb int) Result {
	var maxUsableVerb [3]int
	var msg string
	if tw.BeforeMessage != nil && len(tw.BeforeMessage) > 0 {
		maxUsableVerb[0] = minVerb(verb, tw.BeforeMessage)
		msg = tw.BeforeMessage[maxUsableVerb[0]]
		if msg != "" {
			logrus.Infof("%s ... ", msg)
		}
	}

	ret := tw.Task()

	if ret.Error != nil {
		if tw.ErrorMessage != nil && len(tw.ErrorMessage) > 0 {
			maxUsableVerb[1] = minVerb(verb, tw.ErrorMessage)
			msg = tw.ErrorMessage[maxUsableVerb[1]]
			if tw.BeforeMessage[maxUsableVerb[0]] != "" {
				logrus.Warn("ERROR\n")
			}
			if msg != "" {
				logrus.Warnf("%s\n", msg)
			}
		}
	} else if tw.AfterMessage != nil && len(tw.AfterMessage) > 0 {
		maxUsableVerb[2] = minVerb(verb, tw.AfterMessage)
		msg = tw.AfterMessage[maxUsableVerb[2]]
		if tw.BeforeMessage[maxUsableVerb[0]] != "" {
			logrus.Info("OK\n")
		}
		if msg != "" {
			logrus.Infof("%s\n", msg)
		}
	}
	return ret
}

// minVerb tells which is the max level of verbosity for the specified verbosity level (set by another
// function call) and the provided array of strings.
//
// Refer to Runner struct for the usage of the array.
func minVerb(verb1 int, sentences []string) int {
	return int(math.Min(float64(verb1), float64(len(sentences)-1)))
}
