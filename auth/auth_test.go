package auth_test

import "testing"
import "github.com/bcmi-labs/arduino-cli/auth"

const (
	testUsername = "ARDUINO-CLI-TEST-USER"
	testPassword = "clitestuser"
)

func TestConfig_Token(t *testing.T) {
	tokenConfig := auth.New()
	t.Log(tokenConfig)
	token, err := tokenConfig.Token(testUsername, testPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
}

func TestConfig_Refresh(t *testing.T) {
	tokenConfig := auth.New()
	t.Log(tokenConfig)
	token, err := tokenConfig.Token(testUsername, testPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	newToken, err := tokenConfig.Refresh(token.Refresh)
	if err != nil {
		t.Error(err)
	}
	t.Log(newToken)
}
