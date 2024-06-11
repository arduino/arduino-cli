// This file is part of arduino-cli.
//
// Copyright 2024 ARDUINO SA (http://www.arduino.cc/)
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

package runner_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arduino/arduino-cli/internal/arduino/builder/internal/runner"
	"github.com/stretchr/testify/require"
)

func TestRunMultipleTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r := runner.New(ctx, 0)
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 1 ; echo -n 0"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 2 ; echo -n 1"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 3 ; echo -n 2"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 4 ; echo -n 3"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 5 ; echo -n 4"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 6 ; echo -n 5"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 7 ; echo -n 6"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 8 ; echo -n 7"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 9 ; echo -n 8"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 10 ; echo -n 9"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 11 ; echo -n 10"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 12 ; echo -n 11"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 13 ; echo -n 12"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 14 ; echo -n 13"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 15 ; echo -n 14"))
	r.Enqueue(runner.NewTask("bash", "-c", "sleep 16 ; echo -n 15"))
	require.Nil(t, r.Results(runner.NewTask("bash", "-c", "echo -n 5")))
	fmt.Println(string(r.Results(runner.NewTask("bash", "-c", "sleep 3 ; echo -n 2")).Stdout))
	fmt.Println("Cancelling")
	r.Cancel()
	fmt.Println("Runner completed")
}
