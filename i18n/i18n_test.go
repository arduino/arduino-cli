package i18n

import (
	"testing"

	"github.com/leonelquinteros/gotext"
	"github.com/stretchr/testify/require"
)

func setPo(poFile string) {
	po = new(gotext.Po)
	po.Parse([]byte(poFile))
}

func TestPoTranslation(t *testing.T) {
	setPo(`
		msgid "test-key-ok"
		msgstr "test-key-translated"
	`)
	require.Equal(t, "test-key", Tr("test-key"))
	require.Equal(t, "test-key-translated", Tr("test-key-ok"))
}

func TestNoLocaleSet(t *testing.T) {
	po = new(gotext.Po)
	require.Equal(t, "test-key", Tr("test-key"))
}

func TestTranslationWithVariables(t *testing.T) {
	setPo(`
		msgid "test-key-ok %s"
		msgstr "test-key-translated %s"
	`)
	require.Equal(t, "test-key", Tr("test-key"))
	require.Equal(t, "test-key-translated message", Tr("test-key-ok %s", "message"))
}
