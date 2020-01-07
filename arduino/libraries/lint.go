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
