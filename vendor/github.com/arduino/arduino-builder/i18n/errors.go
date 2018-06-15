package i18n

import "github.com/arduino/arduino-builder/constants"
import "github.com/go-errors/errors"
import "os"

func ErrorfWithLogger(logger Logger, format string, a ...interface{}) *errors.Error {
	if logger.Name() == "machine" {
		logger.Fprintln(os.Stderr, constants.LOG_LEVEL_ERROR, format, a...)
		return errors.Errorf("")
	}
	return errors.Errorf(Format(format, a...))
}

func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return errors.Wrap(err, 0)
}

func UnwrapError(err error) error {
	// Perhaps go-errors can do this already in later versions?
	// See https://github.com/go-errors/errors/issues/14
	switch e := err.(type) {
		case *errors.Error:
			return e.Err
		default:
			return err
	}
}
