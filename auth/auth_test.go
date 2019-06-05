//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package auth_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/arduino/arduino-cli/auth"
	"github.com/stretchr/testify/require"
)

var (
	testUser = os.Getenv("TEST_USERNAME")
	testPass = os.Getenv("TEST_PASSWORD")
)

func TestNewConfig(t *testing.T) {
	conf := auth.New()
	require.Equal(t, "https://hydra.arduino.cc/oauth2/auth", conf.CodeURL)
	require.Equal(t, "https://hydra.arduino.cc/oauth2/token", conf.TokenURL)
	require.Equal(t, "cli", conf.ClientID)
	require.Equal(t, "http://localhost:5000", conf.RedirectURI)
	require.Equal(t, "profile:core offline", conf.Scopes)
}

func TestTokenIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test")
	}

	if testUser == "" || testPass == "" {
		t.Skip("Skipped because user and pass were not provided")
	}
	auth := auth.New()
	token, err := auth.Token(testUser, testPass)
	if err != nil {
		t.Fatal(err)
	}

	// Obtain info
	req, err := http.NewRequest("GET", "https://ddauth.arduino.cc/v1/users/byID/me", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Authorization", "Bearer "+token.Access)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatal(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var data struct{ Username string }
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatal(err)
	}

	if data.Username != testUser {
		t.Fatalf("Expected username '%s', got '%s'", testUser, data.Username)
	}
}
