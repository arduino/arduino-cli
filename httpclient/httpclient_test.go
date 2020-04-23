package httpclient

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserAgentHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("User-Agent"))
	}))
	defer ts.Close()

	client := NewWithConfig(&Config{
		UserAgent: "test-user-agent",
	})

	request, err := http.NewRequest("GET", ts.URL, nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, "test-user-agent", string(b))
}

func TestProxy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	proxyURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	client := NewWithConfig(&Config{
		Proxy: proxyURL,
	})

	request, err := http.NewRequest("GET", "http://arduino.cc", nil)
	require.NoError(t, err)

	response, err := client.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}
