// This file is part of arduino-cli.
//
// Copyright 2023 ARDUINO SA (http://www.arduino.cc/)
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

package progress

// Struct fixdoc
type Struct struct {
	Progress   float32
	StepAmount float32
	Parent     *Struct
}

// AddSubSteps fixdoc
func (p *Struct) AddSubSteps(steps int) {
	p.Parent = &Struct{
		Progress:   p.Progress,
		StepAmount: p.StepAmount,
		Parent:     p.Parent,
	}
	if p.StepAmount == 0.0 {
		p.StepAmount = 100.0
	}
	p.StepAmount /= float32(steps)
}

// RemoveSubSteps fixdoc
func (p *Struct) RemoveSubSteps() {
	p.Progress = p.Parent.Progress
	p.StepAmount = p.Parent.StepAmount
	p.Parent = p.Parent.Parent
}

// CompleteStep fixdoc
func (p *Struct) CompleteStep() {
	p.Progress += p.StepAmount
}
