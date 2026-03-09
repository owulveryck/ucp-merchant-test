package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	mrand "math/rand"

	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
)

// Merchant identity
var (
	merchantName     string
	listenPort       int
	tlsEnabled       bool
	simulationSecret string
	dataDir          string
	dataFormat       string
)

// Global OAuth server instance.
var oauthServer = auth.NewOAuthServer(
	"",
	func() string { return scheme() },
	func() int { return listenPort },
)

var adjectives = []string{"Swift", "Bright", "Golden", "Silver", "Crystal", "Noble", "Royal", "Grand", "Prime", "Elite"}
var nouns = []string{"Falcon", "Coral", "Harbor", "Summit", "Valley", "Atlas", "Phoenix", "Horizon", "Crest", "Bridge"}

func generateMerchantName() string {
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	return adjectives[r.Intn(len(adjectives))] + " " + nouns[r.Intn(len(nouns))]
}

// Session tracking
var (
	sessionCounter int
	sessionMu      sync.Mutex
)

func newSessionID() string {
	sessionMu.Lock()
	defer sessionMu.Unlock()
	sessionCounter++
	return fmt.Sprintf("session-%04d", sessionCounter)
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/events", handleSSE)
	mux.HandleFunc("/api/products", handleAPIProducts)
	mux.HandleFunc("/mcp", handleMCP)
	mux.HandleFunc("/.well-known/ucp", handleUCPDiscovery)
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		oauthServer.MerchantName = merchantName
		oauthServer.HandleMetadata(w, r)
	})
	mux.HandleFunc("/oauth2/authorize", func(w http.ResponseWriter, r *http.Request) {
		oauthServer.MerchantName = merchantName
		oauthServer.HandleAuthorize(w, r)
	})
	mux.HandleFunc("/oauth2/token", oauthServer.HandleToken)
	mux.HandleFunc("/oauth2/revoke", oauthServer.HandleRevoke)

	// REST API routes
	mux.HandleFunc("/shopping-api/checkout-sessions/", restHandleCheckoutSessions)
	mux.HandleFunc("/shopping-api/checkout-sessions", restHandleCheckoutSessions)
	mux.HandleFunc("/orders/", restHandleOrders)
	mux.HandleFunc("/testing/simulate-shipping/", restSimulateShipping)
	mux.HandleFunc("/specs/", handleSpecsAndSchemas)
	mux.HandleFunc("/schemas/", handleSpecsAndSchemas)
	return mux
}

