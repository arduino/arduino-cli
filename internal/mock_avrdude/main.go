//
// This file is part arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/arduino/go-paths-helper"
)

func main() {
	tmp, err := paths.MkTempFile(nil, "test")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tmp.Close()
	tmpPath := paths.New(tmp.Name())

	fmt.Println("CHECKFILE:", tmpPath)

	// Keep the check file "alive" by periodically updating its modification time.
	// A test can detect that this process has been killed by observing that the
	// modification time stops advancing. The loop is capped to 30s so that a
	// leaked process eventually exits on its own instead of running forever.
	for range 150 {
		if err := tmpPath.WriteFile(fmt.Appendf(nil, "%d", time.Now().UnixNano())); err != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Remove the check file on normal completion.
	tmpPath.Remove()

	fmt.Println("COMPLETED")
}
