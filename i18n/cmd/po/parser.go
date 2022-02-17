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

package po

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Parse parses the PO file into a MessageCatalog
func Parse(filename string) MessageCatalog {
	if !fileExists(filename) {
		return MessageCatalog{}
	}

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return ParseReader(file)
}

// ParseReader parses the PO file into a MessageCatalog
func ParseReader(r io.Reader) MessageCatalog {
	scanner := bufio.NewScanner(r)
	return parseCatalog(scanner)
}

func parseCatalog(scanner *bufio.Scanner) MessageCatalog {
	const (
		StateWhitespace   = 0
		StateComment      = 1
		StateMessageID    = 2
		StateMessageValue = 3
	)

	state := StateWhitespace
	catalog := MessageCatalog{}
	comments := []string{}
	id := ""
	value := ""
	for {
		more := scanner.Scan()

		if !more {
			if state != StateWhitespace {
				catalog.Add(id, value, comments)
			}
			break
		}

		line := scanner.Text()

		if state == StateWhitespace && strings.TrimSpace(line) == "" {
			continue
		} else if state != StateWhitespace && strings.TrimSpace(line) == "" {
			catalog.Add(id, value, comments)
			state = StateWhitespace
			id = ""
			value = ""
			comments = []string{}
			continue
		}

		if strings.HasPrefix(line, "#") {
			state = StateComment
			comments = append(comments, line)
			continue
		}

		if strings.HasPrefix(line, "msgid") {
			state = StateMessageID
			id += mustUnquote(strings.TrimLeft(line, "msgid "))
			continue
		}

		if state == StateMessageID && strings.HasPrefix(line, "\"") {
			id += mustUnquote(line)
			continue
		}

		if strings.HasPrefix(line, "msgstr") {
			state = StateMessageValue
			value += mustUnquote(strings.TrimLeft(line, "msgstr "))
			continue
		}

		if state == StateMessageValue && strings.HasPrefix(line, "\"") {
			value += mustUnquote(line)
			continue
		}
	}

	return catalog
}

func mustUnquote(line string) string {
	v, err := strconv.Unquote(line)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return v
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
