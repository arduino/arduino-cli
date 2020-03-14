package i18n

import (
	"os"

	"github.com/arduino/arduino-cli/legacy/builder/constants"
	"github.com/go-errors/errors"
)

func ErrorfWithLogger(logger Logger, format string, a ...interface{}) *errors.Error {
	if logger.Name() == "machine" {
		logger.Fprintln(os.Stderr, constants.LOG_LEVEL_ERROR, format, a...)
		return errors.Errorf("")
	}
	return errors.Errorf(Format(format, a...))
}
