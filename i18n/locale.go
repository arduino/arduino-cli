package i18n

//go:generate rice embed-go

import (
	"sync"

	rice "github.com/GeertJohan/go.rice"
	"github.com/leonelquinteros/gotext"
	"github.com/sirupsen/logrus"
)

var (
	loadOnce sync.Once
	po       *gotext.Po
)

func init() {
	po = new(gotext.Po)
}

func SetLocale(locale string) {
	box := rice.MustFindBox("./data")
	poFile, err := box.Bytes(locale + ".po")

	if err != nil {
		logrus.Warn("i18n not found for ", locale, " using en")
		poFile = box.MustBytes("en.po")
	}

	po = new(gotext.Po)
	po.Parse(poFile)
}
