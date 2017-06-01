package libraries

import (
	"fmt"
	"strconv"

	"path/filepath"

	"strings"

	"github.com/arduino/arduino-cli/config"
	"github.com/arduino/arduino-cli/opts"
)

type libsCommand struct {
	List libsListCommand `command:"list" description:"List/search available libraries"`
}

/*
func (*libsCommand) Usage() string {
	return "[--board package:arch:board[:parameters]] [--pref name=value] [-v|--verbose] [--preserve-temp-files] <FILE.ino>"
}
*/

var options libsCommand

func init() {
	opts.Parser.AddCommand("libs", "Manages libraries", "alsdkmaslkdmaslkdm", &options)
}

type libsListCommand struct {
	Long []bool `long:"long" short:"l" description:"Output detailed information about libraries (use twice for even more verbose output)"`
}

func (*libsListCommand) Usage() string {
	return "[-l] [-l] [<libname>]"
}

func (r *Release) String() string {
	res := "  Release: " + r.Version + "\n"
	res += "    URL: " + r.URL + "\n"
	res += "    ArchiveFileName: " + r.ArchiveFileName + "\n"
	res += "    Size: " + strconv.Itoa(r.Size) + "\n"
	res += "    Checksum: " + r.Checksum + "\n"
	return res
}

func (l *Library) String() string {
	res := "Name: " + l.Name + "\n"
	res += "  Author: " + l.Author + "\n"
	res += "  Maintainer: " + l.Maintainer + "\n"
	res += "  Sentence: " + l.Sentence + "\n"
	res += "  Paragraph: " + l.Paragraph + "\n"
	res += "  Website: " + l.Website + "\n"
	res += "  Category: " + l.Category + "\n"
	res += "  Architecture: " + strings.Join(l.Architectures, ", ") + "\n"
	res += "  Types: " + strings.Join(l.Types, ", ") + "\n"
	res += "  Versions: " + strings.Join(l.Versions(), ", ") + "\n"
	return res
}

func (opts *libsListCommand) Execute(args []string) error {
	fmt.Println("libs list:", args)
	fmt.Println("long =", opts.Long)

	baseFolder, err := config.GetDefaultArduinoFolder()
	if err != nil {
		return fmt.Errorf("Could not determine data folder: %s", err)
	}

	libFile := filepath.Join(baseFolder, "library_index.json")
	index, err := LoadLibrariesIndex(libFile)
	if err != nil {
		return fmt.Errorf("Could not read library index: %s", err)
	}

	libraries, err := CreateStatusContextFromIndex(index, nil, nil)
	if err != nil {
		return fmt.Errorf("Could not synchronize library status: %s", err)
	}

	for _, name := range libraries.Names() {
		if len(opts.Long) > 0 {
			lib := libraries.Libraries[name]
			fmt.Print(lib)
			if len(opts.Long) > 1 {
				for _, r := range lib.Releases {
					fmt.Print(r)
				}
			}
			fmt.Println()
		} else {
			fmt.Println(name)
		}
	}
	return nil
}
