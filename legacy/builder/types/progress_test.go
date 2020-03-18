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

package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProgress(t *testing.T) {
	p := &ProgressStruct{}
	p.AddSubSteps(3)
	require.Equal(t, float32(0.0), p.Progress)
	require.InEpsilon(t, 33.33333, p.StepAmount, 0.00001)
	fmt.Printf("%+v\n", p)
	{
		p.CompleteStep()
		require.InEpsilon(t, 33.33333, p.Progress, 0.00001)
		fmt.Printf("%+v\n", p)

		p.AddSubSteps(4)
		require.InEpsilon(t, 33.33333, p.Progress, 0.00001)
		require.InEpsilon(t, 8.33333, p.StepAmount, 0.00001)
		fmt.Printf("%+v\n", p)
		{
			p.CompleteStep()
			require.InEpsilon(t, 41.66666, p.Progress, 0.00001)
			fmt.Printf("%+v\n", p)

			p.CompleteStep()
			require.InEpsilon(t, 50.0, p.Progress, 0.00001)
			fmt.Printf("%+v\n", p)

			p.AddSubSteps(0) // zero steps
			fmt.Printf("%+v\n", p)
			{
				// p.CompleteStep() invalid here
			}
			p.RemoveSubSteps()
		}
		p.RemoveSubSteps()
		require.InEpsilon(t, 33.33333, p.Progress, 0.00001)
		fmt.Printf("%+v\n", p)

		p.CompleteStep()
		require.InEpsilon(t, 66.66666, p.Progress, 0.00001)
		fmt.Printf("%+v\n", p)
	}
	p.RemoveSubSteps()
	require.Equal(t, float32(0.0), p.Progress)
	require.Equal(t, float32(0.0), p.StepAmount)
}
