package cmd

const (
	errOk           int = iota
	errNoConfigFile int = iota
	errBadCall      int = iota
	errGeneric      int = iota
	errNetwork      int = iota
	errCoreConfig   int = iota // Represents an error in the cli core config, for example some basic files shipped with the installation are missing, or cannot create or get basic folder vital for the CLI to work.
)
