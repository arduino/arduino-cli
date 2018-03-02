/*
 * This file is part of arduino-cli.
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2017 ARDUINO AG (http://www.arduino.cc/)
 */

package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/bcmi-labs/arduino-cli/common/formatter/output"
	"github.com/bcmi-labs/arduino-cli/common/releases"
	"github.com/sirupsen/logrus"
)

type resultMessage struct {
	message interface{}
}

// MarshalJSON allows to marshal this object as a JSON object.
func (res resultMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"result": res.message,
	})
}

func (res resultMessage) String() string {
	return fmt.Sprint(res.message)
}

// PrintResult prints a value as a result from an operation.
func PrintResult(res interface{}) {
	Print(resultMessage{
		message: res,
	})
}

// ExtractProcessResultsFromDownloadResults picks a set of releases.DownloadResource and
// a set of releases.DownloadResult and creates an array of output.ProcessResult to format
// the download results. A label is added to each result.
// FIXME: I don't like this kind of result passing back and forth, it has to be a better way
func ExtractProcessResultsFromDownloadResults(
	resources map[string]*releases.DownloadResource,
	results map[string]*releases.DownloadResult,
	label string) map[string]output.ProcessResult {

	out := map[string]output.ProcessResult{}
	for name, resource := range resources {
		path, err := resource.ArchivePath()
		if err != nil {
			// FIXME: do something!!
			logrus.Error("Could not determine archive path:", err)
			continue
		}
		resultError := results[name].Error
		status := ""
		errorMessage := ""
		if resultError == nil {
			status = label
		} else {
			errorMessage = resultError.Error()
		}
		out[name] = output.ProcessResult{
			ItemName: name,
			Path:     path,
			Error:    errorMessage,
			Status:   status,
		}
	}
	return out
}
