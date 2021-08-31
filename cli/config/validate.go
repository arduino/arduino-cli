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

package config

import (
	"fmt"
	"reflect"
)

var validMap = map[string]reflect.Kind{
	"board_manager.additional_urls": reflect.Slice,
	"daemon.port":                   reflect.String,
	"directories.data":              reflect.String,
	"directories.downloads":         reflect.String,
	"directories.user":              reflect.String,
	"library.enable_unsafe_install": reflect.Bool,
	"logging.file":                  reflect.String,
	"logging.format":                reflect.String,
	"logging.level":                 reflect.String,
	"sketch.always_export_binaries": reflect.Bool,
	"metrics.addr":                  reflect.String,
	"metrics.enabled":               reflect.Bool,
	"network.proxy":                 reflect.String,
	"network.user_agent_ext":        reflect.String,
	"output.no_color":               reflect.Bool,
	"updater.enable_notification":   reflect.Bool,
}

func typeOf(key string) (reflect.Kind, error) {
	t, ok := validMap[key]
	if !ok {
		return reflect.Invalid, fmt.Errorf(tr("Settings key doesn't exist"))
	}
	return t, nil
}
