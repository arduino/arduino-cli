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

package common

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// GetFolder gets a folder on a path, and creates it if not found.
func GetFolder(folder string, messageName string) (string, error) {
	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		fmt.Printf("Cannot find default %s folder, attemping to create it ...", messageName)
		err = os.MkdirAll(folder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Printf("Cannot create %s folder\n", messageName)
			return "", err
		}
		fmt.Println("OK")
	}
	return folder, nil
}

// Unzip extracts a zip file to a specified destination path.
func Unzip(archive *zip.Reader, destination string) error {
	for _, file := range archive.File {
		path := filepath.Join(destination, file.Name)
		if file.FileInfo().IsDir() {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return fmt.Errorf("Cannot create directory during extraction. Process has been aborted")
			}
		} else {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return fmt.Errorf("Cannot create directory tree of file during extraction. Process has been aborted")
			}

			fileOpened, err := file.Open()
			if err != nil {
				return fmt.Errorf("Cannot open archived file, process has been aborted")
			}
			content, err := ioutil.ReadAll(fileOpened)
			if err != nil {
				return fmt.Errorf("Cannot read archived file, process has been aborted")
			}
			err = ioutil.WriteFile(path, content, 0664)
			if err != nil {
				return fmt.Errorf("Cannot copy archived file, process has been aborted, %s", err)
			}
		}
	}
	return nil
}

// TruncateDir removes all content from a directory, without deleting it.
// like `rm -rf dir/*`
func TruncateDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadPackage downloads a package from arduino repository.
func DownloadPackage(URL string) ([]byte, error) {
	client := http.DefaultClient

	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("Cannot create HTTP request")
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Cannot fetch library. Response creation error")
	} else if response.StatusCode != 200 {
		response.Body.Close()
		return nil, fmt.Errorf("Cannot fetch library. Source responded with a status %d code", response.StatusCode)
	}
	defer response.Body.Close()

	// Download completed, now move the archive to temp location and unpack it.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Cannot read response body")
	}
	return body, nil
}
