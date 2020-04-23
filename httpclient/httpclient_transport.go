package httpclient

import "net/http"

type httpClientRoundTripper struct {
	transport http.RoundTripper
	config    *Config
}

func newHTTPClientTransport(config *Config) http.RoundTripper {
	proxy := http.ProxyURL(config.Proxy)

	transport := &http.Transport{
		Proxy: proxy,
	}

	return &httpClientRoundTripper{
		transport: transport,
		config:    config,
	}
}

func (h *httpClientRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", h.config.UserAgent)
	return h.transport.RoundTrip(req)
}
