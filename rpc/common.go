package rpc

//go:generate protoc -I . -I .. --go_out=plugins=grpc:../../../.. common.proto

// Error codes to be used for os.Exit().
const (
	_          = iota // 0 is not a valid exit error code
	ErrGeneric        // 1 is the reserved "catchall" code in Unix
	_                 // 2 is reserved in Unix
	ErrNoConfigFile
	ErrBadCall
	ErrNetwork
	// ErrCoreConfig represents an error in the cli core config, for example some basic
	// files shipped with the installation are missing, or cannot create or get basic
	// directories vital for the CLI to work.
	ErrCoreConfig
	ErrBadArgument
)

func Error(message string, code int) *Result {
	return &Result{
		Code:    int32(code),
		Message: message,
		Failed:  true,
	}
}

func (r *Result) Error() string {
	return r.Message
}

// Success returns true if the Result is successful
func (r *Result) Success() bool {
	if r == nil {
		return true
	}
	return !r.Failed
}
