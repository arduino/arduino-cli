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
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/bcmi-labs/arduino-cli/auth"
)

var (
	testUser = os.Getenv("TEST_USERNAME")
	testPass = os.Getenv("TEST_PASSWORD")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestToken(t *testing.T) {
	if testUser == "" || testPass == "" {
		t.Skip("Skipped because user and pass were not provided")
	}
	auth := auth.New()
	token, err := auth.Token(testUser, testPass)
	if err != nil {
		t.Fatal(err)
	}

	// Obtain info
	req, err := http.NewRequest("GET", "https://auth.arduino.cc/v1/users/byID/me", nil)
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
