package po

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	r := strings.NewReader(`
msgid "a"
msgstr ""

msgid "b"
msgstr "value-b"
	`)
	catalogA := ParseReader(r)

	r = strings.NewReader(`
msgid "a"
msgstr "value-a"

msgid "b"
msgstr "value-b"

msgid "c"
msgstr "value-c"
	`)

	catalogB := ParseReader(r)

	mergedCatalog := Merge(catalogA, catalogB)

	var buf bytes.Buffer
	mergedCatalog.Write(&buf)

	require.Equal(t, `msgid "a"
msgstr "value-a"

msgid "b"
msgstr "value-b"

`, buf.String())

}
