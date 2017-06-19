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
