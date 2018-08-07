/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package formatter

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cavaliercoder/grab"
)

// JSONFormatter is a Formatter that output JSON objects.
// Intermediate results or interactive messages are ignored.
type JSONFormatter struct {
	Debug bool // if false, errors are not shown. Unparsable inputs are skipped. Otherwise an error message is shown.
}

// Format implements Formatter interface
func (jf *JSONFormatter) Format(msg interface{}) (string, error) {
	t := reflect.TypeOf(msg).Kind().String()
	if t == "ptr" {
		t = reflect.Indirect(reflect.ValueOf(msg)).Kind().String()
	}
	switch t {
	case "struct", "map":
		ret, err := json.Marshal(msg)
		return string(ret), err
	default:
		return "", fmt.Errorf("%s ignored", t)
	}
}

// DownloadProgressBar implements Formatter interface
func (jf *JSONFormatter) DownloadProgressBar(resp *grab.Response, prefix string) {
	resp.Wait()
}
