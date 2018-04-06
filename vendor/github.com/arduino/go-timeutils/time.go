/*
 * This file is part of Arduino Builder.
 *
 * Copyright 2016 Arduino LLC (http://www.arduino.cc/)
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
 */

package timeutils

import "time"

// TimezoneOffsetNoDST returns the timezone offset without the DST component
func TimezoneOffsetNoDST(t time.Time) int {
	_, winterOffset := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location()).Zone()
	_, summerOffset := time.Date(t.Year(), 7, 1, 0, 0, 0, 0, t.Location()).Zone()
	if winterOffset > summerOffset {
		winterOffset, summerOffset = summerOffset, winterOffset
	}
	return winterOffset
}

// DaylightSavingsOffset returns the DST offset of the specified time
func DaylightSavingsOffset(t time.Time) int {
	_, offset := t.Zone()
	return offset - TimezoneOffsetNoDST(t)
}

// LocalUnix returns the unix timestamp of the specified time with the
// local timezone offset and DST added
func LocalUnix(t time.Time) int64 {
	_, offset := t.Zone()
	return t.Unix() + int64(offset)
}