func main() {
	var (
		useTLS   bool
		certFile string
		keyFile  string
		dbFile   string
	)
	flag.IntVar(&listenPort, "port", 8081, "port to listen on")
	flag.BoolVar(&useTLS, "tls", false, "enable TLS (auto-generates self-signed cert if --cert/--key not provided)")
	flag.StringVar(&certFile, "cert", "", "path to TLS certificate file")
	flag.StringVar(&keyFile, "key", "", "path to TLS private key file")
	flag.StringVar(&dbFile, "db", "", "path to JSON product database file (overrides built-in catalog)")
	flag.StringVar(&dataDir, "data-dir", "", "path to directory with test data (flower shop dataset)")
	flag.StringVar(&dataFormat, "data-format", "csv", "format of test data files: csv or json")
	flag.StringVar(&simulationSecret, "simulation-secret", "", "secret for /testing/simulate-shipping endpoint")
	flag.Parse()

	tlsEnabled = useTLS
	merchantName = generateMerchantName()

	if dataDir != "" {
		if err := loadFlowerShopData(dataDir, dataFormat); err != nil {
			log.Fatalf("Failed to load flower shop data from %s: %v", dataDir, err)
		}
		log.Printf("Loaded %d products from %s", len(catalog), dataDir)
	} else if dbFile != "" {
		if err := loadCatalogFromFile(dbFile); err != nil {
			log.Fatalf("Failed to load product database %s: %v", dbFile, err)
		}
		log.Printf("Loaded %d products from %s", len(catalog), dbFile)
	} else {
		initCatalog(time.Now().UnixNano())
	}

	if simulationSecret == "" {
		simulationSecret = fmt.Sprintf("sim-%d", time.Now().UnixNano())
	}

	mux := newMux()

	s := "http"
	if tlsEnabled {
		s = "https"
	}
	addr := fmt.Sprintf(":%d", listenPort)
	log.Printf("%s starting on %s://localhost:%d", merchantName, s, listenPort)
	log.Printf("Dashboard:     %s://localhost:%d/", s, listenPort)
	log.Printf("MCP endpoint:  %s://localhost:%d/mcp", s, listenPort)
	log.Printf("REST endpoint: %s://localhost:%d/shopping-api", s, listenPort)
	log.Printf("UCP profile:   %s://localhost:%d/.well-known/ucp", s, listenPort)

	if !tlsEnabled {
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Fatal(err)
		}
		return
	}

	// TLS mode
	if certFile != "" && keyFile != "" {
		log.Printf("TLS enabled with provided certificate: %s", certFile)
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, mux); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Generate self-signed certificate
	log.Println("TLS enabled with self-signed certificate.")
	log.Printf("Visit https://localhost:%d in your browser to accept the certificate.", listenPort)

	tlsCert, err := generateSelfSignedCert()
	if err != nil {
		log.Fatalf("Failed to generate self-signed certificate: %v", err)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	tlsListener := tls.NewListener(ln, server.TLSConfig)
	if err := server.Serve(tlsListener); err != nil {
		log.Fatal(err)
	}
}

func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"UCP Sample Merchant"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func scheme() string {
	if tlsEnabled {
		return "https"
	}
	return "http"
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")
}

func handleUCPDiscovery(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	base := fmt.Sprintf("%s://localhost:%d", scheme(), listenPort)
	json.NewEncoder(w).Encode(model.UCPDiscovery{
		UCP: model.UCPDiscoveryProfile{
			Version: "2026-01-11",
			Services: map[string]model.UCPServiceEntry{
				"dev.ucp.shopping": {
					Version: "2026-01-11",
					Spec:    base + "/specs/shopping",
					REST: &model.UCPRESTConfig{
						Endpoint: base + "/shopping-api",
						Schema:   base + "/schemas/shopping/rest.json",
					},
				},
			},
			Capabilities: []model.UCPCapabilityEntry{
				{Name: "dev.ucp.shopping.checkout", Version: "2026-01-11", Spec: base + "/specs/shopping/checkout", Schema: base + "/schemas/shopping/checkout.json"},
				{Name: "dev.ucp.shopping.order", Version: "2026-01-11", Spec: base + "/specs/shopping/order", Schema: base + "/schemas/shopping/order.json"},
				{Name: "dev.ucp.shopping.discount", Version: "2026-01-11", Spec: base + "/specs/shopping/discount", Schema: base + "/schemas/shopping/discount.json"},
				{Name: "dev.ucp.shopping.fulfillment", Version: "2026-01-11", Spec: base + "/specs/shopping/fulfillment", Schema: base + "/schemas/shopping/fulfillment.json"},
				{Name: "dev.ucp.shopping.buyer_consent", Version: "2026-01-11", Spec: base + "/specs/shopping/buyer_consent", Schema: base + "/schemas/shopping/buyer_consent.json"},
			},
		},
		Payment: model.UCPPaymentProfile{
			Handlers: []map[string]any{
				{
					"id":                 "google_pay",
					"name":               "google.pay",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/google_pay",
					"config_schema":      base + "/schemas/payment/google_pay.json",
					"instrument_schemas": []string{base + "/schemas/payment/google_pay_instrument.json"},
					"config":             map[string]any{},
				},
				{
					"id":                 "mock_payment_handler",
					"name":               "mock_payment_handler",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/mock",
					"config_schema":      base + "/schemas/payment/mock.json",
					"instrument_schemas": []string{base + "/schemas/payment/mock_instrument.json"},
					"config":             map[string]any{},
				},
				{
					"id":                 "shop_pay",
					"name":               "com.shopify.shop_pay",
					"version":            "2026-01-11",
					"spec":               base + "/specs/payment/shop_pay",
					"config_schema":      base + "/schemas/payment/shop_pay.json",
					"instrument_schemas": []string{base + "/schemas/payment/shop_pay_instrument.json"},
					"config":             map[string]any{"shop_id": "merchant_1"},
				},
			},
		},
	})
}

