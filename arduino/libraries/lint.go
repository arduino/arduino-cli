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
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 */

package libraries

// Lint produce warnings about the formal correctness of a Library
func (l *Library) Lint() ([]string, error) {

	// TODO: check for spurious dirs
	// subDirs, err := ioutil.ReadDir(libraryDir)
	// if err != nil {
	// 	return nil, fmt.Errorf("reading dir %s: %s", libraryDir, err)
	// }
	// 	for _, subDir := range subDirs {
	// 		if utils.IsSCCSOrHiddenFile(subDir) {
	// 			if !utils.IsSCCSFile(subDir) && utils.IsHiddenFile(subDir) {
	// 				logger.Fprintln(os.Stdout, "warn",
	// 					"WARNING: Spurious {0} directory in '{1}' library",
	// 					filepath.Base(subDir.Name()), libProperties["name"])
	// 			}
	// 		}
	// 	}

	// TODO: check for invalid category in library.properties
	// if !ValidCategories[libProperties["category"]] {
	//  logger.Fprintln(os.Stdout, "warn",
	// 	"WARNING: Category '{0}' in library {1} is not valid. Setting to '{2}'",
	// 	libProperties["category"], libProperties["name"], "Uncategorized")
	// 	libProperties["category"] = "Uncategorized"
	// }

	return []string{}, nil
}
