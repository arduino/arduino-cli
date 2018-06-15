/*
 * This file is part of PathsHelper library.
 *
 * Copyright 2018 Arduino AG (http://www.arduino.cc/)
 *
 * PropertiesMap library is free software; you can redistribute it and/or modify
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

package paths

import (
	"io/ioutil"
	"os"
	"runtime"
)

// NullPath return the path to the /dev/null equivalent for the current OS
func NullPath() *Path {
	if runtime.GOOS == "windows" {
		return New("nul")
	}
	return New("/dev/null")
}

// TempDir returns the default path to use for temporary files
func TempDir() *Path {
	return New(os.TempDir())
}

// MkTempDir creates a new temporary directory in the directory
// dir with a name beginning with prefix and returns the path of
// the new directory. If dir is the empty string, TempDir uses the
// default directory for temporary files
func MkTempDir(dir, prefix string) (*Path, error) {
	path, err := ioutil.TempDir(dir, prefix)
	if err != nil {
		return nil, err
	}
	return New(path), nil
}

// Getwd returns a rooted path name corresponding to the current
// directory.
func Getwd() (*Path, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return New(wd), nil
}