func handleSpecsAndSchemas(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if strings.HasPrefix(r.URL.Path, "/schemas/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)
	} else {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `<!DOCTYPE html><html><head><title>UCP Spec</title></head><body><h1>%s</h1></body></html>`, r.URL.Path)
	}
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for expired Bearer token — return 401 so platform can refresh
	if oauthServer.IsTokenExpired(r) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "token_expired"})
		return
	}

	// Extract authenticated user (empty string = guest)
	userID := oauthServer.ExtractUserFromToken(r)
	userCountry := oauthServer.ExtractUserCountry(r)

	var req model.JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &model.RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	// Assign session ID on initialize
	if req.Method == "initialize" {
		sid := newSessionID()
		w.Header().Set("Mcp-Session-Id", sid)
	} else if sid := r.Header.Get("Mcp-Session-Id"); sid != "" {
		w.Header().Set("Mcp-Session-Id", sid)
	}

	switch req.Method {
	case "initialize":
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPInitializeResult{
				ProtocolVersion: "2025-03-26",
				Capabilities:    model.MCPCapabilities{Tools: map[string]any{}},
				ServerInfo:      model.MCPServerInfo{Name: merchantName, Version: "1.0.0"},
			},
		})

	case "notifications/initialized":
		// Notification — no response needed
		w.WriteHeader(http.StatusNoContent)

	case "tools/list":
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPToolsListResult{
				Tools: getToolDefinitions(),
			},
		})

	case "tools/call":
		handleToolCall(w, req, userID, userCountry)

	default:
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		})
	}
}

func handleToolCall(w http.ResponseWriter, req model.JSONRPCRequest, userID, userCountry string) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32602, Message: "Invalid params"},
		})
		return
	}

	// Inject userID and country into arguments so handlers can scope data
	params.Arguments["_user_id"] = userID
	params.Arguments["_user_country"] = userCountry

	handlers := map[string]func(map[string]interface{}) (interface{}, error){
		"list_products":        handleListProducts,
		"get_product_details":  handleGetProductDetails,
		"search_catalog":       handleSearchCatalog,
		"lookup_product":       handleLookupProduct,
		"create_cart":          handleCreateCart,
		"get_cart":             handleGetCart,
		"update_cart":          handleUpdateCart,
		"cancel_cart":          handleCancelCart,
		"create_checkout":      handleCreateCheckout,
		"get_checkout":         handleGetCheckout,
		"update_checkout":      handleUpdateCheckout,
		"complete_checkout":    handleCompleteCheckout,
		"cancel_checkout":      handleCancelCheckout,
		"get_order":            handleGetOrder,
		"list_orders":          handleListOrders,
		"cancel_order":         handleCancelOrder,
		"track_order":          handleTrackOrder, // kept for backward compat
		"get_shipping_options": handleGetShippingOptions,
		"track_shipment":       handleTrackShipment,
	}

	handler, ok := handlers[params.Name]
	if !ok {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &model.RPCError{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
		})
		return
	}

	result, err := handler(params.Arguments)
	if err != nil {
		writeJSON(w, model.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: model.MCPToolResult{
				Content: []model.MCPContentBlock{
					{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())},
				},
				IsError: true,
			},
		})
		return
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	content := []model.MCPContentBlock{
		{Type: "text", Text: string(resultJSON)},
	}

	// Extract image URLs from the result and add as image content blocks (cap at 5)
	imageURLs := extractImageURLs(result)
	if len(imageURLs) > 5 {
		imageURLs = imageURLs[:5]
	}
	for _, imgURL := range imageURLs {
		data, mime, err := fetchAndEncodeImage(imgURL)
		if err != nil {
			log.Printf("Failed to fetch image %s: %v", imgURL, err)
			continue
		}
		content = append(content, model.MCPContentBlock{
			Type:     "image",
			Data:     data,
			MimeType: mime,
		})
	}

	writeJSON(w, model.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: model.MCPToolResult{
			Content: content,
		},
	})
}

