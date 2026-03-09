// Package auth implements the OAuth 2.0 Authorization Server for the UCP
// Identity Linking capability (dev.ucp.common.identity_linking).
//
// Identity Linking enables a platform (e.g., an AI agent, mobile app, or search
// engine) to obtain authorization to perform actions on behalf of a user on the
// business's site. This linkage is foundational for commerce experiences such as
// accessing loyalty benefits, utilizing personalized offers, managing wishlists,
// and executing authenticated checkouts.
//
// # OAuth 2.0 Implementation
//
// The OAuthServer implements the OAuth 2.0 Authorization Code flow (RFC 6749 §4.1)
// with PKCE (RFC 7636) using the S256 code challenge method. Per the UCP
// specification, businesses:
//
//   - MUST implement OAuth 2.0 (RFC 6749)
//   - MUST publish Authorization Server Metadata at /.well-known/oauth-authorization-server
//     per RFC 8414, declaring endpoints for authorization, token exchange, and revocation
//   - MUST support token revocation per RFC 7009, including recursive revocation
//     of all tokens associated with a revoked refresh token
//   - MUST provide an account creation flow if the user does not already have an account
//
// # Endpoints
//
// The OAuthServer exposes four HTTP handlers:
//
//   - HandleMetadata: serves RFC 8414 OAuth Authorization Server Metadata at
//     /.well-known/oauth-authorization-server, declaring issuer, authorization_endpoint,
//     token_endpoint, revocation_endpoint, and supported grant types
//
//   - HandleAuthorize: serves a login/consent page (GET) and processes authorization
//     grants (POST). On successful login, issues an authorization code and redirects
//     to the platform's redirect_uri with the code and state parameters
//
//   - HandleToken: exchanges authorization codes for access/refresh tokens
//     (grant_type=authorization_code) and refreshes expired access tokens
//     (grant_type=refresh_token). Validates PKCE code_verifier against the stored
//     code_challenge using SHA-256
//
//   - HandleRevoke: revokes access or refresh tokens per RFC 7009. Revoking a
//     refresh token also revokes all associated access tokens for the same user
//
// # Scopes and Capabilities
//
// UCP defines standard scopes that map to capabilities. The scope
// ucp:scopes:checkout_session grants access to all checkout operations (Get,
// Create, Update, Delete, Cancel, Complete). The platform includes the access
// token in the HTTP Authorization header using the Bearer scheme.
//
// # Token Management
//
// Access tokens expire after 1 hour (AccessTokenDuration). Refresh tokens do not
// expire but can be revoked. The ExtractUserFromToken and ExtractUserCountry
// methods validate Bearer tokens from incoming requests and return the associated
// user identity and country, enabling the merchant server to personalize
// checkout sessions with buyer-specific data (addresses, payment instruments).
//
// The InjectToken method creates tokens directly for testing purposes, bypassing
// the authorization code flow. This is used by the UCP conformance test suite
// to set up authenticated sessions without browser interaction.
package auth
