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
		t.Fatal("Failed because user and pass were not provided")
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
		t.Fatal(resp.StatusCode)
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
