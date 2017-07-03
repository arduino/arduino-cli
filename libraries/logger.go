package libraries

import (
	lcf "github.com/Robpol86/logrus-custom-formatter"
	"github.com/sirupsen/logrus"
)

var log *logrus.Entry

func init() {
	log = logrus.WithFields(logrus.Fields{})
	logrus.SetFormatter(lcf.NewFormatter("%[message]s", nil))
}
