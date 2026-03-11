package auth_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
)

func ExampleOAuthServer() {
	srv := auth.NewOAuthServer("FlowerShop", func() string { return "http" }, func() int { return 8182 })

	// Inject a test token (bypasses browser-based auth code flow)
	token := srv.InjectToken("john@example.com", ucp.Country("US"), time.Now().Add(auth.AccessTokenDuration))

	// Simulate an authenticated request
	req, _ := http.NewRequest("GET", "/checkout", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	user := srv.ExtractUserFromToken(req)
	country := srv.ExtractUserCountry(req)

	fmt.Println(user)
	fmt.Println(country)
	// Output:
	// john@example.com
	// US
}

// ExampleOAuthServer_tokenExchange demonstrates the full OAuth 2.0 Authorization
// Code flow with PKCE and client_secret_basic authentication.
func ExampleOAuthServer_tokenExchange() {
	srv := auth.NewOAuthServer("FlowerShop", func() string { return "http" }, func() int { return 8182 })

	// 1. Generate PKCE code verifier and challenge (S256)
	verifier := "test-verifier-string-at-least-43-chars-long-for-pkce"
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	// 2. POST to authorize endpoint (simulate user login)
	authForm := url.Values{
		"username":       {"john@example.com"},
		"country":        {"US"},
		"redirect_uri":   {"http://localhost/callback"},
		"state":          {"xyz"},
		"code_challenge": {challenge},
	}
	authReq := httptest.NewRequest("POST",
		"/oauth2/authorize?client_id="+auth.OAuthClientID+
			"&redirect_uri=http://localhost/callback"+
			"&state=xyz"+
			"&code_challenge="+challenge+
			"&code_challenge_method=S256",
		strings.NewReader(authForm.Encode()))
	authReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	srv.HandleAuthorize(w, authReq)

	// 3. Extract authorization code from redirect
	loc := w.Header().Get("Location")
	u, _ := url.Parse(loc)
	code := u.Query().Get("code")

	// 4. Exchange code for tokens using client_secret_basic authentication
	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"code_verifier": {verifier},
	}
	tokenReq := httptest.NewRequest("POST", "/oauth2/token",
		strings.NewReader(tokenForm.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.SetBasicAuth(auth.OAuthClientID, auth.OAuthClientSecret)
	w = httptest.NewRecorder()
	srv.HandleToken(w, tokenReq)

	// 5. Parse the token response
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Println(resp["token_type"])
	fmt.Println(resp["access_token"] != "")
	fmt.Println(resp["refresh_token"] != "")
	// Output:
	// Bearer
	// true
	// true
}
