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

// download downloads a library from arduino repository.
func download(library *Library) ([]byte, error) {
	client := http.DefaultClient

	request, err := http.NewRequest("GET", library.Latest.URL, nil)
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

// getDownloadCacheFolder gets the folder where temp installs are stored until installation complete (libraries).
func getDownloadCacheFolder(library *Library) (string, error) {
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return "", err
	}

	stagingFolder := filepath.Join(libFolder, "cache")
	_, err = os.Stat(stagingFolder)
	if err != nil {
		fmt.Print("Cannot find cache folder of libraries, attempting to create it ... ")
		err = os.MkdirAll(stagingFolder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create cache folder of libraries")
			return "", err
		}
		fmt.Println("OK")
	}
	return stagingFolder, nil
}

// getLibFolder returns the destination folder of the downloaded specified library.
// It creates the folder if does not find it.
func getLibFolder(library *Library) (string, error) {
	baseFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return "", err
	}

	libFolder := filepath.Join(baseFolder, fmt.Sprintf("%s-%s", library.Name, library.Latest.Version))

	_, err = os.Stat(libFolder)
	if os.IsNotExist(err) {
		fmt.Print("Cannot find lib folder, trying to create it ... ")
		err = os.MkdirAll(libFolder, 0755)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println("Cannot create lib folder")
			return "", err
		}
		fmt.Println("OK")
	}

	return libFolder, nil
}

// DownloadAndInstall downloads a library and installs it to its specified location.
func DownloadAndInstall(library *Library) error {
	libFolder, err := common.GetDefaultLibFolder()
	if err != nil {
		return fmt.Errorf("Cannot get Lib destination directory")
	}

	stagingFolder, err := getDownloadCacheFolder(library)
	if err != nil {
		return fmt.Errorf("Cannot get staging installs folder")
	}

	body, err := download(library)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(body)

	archive, err := zip.NewReader(reader, int64(reader.Len()))
	if err != nil {
		return fmt.Errorf("Cannot read downloaded archive")
	}

	// if I can read the archive I save it to staging folder.
	err = ioutil.WriteFile(filepath.Join(stagingFolder, fmt.Sprintf("%s-%s.zip", library.Name, library.Latest.Version)), body, 0666)
	if err != nil {
		return fmt.Errorf("Cannot write download to cache folder, %s", err.Error())
	}

	err = common.Unzip(archive, libFolder)
	if err != nil {
		return err
	}

	return nil
}
