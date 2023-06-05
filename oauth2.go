package gcp

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

// This function is a method of the `Gcp` struct and is used to obtain an OAuth2 access token for a
// given set of scopes. It takes in a variable number of scope strings as arguments and returns an
// `oauth2.Token` and an error.
func (r *Gcp) GetOAuth2AccessToken(scope []string) (*oauth2.Token, error) {
	ctx := context.Background()

	if scope == nil {
		scope = r.scope
	}

	jwt, err := getJwtConfig(r.keyByte, r.scope)
	if err != nil {
		return nil, err
	}

	token, err := jwt.TokenSource(ctx).Token()
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain Access Token from JWT config with scope %s <%w>", scope, err)
	}

	return token, nil
}

// This is a method of the `Gcp` struct that is used to obtain an OAuth2 ID token for a given set of
// scopes. It takes in a variable number of scope strings as arguments and returns an `oauth2.Token`
// and an error. It first checks if the `scope` argument is nil, and if so, it sets it to the default
// `scope` value of the `Gcp` struct. It then calls the `getTokenSource` function to obtain a token
// source with the specified scopes and uses it to obtain the ID token by calling the `Token` method on
// the token source. If there is an error obtaining the token source or the token itself, an error is
// returned.
func (r *Gcp) GetOAuth2IdToken(scope []string) (*oauth2.Token, error) {
	if scope == nil {
		scope = r.scope
	}

	ts, err := getTokenSource(r.keyByte, r.scope)
	if err != nil {
		return nil, err
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain ID Token from JWT Token Source for scope %s <%w>", scope, err)
	}

	return token, nil
}

// The function returns a JWT configuration and an error, given a key byte and a scope.
func getJwtConfig(keyByte []byte, scope []string) (*jwt.Config, error) {
	jwt, err := google.JWTConfigFromJSON(keyByte, scope...)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain JWT Config for scope %s <%w>", scope, err)
	}

	return jwt, nil
}

// The function returns a JWT token source for a given set of credentials and scope.
func getTokenSource(keyByte []byte, scope []string) (oauth2.TokenSource, error) {
	ts, err := google.JWTAccessTokenSourceWithScope(keyByte, scope...)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain JWT Token Source for scope %s <%w>", scope, err)
	}

	return ts, nil
}
