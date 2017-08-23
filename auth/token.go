package auth

import (
	"fmt"
)

// Token is the response of the two authentication functions
type Token struct {
	// Access is the token to use to authenticate requests
	Access string `json:"access_token"`

	// Refresh is the token to use to request another access token. It's only returned if one of the scopes is "offline"
	Refresh string `json:"refresh_token"`

	// TTL is the number of seconds that the tokens will last
	TTL int `json:"expires_in"`

	// Scopes is a space-separated list of scopes associated to the access token
	Scopes string `json:"scope"`

	// Type is the type of token
	Type string `json:"token_type"`
}

func (t Token) String() string {
	return fmt.Sprintln("ACCESS: ", t.Access) +
		fmt.Sprintln("REFRESH: ", t.Refresh) +
		fmt.Sprintln("TTL: ", t.TTL) +
		fmt.Sprintln("TYPE: ", t.Type) +
		fmt.Sprintln("SCOPES:", t.Scopes)
}

// Config contains the variables you may want to change
type Config struct {
	// CodeURL is the endpoint to redirect to obtain a code
	CodeURL string

	// TokenURL is the endpoint where you can request an access code
	TokenURL string

	// ClientID is the client id you are using
	ClientID string

	// RedirectURI is the redirectURI where the oauth process will redirect. It's only required since the oauth system checks for it, but we intercept the redirect before hitting it
	RedirectURI string

	// Scopes is a space-separated list of scopes to require
	Scopes string
}

func (c Config) String() string {
	return fmt.Sprintln("CODE URL: ", c.CodeURL) +
		fmt.Sprintln("TOKEN URL: ", c.TokenURL) +
		fmt.Sprintln("CLIENT ID: ", c.ClientID) +
		fmt.Sprintln("REDIRECT URI: ", c.RedirectURI) +
		fmt.Sprintln("SCOPES:", c.Scopes)
}
