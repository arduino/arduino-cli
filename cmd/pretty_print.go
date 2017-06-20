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

package cmd

import (
	"fmt"
	"strings"

	"github.com/bcmi-labs/arduino-cli/libraries"
)

func prettyPrintStatus(status *libraries.StatusContext) {
	//Pretty print libraries from index.
	for _, name := range status.Names() {
		if GlobalFlags.Verbose > 0 {
			lib := status.Libraries[name]
			fmt.Print(lib)
			if GlobalFlags.Verbose > 1 {
				for _, r := range lib.Releases {
					fmt.Print(r)
				}
			}
			fmt.Println()
		} else {
			fmt.Println(name)
		}
	}
}

func prettyPrintDownloadFileIndex() error {
	if GlobalFlags.Verbose > 0 {
		fmt.Print("Downloading a new index file from download.arduino.cc ... ")
	}

	err := libraries.DownloadLibrariesFile()
	if err != nil {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("ERROR")
		}
		fmt.Println("Cannot download index file, check your network connection.")
		return err
	}

	if GlobalFlags.Verbose > 0 {
		fmt.Println("OK")
	}

	return nil
}

func prettyPrintInstall(libraryOK []string, libraryFails map[string]string) {
	if len(libraryFails) > 0 {
		if len(libraryOK) > 0 {
			fmt.Println("The following libraries were succesfully installed:")
			fmt.Println(strings.Join(libraryOK, " "))
			fmt.Print("However, t")
		} else { //UGLYYYY but it works
			fmt.Print("T")
		}
		fmt.Println("he installation process failed on the following libraries:")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("All libraries successfully installed")
	}
}

//TODO: remove copypasting from prettyPrintInstall and merge them in a single function
func prettyPrintDownload(libraryOK []string, libraryFails map[string]string) {
	if len(libraryFails) > 0 {
		if len(libraryOK) > 0 {
			fmt.Println("The following libraries were succesfully downloaded:")
			fmt.Println(strings.Join(libraryOK, " "))
			fmt.Print("However, t")
		} else { //UGLYYYY but it works
			fmt.Print("T")
		}
		fmt.Println("he download of the following libraries failed:")
		for library, failure := range libraryFails {
			fmt.Printf("%s - %s\n", library, failure)
		}
	} else {
		fmt.Println("All libraries successfully downloaded")
	}
}

func prettyPrintCorruptedIndexFix(index *libraries.Index) (*libraries.StatusContext, error) {
	if GlobalFlags.Verbose > 0 {
		fmt.Println("Cannot parse index file, it may be corrupted.")
	}

	err := prettyPrintDownloadFileIndex()
	if err != nil {
		return nil, err
	}

	return prettyIndexParse(index)
}

func prettyIndexParse(index *libraries.Index) (*libraries.StatusContext, error) {
	if GlobalFlags.Verbose > 0 {
		fmt.Print("Parsing downloaded index file ... ")
	}

	//after download, I retry.
	status, err := libraries.CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		if GlobalFlags.Verbose > 0 {
			fmt.Println("ERROR")
		}
		fmt.Println("Cannot parse index file")
		return nil, err
	}
	if GlobalFlags.Verbose > 0 {
		fmt.Println("OK")
	}

	return status, nil
}
