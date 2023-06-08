// This file is part of arduino-cli.
//
// Copyright 2020 ARDUINO SA (http://www.arduino.cc/)
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

package types

import "golang.org/x/exp/slices"

type UniqueSourceFileQueue []*SourceFile

func (queue *UniqueSourceFileQueue) Push(value *SourceFile) {
	if !queue.Contains(value) {
		*queue = append(*queue, value)
	}
}

func (queue UniqueSourceFileQueue) Contains(target *SourceFile) bool {
	return slices.ContainsFunc(queue, target.Equals)
}

func (queue *UniqueSourceFileQueue) Pop() *SourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

func (queue UniqueSourceFileQueue) Empty() bool {
	return len(queue) == 0
}
