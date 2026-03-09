package main

import (
	"net/http"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
)

// Global OAuth server instance.
var oauthServer = auth.NewOAuthServer(
	"",
	func() string { return scheme() },
	func() int { return listenPort },
)

func handleOAuthMetadata(w http.ResponseWriter, r *http.Request) {
	oauthServer.MerchantName = merchantName
	oauthServer.HandleMetadata(w, r)
}

func handleOAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	oauthServer.MerchantName = merchantName
	oauthServer.HandleAuthorize(w, r)
}

func handleOAuthToken(w http.ResponseWriter, r *http.Request) {
	oauthServer.HandleToken(w, r)
}

func handleOAuthRevoke(w http.ResponseWriter, r *http.Request) {
	oauthServer.HandleRevoke(w, r)
}

func extractUserFromToken(r *http.Request) string {
	return oauthServer.ExtractUserFromToken(r)
}

func extractUserCountry(r *http.Request) string {
	return oauthServer.ExtractUserCountry(r)
}

func isTokenExpired(r *http.Request) bool {
	return oauthServer.IsTokenExpired(r)
}
