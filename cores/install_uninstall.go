package cores

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bcmi-labs/arduino-cli/common"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/codeclysm/extract"
)

// DirPermissions is the default permission for create directories.
// respects umask on linux.
var DirPermissions os.FileMode = 0777

// Install installs a specific release of a core.
func Install(packager, arch string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the core")
	}

	arduinoFolder, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		return err
	}
	tempFolder := filepath.Join(arduinoFolder, "tmp", "packages",
		fmt.Sprintf("core-%d", time.Now().Unix()))

	coresFolder, err := common.GetDefaultCoresFolder(packager)
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	destCoresDirParent := filepath.Join(coresFolder, arch)
	destCoresDir := filepath.Join(destCoresDirParent, release.VersionName())

	defer func() {
		//cleaning empty directories
		if empty, _ := IsDirEmpty(destCoresDir); empty {
			os.RemoveAll(destCoresDir)
		}
		if empty, _ := IsDirEmpty(destCoresDirParent); empty {
			os.RemoveAll(destCoresDirParent)
		}
	}()

	err = os.MkdirAll(tempFolder, DirPermissions)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempFolder)

	err = os.MkdirAll(destCoresDirParent, DirPermissions)
	if err != nil {
		return err
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = extract.Archive(file, tempFolder, nil)
	if err != nil {
		return err
	}

	realDir := coreTempDir(tempFolder)
	if realDir == "invalid" {
		return errors.New("invalid archive structure")
	}

	err = os.Rename(realDir, destCoresDir)
	if err != nil {
		return err
	}

	return nil
}

// InstallTool installs a specific release of a tool.
func InstallTool(packager, name string, release releases.Release) error {
	if release == nil {
		return errors.New("Not existing version of the tool")
	}

	arduinoFolder, err := common.GetDefaultArduinoHomeFolder()
	if err != nil {
		return err
	}
	tempFolder := filepath.Join(arduinoFolder, "tmp", "tools",
		fmt.Sprintf("tool-%d", time.Now().Unix()))

	toolsFolder, err := common.GetDefaultToolsFolder(packager)
	if err != nil {
		return err
	}

	cacheFilePath, err := releases.ArchivePath(release)
	if err != nil {
		return err
	}

	destToolsDirParent := filepath.Join(toolsFolder, name)
	destToolsDir := filepath.Join(destToolsDirParent, release.VersionName())

	defer func() {
		//cleaning empty directories
		if empty, _ := IsDirEmpty(destToolsDir); empty {
			os.RemoveAll(destToolsDir)
		}
		if empty, _ := IsDirEmpty(destToolsDirParent); empty {
			os.RemoveAll(destToolsDirParent)
		}
	}()

	err = os.MkdirAll(tempFolder, DirPermissions)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempFolder)

	err = os.MkdirAll(destToolsDirParent, DirPermissions)
	if err != nil {
		return err
	}

	file, err := os.Open(cacheFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = extract.Archive(file, tempFolder, nil)
	if err != nil {
		return err
	}

	realDir := toolTempDir(tempFolder)
	if realDir == "invalid" {
		return errors.New("invalid archive structure")
	}

	err = os.Rename(realDir, destToolsDir)
	if err != nil {
		return err
	}

	return nil
}

// IsDirEmpty returns if the directory specified by path is empty,
// and an error if occurred.
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

func coreTempDir(tempDir string) string {
	realDir := "invalid"
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "platform.txt") {
			realDir = filepath.Dir(path)
			return errors.New("stopped, ok") //error put to stop the search of the root
		}
		return nil
	})
	return realDir
}

func toolTempDir(tempDir string) string {
	realDir := tempDir
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil //ignore this step
		}
		dir, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer dir.Close()
		_, err = dir.Readdir(3)
		if err == io.EOF { // read 3 files failed with EOF, dir has 2 files or more.
			realDir = path
			return errors.New("stopped, ok") //error put to stop the search of the root
		}
		return nil
	})
	return realDir
}
