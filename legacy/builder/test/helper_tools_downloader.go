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

package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/arduino/arduino-cli/legacy/builder/gohasissues"
	"github.com/arduino/arduino-cli/legacy/builder/utils"
	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-properties-orderedmap"
	"github.com/pkg/errors"
)

var hardwareFolder = paths.New("downloaded_hardware")
var boardManagerFolder = paths.New("downloaded_board_manager_stuff")
var toolsFolder = paths.New("downloaded_tools")
var librariesFolder = paths.New("downloaded_libraries")
var patchesFolder = paths.New("downloaded_stuff_patches")

type Tool struct {
	Name    string
	Package string
	Version string
	OsUrls  []OsUrl
}

type OsUrl struct {
	Os  string
	Url string
}

type Library struct {
	Name                   string
	Version                string
	VersionInLibProperties string
	Url                    string
}

type Core struct {
	Maintainer string
	Arch       string
	Version    string
	Url        string
}

func DownloadCoresAndToolsAndLibraries(t *testing.T) {
	cores := []Core{
		{Maintainer: "arduino", Arch: "avr", Version: "1.6.10"},
		{Maintainer: "arduino", Arch: "sam", Version: "1.6.7"},
	}

	boardsManagerCores := []Core{
		{Maintainer: "arduino", Arch: "samd", Version: "1.6.5"},
	}

	boardsManagerRedBearCores := []Core{
		{Maintainer: "RedBearLab", Arch: "avr", Version: "1.0.0", Url: "https://redbearlab.github.io/arduino/Blend/blend_boards.zip"},
	}

	toolsMultipleVersions := []Tool{
		{Name: "bossac", Version: "1.6.1-arduino"},
		{Name: "bossac", Version: "1.5-arduino"},
	}

	tools := []Tool{
		{Name: "avrdude", Version: "6.0.1-arduino5"},
		{Name: "avr-gcc", Version: "4.8.1-arduino5"},
		{Name: "arm-none-eabi-gcc", Version: "4.8.3-2014q1"},
		{Name: "ctags", Version: "5.8-arduino11",
			OsUrls: []OsUrl{
				{Os: "i686-pc-linux-gnu", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-i686-pc-linux-gnu.tar.bz2"},
				{Os: "x86_64-pc-linux-gnu", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-x86_64-pc-linux-gnu.tar.bz2"},
				{Os: "i686-mingw32", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-i686-mingw32.zip"},
				{Os: "x86_64-apple-darwin", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-x86_64-apple-darwin.zip"},
				{Os: "arm-linux-gnueabihf", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-armv6-linux-gnueabihf.tar.bz2"},
				{Os: "aarch64-linux-gnu", Url: "http://downloads.arduino.cc/tools/ctags-5.8-arduino11-aarch64-linux-gnu.tar.bz2"},
			},
		},
		{Name: "arduino-preprocessor", Version: "0.1.5",
			OsUrls: []OsUrl{
				{Os: "i686-pc-linux-gnu", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-i686-pc-linux-gnu.tar.bz2"},
				{Os: "x86_64-pc-linux-gnu", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-x86_64-pc-linux-gnu.tar.bz2"},
				{Os: "i686-mingw32", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-i686-w64-mingw32.tar.bz2"},
				{Os: "x86_64-apple-darwin", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-x86_64-apple-darwin11.tar.bz2"},
				{Os: "arm-linux-gnueabihf", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-arm-linux-gnueabihf.tar.bz2"},
				{Os: "aarch64-linux-gnu", Url: "https://github.com/arduino/arduino-preprocessor/releases/download/0.1.5/arduino-preprocessor-0.1.5-aarch64-linux-gnu.tar.bz2"},
			},
		},
	}

	boardsManagerTools := []Tool{
		{Name: "openocd", Version: "0.9.0-arduino", Package: "arduino"},
		{Name: "CMSIS", Version: "4.0.0-atmel", Package: "arduino"},
	}

	boardsManagerRFduinoTools := []Tool{
		{Name: "arm-none-eabi-gcc", Version: "4.8.3-2014q1", Package: "RFduino"},
	}

	libraries := []Library{
		{Name: "Audio", Version: "1.0.4"},
		{Name: "Adafruit PN532", Version: "1.0.0"},
		{Name: "Bridge", Version: "1.6.1"},
		{Name: "CapacitiveSensor", Version: "0.5.0", VersionInLibProperties: "0.5"},
		{Name: "Ethernet", Version: "1.1.1"},
		{Name: "Robot IR Remote", Version: "2.0.0"},
		{Name: "FastLED", Version: "3.1.0"},
	}

	download(t, cores, boardsManagerCores, boardsManagerRedBearCores, tools, toolsMultipleVersions, boardsManagerTools, boardsManagerRFduinoTools, libraries)

	patchFiles(t)
}

func patchFiles(t *testing.T) {
	err := patchesFolder.MkdirAll()
	NoError(t, err)
	files, err := patchesFolder.ReadDir()
	NoError(t, err)

	for _, file := range files {
		if file.Ext() == ".patch" {
			panic("Patching for downloaded tools is not available! (see https://github.com/arduino/arduino-builder/issues/147)")
			// XXX: find an alternative to x/codereview/patch
			// https://github.com/arduino/arduino-builder/issues/147
			/*
				data, err := ioutil.ReadFile(Abs(t, filepath.Join(PATCHES_FOLDER, file.Name())))
				NoError(t, err)
				patchSet, err := patch.Parse(data)
				NoError(t, err)
				operations, err := patchSet.Apply(ioutil.ReadFile)
				for _, op := range operations {
					utils.WriteFileBytes(op.Dst, op.Data)
				}
			*/
		}
	}
}

func download(t *testing.T, cores, boardsManagerCores, boardsManagerRedBearCores []Core, tools, toolsMultipleVersions, boardsManagerTools, boardsManagerRFduinoTools []Tool, libraries []Library) {
	allCoresDownloaded, err := allCoresAlreadyDownloadedAndUnpacked(hardwareFolder, cores)
	NoError(t, err)
	if allCoresDownloaded &&
		allBoardsManagerCoresAlreadyDownloadedAndUnpacked(boardManagerFolder, boardsManagerCores) &&
		allBoardsManagerCoresAlreadyDownloadedAndUnpacked(boardManagerFolder, boardsManagerRedBearCores) &&
		allBoardsManagerToolsAlreadyDownloadedAndUnpacked(boardManagerFolder, boardsManagerTools) &&
		allBoardsManagerToolsAlreadyDownloadedAndUnpacked(boardManagerFolder, boardsManagerRFduinoTools) &&
		allToolsAlreadyDownloadedAndUnpacked(toolsFolder, tools) &&
		allToolsAlreadyDownloadedAndUnpacked(toolsFolder, toolsMultipleVersions) &&
		allLibrariesAlreadyDownloadedAndUnpacked(librariesFolder, libraries) {
		return
	}

	index, err := downloadIndex("http://downloads.arduino.cc/packages/package_index.json")
	NoError(t, err)

	err = downloadCores(cores, index)
	NoError(t, err)

	err = downloadBoardManagerCores(boardsManagerCores, index)
	NoError(t, err)

	err = downloadTools(tools, index)
	NoError(t, err)

	err = downloadToolsMultipleVersions(toolsMultipleVersions, index)
	NoError(t, err)

	err = downloadBoardsManagerTools(boardsManagerTools, index)
	NoError(t, err)

	rfduinoIndex, err := downloadIndex("http://downloads.arduino.cc/packages/test_package_rfduino_index.json")
	NoError(t, err)

	err = downloadBoardsManagerTools(boardsManagerRFduinoTools, rfduinoIndex)
	NoError(t, err)

	err = downloadBoardManagerCores(boardsManagerRedBearCores, nil)
	NoError(t, err)

	librariesIndex, err := downloadIndex("http://downloads.arduino.cc/libraries/library_index.json")
	NoError(t, err)

	err = downloadLibraries(libraries, librariesIndex)
	NoError(t, err)
}

func downloadIndex(url string) (map[string]interface{}, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body.Close()

	index := make(map[string]interface{})
	err = json.Unmarshal(bytes, &index)
	if err != nil {
		return nil, err
	}

	return index, nil
}

func downloadCores(cores []Core, index map[string]interface{}) error {
	for _, core := range cores {
		url, err := findCoreUrl(index, core)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackCore(core, url, hardwareFolder)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func downloadBoardManagerCores(cores []Core, index map[string]interface{}) error {
	for _, core := range cores {
		url, err := findCoreUrl(index, core)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackBoardManagerCore(core, url, boardManagerFolder)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func findCoreUrl(index map[string]interface{}, core Core) (string, error) {
	if core.Url != "" {
		return core.Url, nil
	}
	packages := index["packages"].([]interface{})
	for _, p := range packages {
		pack := p.(map[string]interface{})
		if pack[constants.PACKAGE_NAME].(string) == core.Maintainer {
			packagePlatforms := pack["platforms"].([]interface{})
			for _, pt := range packagePlatforms {
				packagePlatform := pt.(map[string]interface{})
				if packagePlatform[constants.PLATFORM_ARCHITECTURE] == core.Arch && packagePlatform[constants.PLATFORM_VERSION] == core.Version {
					return packagePlatform[constants.PLATFORM_URL].(string), nil
				}
			}
		}
	}

	return constants.EMPTY_STRING, errors.Errorf("Unable to find tool " + core.Maintainer + " " + core.Arch + " " + core.Version)
}

func downloadTools(tools []Tool, index map[string]interface{}) error {
	host := translateGOOSGOARCHToPackageIndexValue()

	for _, tool := range tools {
		url, err := findToolUrl(index, tool, host)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackTool(tool, url, toolsFolder, true)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func downloadToolsMultipleVersions(tools []Tool, index map[string]interface{}) error {
	host := translateGOOSGOARCHToPackageIndexValue()

	for _, tool := range tools {
		if !toolAlreadyDownloadedAndUnpacked(toolsFolder, tool) {
			if toolsFolder.Join(tool.Name).Exist() {
				if err := toolsFolder.Join(tool.Name).RemoveAll(); err != nil {
					return errors.WithStack(err)
				}
			}
		}
	}

	for _, tool := range tools {
		url, err := findToolUrl(index, tool, host)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackTool(tool, url, toolsFolder, false)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func downloadBoardsManagerTools(tools []Tool, index map[string]interface{}) error {
	host := translateGOOSGOARCHToPackageIndexValue()

	for _, tool := range tools {
		url, err := findToolUrl(index, tool, host)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackBoardsManagerTool(tool, url, boardManagerFolder)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func allBoardsManagerCoresAlreadyDownloadedAndUnpacked(targetPath *paths.Path, cores []Core) bool {
	for _, core := range cores {
		if !boardManagerCoreAlreadyDownloadedAndUnpacked(targetPath, core) {
			return false
		}
	}
	return true
}

func boardManagerCoreAlreadyDownloadedAndUnpacked(targetPath *paths.Path, core Core) bool {
	return targetPath.Join(core.Maintainer, "hardware", core.Arch, core.Version).Exist()
}

func allCoresAlreadyDownloadedAndUnpacked(targetPath *paths.Path, cores []Core) (bool, error) {
	for _, core := range cores {
		alreadyDownloaded, err := coreAlreadyDownloadedAndUnpacked(targetPath, core)
		if err != nil {
			return false, errors.WithStack(err)
		}
		if !alreadyDownloaded {
			return false, nil
		}
	}
	return true, nil
}

func coreAlreadyDownloadedAndUnpacked(targetPath *paths.Path, core Core) (bool, error) {
	corePath := targetPath.Join(core.Maintainer, core.Arch)

	if corePath.NotExist() {
		return false, nil
	}
	platform, err := properties.LoadFromPath(corePath.Join("platform.txt"))
	if err != nil {
		return false, errors.WithStack(err)
	}

	if core.Version != platform.Get("version") {
		err := corePath.RemoveAll()
		return false, errors.WithStack(err)
	}

	return true, nil
}

func allBoardsManagerToolsAlreadyDownloadedAndUnpacked(targetPath *paths.Path, tools []Tool) bool {
	for _, tool := range tools {
		if !boardManagerToolAlreadyDownloadedAndUnpacked(targetPath, tool) {
			return false
		}
	}
	return true
}

func boardManagerToolAlreadyDownloadedAndUnpacked(targetPath *paths.Path, tool Tool) bool {
	return targetPath.Join(tool.Package, constants.FOLDER_TOOLS, tool.Name, tool.Version).Exist()
}

func allToolsAlreadyDownloadedAndUnpacked(targetPath *paths.Path, tools []Tool) bool {
	for _, tool := range tools {
		if !toolAlreadyDownloadedAndUnpacked(targetPath, tool) {
			return false
		}
	}
	return true
}

func toolAlreadyDownloadedAndUnpacked(targetPath *paths.Path, tool Tool) bool {
	return targetPath.Join(tool.Name, tool.Version).Exist()
}

func allLibrariesAlreadyDownloadedAndUnpacked(targetPath *paths.Path, libraries []Library) bool {
	for _, library := range libraries {
		if !libraryAlreadyDownloadedAndUnpacked(targetPath, library) {
			return false
		}
	}
	return true
}

func libraryAlreadyDownloadedAndUnpacked(targetPath *paths.Path, library Library) bool {
	libPath := targetPath.Join(strings.Replace(library.Name, " ", "_", -1))
	if !libPath.Exist() {
		return false
	}

	libProps, err := properties.LoadFromPath(libPath.Join("library.properties"))
	if err != nil {
		return false
	}
	return libProps.Get("version") == library.Version || libProps.Get("version") == library.VersionInLibProperties
}

func downloadAndUnpackCore(core Core, url string, targetPath *paths.Path) error {
	alreadyDownloaded, err := coreAlreadyDownloadedAndUnpacked(targetPath, core)
	if err != nil {
		return errors.WithStack(err)
	}
	if alreadyDownloaded {
		return nil
	}

	if err := targetPath.ToAbs(); err != nil {
		return errors.WithStack(err)
	}

	unpackFolder, files, err := downloadAndUnpack(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer unpackFolder.RemoveAll()

	packagerPath := targetPath.Join(core.Maintainer)
	corePath := targetPath.Join(core.Maintainer, core.Arch)

	if corePath.Exist() {
		if err := corePath.RemoveAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	if len(files) == 1 && files[0].IsDir() {
		if err := packagerPath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		err = copyRecursive(unpackFolder.Join(files[0].Name()), targetPath.Join(core.Maintainer, core.Arch))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := targetPath.Join(core.Maintainer, core.Arch).MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		for _, file := range files {
			err = copyRecursive(unpackFolder.Join(file.Name()), targetPath.Join(core.Maintainer, core.Arch, file.Name()))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func downloadAndUnpackBoardManagerCore(core Core, url string, targetPath *paths.Path) error {
	if boardManagerCoreAlreadyDownloadedAndUnpacked(targetPath, core) {
		return nil
	}

	if err := targetPath.ToAbs(); err != nil {
		return errors.WithStack(err)
	}

	unpackFolder, files, err := downloadAndUnpack(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer unpackFolder.RemoveAll()

	corePath := targetPath.Join(core.Maintainer, "hardware", core.Arch)
	if corePath.Exist() {
		if err := corePath.RemoveAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	if len(files) == 1 && files[0].IsDir() {
		if err := corePath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		err = copyRecursive(unpackFolder.Join(files[0].Name()), corePath.Join(core.Version))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := corePath.Join(core.Version).MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		for _, file := range files {
			err = copyRecursive(unpackFolder.Join(file.Name()), corePath.Join(core.Version, file.Name()))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func downloadAndUnpackBoardsManagerTool(tool Tool, url string, targetPath *paths.Path) error {
	if boardManagerToolAlreadyDownloadedAndUnpacked(targetPath, tool) {
		return nil
	}

	if err := targetPath.ToAbs(); err != nil {
		return errors.WithStack(err)
	}

	unpackFolder, files, err := downloadAndUnpack(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer unpackFolder.RemoveAll()

	if len(files) == 1 && files[0].IsDir() {
		if err := targetPath.Join(tool.Package, constants.FOLDER_TOOLS, tool.Name).MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		err = copyRecursive(unpackFolder.Join(files[0].Name()), targetPath.Join(tool.Package, constants.FOLDER_TOOLS, tool.Name, tool.Version))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := targetPath.Join(tool.Package, constants.FOLDER_TOOLS, tool.Name, tool.Version).MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		for _, file := range files {
			err = copyRecursive(unpackFolder.Join(file.Name()), targetPath.Join(tool.Package, constants.FOLDER_TOOLS, tool.Name, tool.Version, file.Name()))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func downloadAndUnpackTool(tool Tool, url string, targetPath *paths.Path, deleteIfMissing bool) error {
	if toolAlreadyDownloadedAndUnpacked(targetPath, tool) {
		return nil
	}

	if err := targetPath.ToAbs(); err != nil {
		return errors.WithStack(err)
	}

	unpackFolder, files, err := downloadAndUnpack(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer unpackFolder.RemoveAll()

	toolPath := targetPath.Join(tool.Name)
	if deleteIfMissing {
		if toolPath.Exist() {
			if err := toolPath.MkdirAll(); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	if len(files) == 1 && files[0].IsDir() {
		if err := toolPath.MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		err = copyRecursive(unpackFolder.Join(files[0].Name()), toolPath.Join(tool.Version))
		if err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := toolPath.Join(tool.Version).MkdirAll(); err != nil {
			return errors.WithStack(err)
		}
		for _, file := range files {
			err = copyRecursive(unpackFolder.Join(file.Name()), toolPath.Join(tool.Version, file.Name()))
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

func downloadAndUnpack(url string) (*paths.Path, []os.FileInfo, error) {
	fmt.Fprintln(os.Stderr, "Downloading "+url)

	unpackFolder, err := paths.MkTempDir("", "arduino-builder-tool")
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	urlParts := strings.Split(url, "/")
	archiveFileName := urlParts[len(urlParts)-1]
	archiveFilePath := unpackFolder.Join(archiveFileName)

	res, err := http.Get(url)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	res.Body.Close()

	archiveFilePath.WriteFile(bytes)

	cmd := buildUnpackCmd(archiveFilePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		return nil, nil, errors.WithStack(err)
	}
	if len(out) > 0 {
		fmt.Println(string(out))
	}

	archiveFilePath.Remove()

	files, err := gohasissues.ReadDir(unpackFolder.String())
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	return unpackFolder, files, nil
}

func buildUnpackCmd(file *paths.Path) *exec.Cmd {
	var cmd *exec.Cmd
	if file.Ext() == ".zip" {
		cmd = exec.Command("unzip", "-qq", file.Base())
	} else {
		cmd = exec.Command("tar", "xf", file.Base())
	}
	cmd.Dir = file.Parent().String()
	return cmd
}

func translateGOOSGOARCHToPackageIndexValue() []string {
	switch value := runtime.GOOS + "-" + runtime.GOARCH; value {
	case "linux-amd64":
		return []string{"x86_64-pc-linux-gnu", "x86_64-linux-gnu"}
	case "linux-386":
		return []string{"i686-pc-linux-gnu", "i686-linux-gnu"}
	case "windows-amd64":
		return []string{"i686-mingw32", "i686-cygwin"}
	case "windows-386":
		return []string{"i686-mingw32", "i686-cygwin"}
	case "darwin-amd64":
		return []string{"i386-apple-darwin11", "x86_64-apple-darwin"}
	case "linux-arm":
		return []string{"arm-linux-gnueabihf"}
	default:
		panic("Unknown OS: " + value)
	}
}

func findToolUrl(index map[string]interface{}, tool Tool, host []string) (string, error) {
	if len(tool.OsUrls) > 0 {
		for _, osUrl := range tool.OsUrls {
			if utils.SliceContains(host, osUrl.Os) {
				return osUrl.Url, nil
			}
		}
	} else {
		packages := index["packages"].([]interface{})
		for _, p := range packages {
			pack := p.(map[string]interface{})
			packageTools := pack[constants.PACKAGE_TOOLS].([]interface{})
			for _, pt := range packageTools {
				packageTool := pt.(map[string]interface{})
				name := packageTool[constants.TOOL_NAME].(string)
				version := packageTool[constants.TOOL_VERSION].(string)
				if name == tool.Name && version == tool.Version {
					systems := packageTool["systems"].([]interface{})
					for _, s := range systems {
						system := s.(map[string]interface{})
						if utils.SliceContains(host, system["host"].(string)) {
							return system[constants.TOOL_URL].(string), nil
						}
					}
				}
			}
		}
	}

	return constants.EMPTY_STRING, errors.Errorf("Unable to find tool " + tool.Name + " " + tool.Version)
}

func downloadLibraries(libraries []Library, index map[string]interface{}) error {
	for _, library := range libraries {
		url, err := findLibraryUrl(index, library)
		if err != nil {
			return errors.WithStack(err)
		}
		err = downloadAndUnpackLibrary(library, url, librariesFolder)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func findLibraryUrl(index map[string]interface{}, library Library) (string, error) {
	if library.Url != "" {
		return library.Url, nil
	}
	libs := index["libraries"].([]interface{})
	for _, l := range libs {
		lib := l.(map[string]interface{})
		if library.Name == lib["name"].(string) && library.Version == lib["version"].(string) {
			return lib["url"].(string), nil
		}
	}

	return constants.EMPTY_STRING, errors.Errorf("Unable to find library " + library.Name + " " + library.Version)
}

func downloadAndUnpackLibrary(library Library, url string, targetPath *paths.Path) error {
	if libraryAlreadyDownloadedAndUnpacked(targetPath, library) {
		return nil
	}

	if err := targetPath.ToAbs(); err != nil {
		return errors.WithStack(err)
	}

	unpackFolder, files, err := downloadAndUnpack(url)
	if err != nil {
		return errors.WithStack(err)
	}
	defer unpackFolder.RemoveAll()

	libPath := targetPath.Join(strings.Replace(library.Name, " ", "_", -1))
	if libPath.Exist() {
		if err := libPath.RemoveAll(); err != nil {
			return errors.WithStack(err)
		}
	}

	err = copyRecursive(unpackFolder.Join(files[0].Name()), libPath)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func copyRecursive(from, to *paths.Path) error {
	copyFunc := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(from.String(), currentPath)
		if err != nil {
			return errors.WithStack(err)
		}
		targetPath := filepath.Join(to.String(), rel)
		if info.IsDir() {
			err := os.MkdirAll(targetPath, info.Mode())
			if err != nil {
				return errors.WithStack(err)
			}
		} else if info.Mode().IsRegular() {
			fromFile, err := os.Open(currentPath)
			if err != nil {
				return errors.WithStack(err)
			}
			defer fromFile.Close()
			targetFile, err := os.Create(targetPath)
			if err != nil {
				return errors.WithStack(err)
			}
			defer targetFile.Close()
			_, err = io.Copy(targetFile, fromFile)
			if err != nil {
				return errors.WithStack(err)
			}
			err = os.Chmod(targetPath, info.Mode())
			if err != nil {
				return errors.WithStack(err)
			}
		} else if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkedFile, err := os.Readlink(currentPath)
			if err != nil {
				return errors.WithStack(err)
			}
			fromFile := filepath.Join(filepath.Dir(targetPath), linkedFile)
			err = os.Symlink(fromFile, targetPath)
			if err != nil {
				return errors.WithStack(err)
			}
		} else {
			return errors.Errorf("unable to copy file " + currentPath)
		}

		return nil
	}
	err := gohasissues.Walk(from.String(), copyFunc)
	return errors.WithStack(err)
}
