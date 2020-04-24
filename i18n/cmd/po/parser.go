package po

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func Parse(filename string) MessageCatalog {
	if !fileExists(filename) {
		return MessageCatalog{}
	}

	file, err := os.Open(filename)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
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
			id += "\n" + mustUnquote(line)
			continue
		}

		if strings.HasPrefix(line, "msgstr") {
			state = StateMessageValue
			value += mustUnquote(strings.TrimLeft(line, "msgstr "))
			continue
		}

		if state == StateMessageValue && strings.HasPrefix(line, "\"") {
			value += "\n" + mustUnquote(line)
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
