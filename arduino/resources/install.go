package resources

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codeclysm/extract"
)

// Install installs the resource in three steps:
// - the archive is unpacked in a temporary subfolder of tempPath
// - there should be only one root folder in the unpacked content
// - the only root folder is moved/renamed to/as the destination directory
// Note that tempPath and destDir must be on the same filesystem partition
// otherwise the last step will fail.
func (release *DownloadResource) Install(tempPath string, destDir string) error {
	// Create a temporary folder to extract package
	if err := os.MkdirAll(tempPath, 0777); err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	tempDir, err := ioutil.TempDir(tempPath, "package-")
	if err != nil {
		return fmt.Errorf("creating temp dir for extraction: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Obtain the archive path and open it
	archivePath, err := release.ArchivePath()
	if err != nil {
		return fmt.Errorf("getting archive path: %s", err)
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive file: %s", err)
	}
	defer file.Close()

	// Extract into temp directory
	if err := extract.Archive(file, tempDir, nil); err != nil {
		return fmt.Errorf("extracting archive: %s", err)
	}

	// Check package content and find package root dir
	root, err := findPackageRoot(tempDir)
	if err != nil {
		return fmt.Errorf("searching package root dir: %s", err)
	}

	// Ensure container dir exists
	destDirParent := filepath.Dir(destDir)
	if err := os.MkdirAll(destDirParent, 0777); err != nil {
		return err
	}
	defer func() {
		if empty, err := IsDirEmpty(destDirParent); err == nil && empty {
			os.RemoveAll(destDirParent)
		}
	}()

	// Move/rename the extracted root directory in the destination directory
	if err := os.Rename(root, destDir); err != nil {
		return err
	}

	// Create a package file
	if err := createPackageFile(destDir); err != nil {
		return err
	}

	return nil
}

// IsDirEmpty returns true if the directory specified by path is empty.
func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	// read in ONLY one file
	_, err = f.Readdir(1)

	// and if the file is EOF... well, the dir is empty.
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func findPackageRoot(parent string) (string, error) {
	files, err := ioutil.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("reading package root dir: %s", err)
	}
	root := ""
	for _, fileInfo := range files {
		if !fileInfo.IsDir() {
			continue
		}
		if root == "" {
			root = fileInfo.Name()
		} else {
			return "", fmt.Errorf("no unique root dir in archive, found '%s' and '%s'", root, fileInfo.Name())
		}
	}
	return filepath.Join(parent, root), nil
}
