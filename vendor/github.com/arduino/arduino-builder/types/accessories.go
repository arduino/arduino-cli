/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
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
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package types

type UniqueStringQueue []string

func (queue UniqueStringQueue) Len() int           { return len(queue) }
func (queue UniqueStringQueue) Less(i, j int) bool { return false }
func (queue UniqueStringQueue) Swap(i, j int)      { panic("Who called me?!?") }

func (queue *UniqueStringQueue) Push(value string) {
	if !sliceContains(*queue, value) {
		*queue = append(*queue, value)
	}
}

func (queue *UniqueStringQueue) Pop() interface{} {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

func (queue *UniqueStringQueue) Empty() bool {
	return queue.Len() == 0
}

type UniqueSourceFileQueue []SourceFile

func (queue UniqueSourceFileQueue) Len() int           { return len(queue) }
func (queue UniqueSourceFileQueue) Less(i, j int) bool { return false }
func (queue UniqueSourceFileQueue) Swap(i, j int)      { panic("Who called me?!?") }

func (queue *UniqueSourceFileQueue) Push(value SourceFile) {
	if !sliceContainsSourceFile(*queue, value) {
		*queue = append(*queue, value)
	}
}

func (queue *UniqueSourceFileQueue) Pop() SourceFile {
	old := *queue
	x := old[0]
	*queue = old[1:]
	return x
}

func (queue *UniqueSourceFileQueue) Empty() bool {
	return queue.Len() == 0
}
