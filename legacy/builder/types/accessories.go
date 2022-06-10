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

func (queue UniqueSourceFileQueue) Len() int           { return len(queue) }
func (queue UniqueSourceFileQueue) Less(i, j int) bool { return false }
func (queue UniqueSourceFileQueue) Swap(i, j int)      { panic("Who called me?!?") }

func (queue *UniqueSourceFileQueue) Push(value *SourceFile) {
	equals := func(elem *SourceFile) bool {
		return elem.Origin == value.Origin && elem.RelativePath.EqualsTo(value.RelativePath)
	}
	if !slices.ContainsFunc(*queue, equals) {
		*queue = append(*queue, value)
	}
}

func (queue *UniqueSourceFileQueue) Pop() *SourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

func (queue *UniqueSourceFileQueue) Empty() bool {
	return queue.Len() == 0
}
