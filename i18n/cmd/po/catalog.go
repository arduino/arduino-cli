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
	"fmt"
	"io"
	"sort"
	"strings"
)

type (
	// MessageCatalog is the catalog of i18n messages for a given locale
	MessageCatalog struct {
		Messages map[string]*Message
	}

	// Message represents a i18n message
	Message struct {
		Comments []string
		Value    string
	}
)

// Add adds a new message in the i18n catalog
func (catalog *MessageCatalog) Add(id, value string, comment []string) {
	if catalog.Messages == nil {
		catalog.Messages = map[string]*Message{}
	}

	if catalog.Messages[id] == nil {
		catalog.Messages[id] = &Message{Value: value}
	}

	if len(comment) != 0 {
		catalog.Messages[id].Comments = append(catalog.Messages[id].Comments, comment...)
	}
}

// AddMessage adds a new message in the i18n catalog
func (catalog *MessageCatalog) AddMessage(id string, message Message) {
	catalog.Add(id, message.Value, message.Comments)
}

// SortedKeys returns the sorted keys in the catalog
func (catalog *MessageCatalog) SortedKeys() []string {
	keys := []string{}
	for k := range catalog.Messages {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// Write writes the catalog in PO file format into w
func (catalog *MessageCatalog) Write(w io.Writer) {
	keys := []string{}
	for k := range catalog.Messages {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		msg := catalog.Messages[k]

		for _, comment := range msg.Comments {
			fmt.Fprintln(w, comment)
		}

		printValue(w, "msgid", k)
		printValue(w, "msgstr", msg.Value)
		fmt.Fprintln(w)
	}
}

func printValue(w io.Writer, field, value string) {
	if strings.Contains(value, "\n") {
		fmt.Fprintf(w, "%s ", field)
		lines := strings.Split(value, "\n")
		for i, line := range lines {
			if i == len(lines)-1 {
				fmt.Fprintf(w, "\"%s\"\n", escape(line))
			} else {
				fmt.Fprintf(w, "\"%s\\n\"\n", escape(line))
			}
		}
	} else {
		fmt.Fprintf(w, "%s \"%s\"\n", field, escape(value))
	}
}

func escape(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}
