package libraries

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bcmi-labs/arduino-cli/common"
)

// DownloadAndInstall downloads a library and installs it to its specified location.
func DownloadAndInstall(library *Library) error {
	client := http.DefaultClient
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		fmt.Print("Cannot find lib folder, trying to create it ... ")
		baseFolder, err := common.GetDefaultArduinoFolder()
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create libraries folder, no current user defined")
			return fmt.Errorf("Cannot create libraries folder")
		}
		libFolder = filepath.Join(baseFolder, "Arduino", "libraries")
		err = os.MkdirAll(libFolder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create libraries folder")
			return fmt.Errorf("Cannot create libraries folder")
		}
		fmt.Println("OK")
	}

	request, err := http.NewRequest("GET", library.Latest.URL, nil)
	if err != nil {
		return fmt.Errorf("Cannot create HTTP request")
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Cannot fetch library. Response creation error")
	} else if response.StatusCode != 200 {
		response.Body.Close()
		return fmt.Errorf("Cannot fetch library. Source responded with a status %d code", response.StatusCode)
	}
	defer response.Body.Close()

	// Download completed, now move the archive to temp location and unpack it.
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Cannot read response body")
	}

	reader := bytes.NewReader(body)

	archive, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("Cannot read downloaded archive")
	}

	zipExtractionPath := filepath.Join(libFolder, fmt.Sprintf("%s-%s", library.Name, library.Latest.Version))

	err = os.MkdirAll(zipExtractionPath, 0755)
	if err != nil {
		return fmt.Errorf("Cannot create library final destination directory")
	}

	for _, file := range archive.File {
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(filepath.Join(zipExtractionPath, file.Name), 0755)
			if err != nil {
				return fmt.Errorf("Cannot create directory during extraction. Process has been aborted")
			}
		} else {

		}
		fileOpened, err := file.Open()
		if err != nil {
			return fmt.Errorf("Cannot open archived file, process has been aborted")
		}
		content, err := ioutil.ReadAll(fileOpened)
		if err != nil {
			return fmt.Errorf("Cannot read archived file, process has been aborted")
		}
		err = ioutil.WriteFile(filepath.Join(zipExtractionPath, file.Name), content, 0666)
		if err != nil {
			return fmt.Errorf("Cannot copy archived file, process has been aborted")
		}
	}

	return nil
}
