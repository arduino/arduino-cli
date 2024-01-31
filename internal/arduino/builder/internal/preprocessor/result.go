package preprocessor

type Result struct {
	args   []string
	stdout []byte
	stderr []byte
}

func (r Result) Args() []string {
	return r.args
}

func (r Result) Stdout() []byte {
	return r.stdout
}

func (r Result) Stderr() []byte {
	return r.stderr
}