// fetchAndEncodeImage fetches an image URL and returns its base64-encoded data and MIME type.
func fetchAndEncodeImage(url string) (string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	return base64.StdEncoding.EncodeToString(body), mimeType, nil
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func getToolDefinitions() []model.ToolDef {
	metaSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"ucp-agent":       map[string]interface{}{"type": "string", "description": "URI identifying the calling agent"},
			"idempotency-key": map[string]interface{}{"type": "string", "description": "UUID for idempotent operations"},
		},
	}
	lineItemsSchema := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"product_id": map[string]interface{}{"type": "string", "description": "Product SKU ID from the catalog"},
				"quantity":   map[string]interface{}{"type": "integer", "description": "Quantity to add", "minimum": 1},
			},
			"required": []string{"product_id"},
		},
	}

	return []model.ToolDef{
		{
			Name:        "list_products",
			Description: "List products from the catalog. Call with no arguments first to see featured products and available categories. Use category, brand, query, or usage_type filters to narrow results. Each product includes a usage_type (intensive, occasional, versatile). Use get_product_details to get the full product sheet with description before recommending.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"category":   map[string]interface{}{"type": "string", "description": "Filter by category (case-insensitive)"},
					"brand":      map[string]interface{}{"type": "string", "description": "Filter by brand (case-insensitive)"},
					"query":      map[string]interface{}{"type": "string", "description": "Text search on product title (case-insensitive partial match)"},
					"usage_type": map[string]interface{}{"type": "string", "description": "Filter by usage type: intensive, occasional, or versatile", "enum": []string{"intensive", "occasional", "versatile"}},
					"limit":      map[string]interface{}{"type": "integer", "description": "Max results per page (default 20, max 50)"},
					"offset":     map[string]interface{}{"type": "integer", "description": "Skip N products for pagination (default 0)"},
				},
			},
		},
		{
			Name:        "get_product_details",
			Description: "Get the full product sheet (fiche produit) for a specific product, including description, usage type, and availability. Use this after list_products to get details before recommending a product.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Product SKU ID (e.g. SKU-001)"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "search_catalog",
			Description: "Search the product catalog by keyword. Returns matching products with availability info. Supports price range filtering and in-stock filtering.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query":              map[string]interface{}{"type": "string", "description": "Search query string (matches title, description, category)"},
					"limit":              map[string]interface{}{"type": "integer", "description": "Max results to return (1-300, default 10)"},
					"min_price":          map[string]interface{}{"type": "integer", "description": "Minimum price in cents (0 = no minimum)"},
					"max_price":          map[string]interface{}{"type": "integer", "description": "Maximum price in cents (0 = no maximum)"},
					"available_for_sale": map[string]interface{}{"type": "boolean", "description": "If true, only return in-stock products"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "lookup_product",
			Description: "Look up a single product by its ID. Returns full product details including description, price, stock, and available countries.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "string", "description": "Product ID (e.g. SKU-001)"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "create_cart",
			Description: "Create a new shopping cart with line items. Each line item needs a product_id (from list_products) and quantity.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"cart": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
						},
						"required": []string{"line_items"},
					},
				},
				"required": []string{"cart"},
			},
		},
		{
			Name:        "get_cart",
			Description: "Retrieve a cart by its ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_cart",
			Description: "Update the line items of an existing cart",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
					"cart": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
						},
						"required": []string{"line_items"},
					},
				},
				"required": []string{"id", "cart"},
			},
		},
		{
			Name:        "cancel_cart",
			Description: "Cancel and remove a cart",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Cart ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "create_checkout",
			Description: "Create a checkout session. Provide either line_items directly or a cart_id to create from an existing cart. Optionally include buyer information.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items": lineItemsSchema,
							"cart_id":    map[string]interface{}{"type": "string", "description": "Create checkout from existing cart"},
							"buyer": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name":  map[string]interface{}{"type": "string"},
									"email": map[string]interface{}{"type": "string"},
									"address": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"street":  map[string]interface{}{"type": "string"},
											"city":    map[string]interface{}{"type": "string"},
											"state":   map[string]interface{}{"type": "string"},
											"zip":     map[string]interface{}{"type": "string"},
											"country": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
				"required": []string{"checkout"},
			},
		},
		{
			Name:        "get_checkout",
			Description: "Retrieve a checkout by its ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "update_checkout",
			Description: "Update a checkout's line items, buyer information, or shipping option. When buyer address is provided, status transitions to ready_for_complete. Use shipping_option_id from get_shipping_options to select a delivery method.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"line_items":         lineItemsSchema,
							"shipping_option_id": map[string]interface{}{"type": "string", "description": "ID of selected shipping option from get_shipping_options"},
							"buyer": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"name":  map[string]interface{}{"type": "string"},
									"email": map[string]interface{}{"type": "string"},
									"address": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"street":  map[string]interface{}{"type": "string"},
											"city":    map[string]interface{}{"type": "string"},
											"state":   map[string]interface{}{"type": "string"},
											"zip":     map[string]interface{}{"type": "string"},
											"country": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
				"required": []string{"id", "checkout"},
			},
		},
		{
			Name:        "complete_checkout",
			Description: "Complete a checkout and place the order. Checkout must be in ready_for_complete status. Requires an approval object with checkout_hash obtained from get_checkout, confirming the user has reviewed and approved the purchase.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
					"approval": map[string]interface{}{
						"type":        "object",
						"description": "User approval with checkout hash to verify the user reviewed the exact checkout state",
						"properties": map[string]interface{}{
							"checkout_hash": map[string]interface{}{"type": "string", "description": "The checkout_hash from get_checkout, proving the user approved this exact state"},
						},
						"required": []string{"checkout_hash"},
					},
					"checkout": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"payment": map[string]interface{}{
								"type":        "object",
								"description": "Payment information (any object accepted for testing)",
							},
						},
					},
				},
				"required": []string{"id", "approval"},
			},
		},
		{
			Name:        "cancel_checkout",
			Description: "Cancel a checkout session. Cannot cancel completed checkouts.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Checkout ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "get_order",
			Description: "Retrieve a placed order by its ID, including line items, totals, buyer info, and shipment tracking details.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Order ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "list_orders",
			Description: "List all placed orders with their current status, confirmation number, total, and creation date.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
				},
			},
		},
		{
			Name:        "cancel_order",
			Description: "Cancel an order. Only possible when order status is 'confirmed' or 'processing' (before it has shipped). Once shipped, the order cannot be canceled.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta": metaSchema,
					"id":   map[string]interface{}{"type": "string", "description": "Order ID"},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "get_shipping_options",
			Description: "Get available shipping options for a checkout. Returns delivery methods with estimated delivery times and costs. Call this after creating a checkout to present delivery choices to the user.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta":        metaSchema,
					"checkout_id": map[string]interface{}{"type": "string", "description": "Checkout ID to get shipping options for"},
				},
				"required": []string{"checkout_id"},
			},
		},
		{
			Name:        "track_shipment",
			Description: "Track a shipment for an order. Returns tracking number, carrier, estimated delivery date, and current delivery status.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"meta":     metaSchema,
					"order_id": map[string]interface{}{"type": "string", "description": "Order ID to track"},
				},
				"required": []string{"order_id"},
			},
		},
	}
}
