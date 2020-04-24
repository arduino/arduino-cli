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
		catalog.Messages[id].Comments = comment
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
		for _, line := range strings.Split(value, "\n") {
			fmt.Fprintf(w, "\"%s\"\n", strings.ReplaceAll(line, `"`, `\"`))
		}
	} else {
		fmt.Fprintf(w, "%s \"%s\"\n", field, strings.ReplaceAll(value, `"`, `\"`))
	}
}
