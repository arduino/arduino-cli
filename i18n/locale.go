package i18n

//go:generate ./embed-i18n.sh

import (
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/leonelquinteros/gotext"
)

var (
	loadOnce sync.Once
	po       *gotext.Po
)

func init() {
	po = new(gotext.Po)
}

// SetLocale sets the locate used for i18n
func SetLocale(locale string) {
	box := rice.MustFindBox("./data")
	poFile, err := box.Bytes(locale + ".po")

	if err != nil {
		poFile = box.MustBytes("en.po")
	}

	po = new(gotext.Po)
	po.Parse(poFile)
}
