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

package i18n

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/leonelquinteros/gotext"
	"github.com/stretchr/testify/require"
)

func setPo(poFile string) {
	po = gotext.NewPo()
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
	po = gotext.NewPo()
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

func TestTranslationInTemplate(t *testing.T) {
	setPo(`
		msgid "test-key"
		msgstr "test-key-translated %s"
	`)

	tpl, err := template.New("test-template").Funcs(template.FuncMap{
		"tr": Tr,
	}).Parse(`{{ tr "test-key" .Value }}`)
	require.NoError(t, err)

	data := struct {
		Value string
	}{
		"value",
	}
	var buf bytes.Buffer
	require.NoError(t, tpl.Execute(&buf, data))

	require.Equal(t, "test-key-translated value", buf.String())
}

func TestTranslationWithQuotedStrings(t *testing.T) {
	setPo(`
		msgid "test-key \"quoted\""
		msgstr "test-key-translated"
	`)

	require.Equal(t, "test-key-translated", Tr("test-key \"quoted\""))
	require.Equal(t, "test-key-translated", Tr(`test-key "quoted"`))
}

func TestTranslationWithLineBreaks(t *testing.T) {
	setPo(`
		msgid "test-key \"quoted\"\n"
		"new line"
		msgstr "test-key-translated"
	`)

	require.Equal(t, "test-key-translated", Tr("test-key \"quoted\"\nnew line"))
	require.Equal(t, "test-key-translated", Tr(`test-key "quoted"
new line`))
}
