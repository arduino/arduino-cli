package i18n

import (
	"os"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/pkg/errors"
)

func ErrorfWithLogger(logger Logger, format string, a ...interface{}) error {
	if logger.Name() == "machine" {
		logger.Fprintln(os.Stderr, constants.LOG_LEVEL_ERROR, format, a...)
		return errors.New("")
	}
	return errors.New(Format(format, a...))
}
