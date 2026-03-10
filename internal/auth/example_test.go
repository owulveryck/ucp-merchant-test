package auth_test

import (
	"fmt"
	"net/http"
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
