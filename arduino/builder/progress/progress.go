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
