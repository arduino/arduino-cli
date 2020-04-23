package httpclient

import (
	"net/http"
)

// New returns a default http client for use in the cli API calls
func New() (*http.Client, error) {
	config, err := DefaultConfig()

	if err != nil {
		return nil, err
	}

	return NewWithConfig(config), nil
}

// NewWithConfig creates a http client for use in the cli API calls with a given configuration
func NewWithConfig(config *Config) *http.Client {
	transport := newHTTPClientTransport(config)

	return &http.Client{
		Transport: transport,
	}
}
