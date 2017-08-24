package auth_test

import "testing"
import "github.com/bcmi-labs/arduino-cli/auth"

const (
	testUsername = "ARDUINO-CLI-TEST-USER"
	testPassword = "clitestuser"
)

func testConfig_Token(t *testing.T) (*auth.Config, *auth.Token) {
	tokenConfig := auth.New()
	t.Log(tokenConfig)
	token, err := tokenConfig.Token(testUsername, testPassword)
	if err != nil {
		t.Error(err)
	}
	t.Log(token)
	return tokenConfig, token
}

func TestConfig_Refresh(t *testing.T) {
	tokenConfig, token := testConfig_Token(t)

	newToken, err := tokenConfig.Refresh(token.Refresh)
	if err != nil {
		t.Error(err)
	}
	t.Log(newToken)
}

func TestConfig_MyUser(t *testing.T) {
	_, token := testConfig_Token(t)
	user, err := token.LoggedUser()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(user)
}
