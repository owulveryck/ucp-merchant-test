package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/owulveryck/ucp-merchant-test/internal/auth"
	"github.com/owulveryck/ucp-merchant-test/internal/idempotency"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/transport/mcp"
	"github.com/owulveryck/ucp-merchant-test/internal/merchant/transport/rest"
	"github.com/owulveryck/ucp-merchant-test/internal/model"
	"github.com/owulveryck/ucp-merchant-test/internal/ucp"
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

// Global idempotency store instance.
var idempotencyStoreInstance = idempotency.NewStore()

// Global merchant instance — initialized in newMux or main.
var merchantInstance *simpleMerchant

// Global transport instances — initialized in newMux or main.
var restServer *rest.Server
var mcpServer *mcp.Server

var adjectives = []string{"Swift", "Bright", "Golden", "Silver", "Crystal", "Noble", "Royal", "Grand", "Prime", "Elite"}
var nouns = []string{"Falcon", "Coral", "Harbor", "Summit", "Valley", "Atlas", "Phoenix", "Horizon", "Crest", "Bridge"}

func generateMerchantName() string {
	r := mrand.New(mrand.NewSource(time.Now().UnixNano()))
	return adjectives[r.Intn(len(adjectives))] + " " + nouns[r.Intn(len(nouns))]
}

func newMux() *http.ServeMux {
	merchantInstance = newSimpleMerchant(
		catalogInstance,
		shopData,
		func() int { return listenPort },
		func() string { return scheme() },
	)

	restServer = rest.New(merchantInstance, oauthServer,
		rest.WithIdempotency(idempotencyStoreInstance),
		rest.WithSimulationSecret(simulationSecret),
		rest.WithScheme(func() string { return scheme() }),
		rest.WithListenPort(func() int { return listenPort }),
	)

	mcpServer = mcp.New(merchantInstance, oauthServer,
		mcp.WithMerchantName(merchantName),
		mcp.WithListenPort(func() int { return listenPort }),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/events", handleSSE)
	mux.HandleFunc("/api/products", handleAPIProducts)

	// MCP transport
	mux.Handle("/mcp", mcpServer)

	// Discovery and OAuth
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

	// REST API routes (via transport)
	restHandler := restServer.Handler()
	mux.Handle("/shopping-api/checkout-sessions/", restHandler)
	mux.Handle("/shopping-api/checkout-sessions", restHandler)
	mux.Handle("/orders/", restHandler)
	mux.Handle("/testing/simulate-shipping/", restHandler)

	// Specs and schemas
	mux.HandleFunc("/specs/", handleSpecsAndSchemas)
	mux.HandleFunc("/schemas/", handleSpecsAndSchemas)

	// Reset endpoint
	mux.HandleFunc("/testing/reset", handleReset)

	return mux
}

func handleReset(w http.ResponseWriter, r *http.Request) {
	if merchantInstance != nil {
		merchantInstance.Reset()
	}
	if restServer != nil {
		restServer.Reset()
	}
	if mcpServer != nil {
		mcpServer.Reset()
	}
	oauthServer.Reset()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}

func resetStores() {
	if merchantInstance != nil {
		merchantInstance.Reset()
	}
	if restServer != nil {
		restServer.Reset()
	}
	if mcpServer != nil {
		mcpServer.Reset()
	}
	idempotencyStoreInstance.Reset()
	oauthServer.Reset()
}

func main() {
	var (
		useTLS   bool
		certFile string
		keyFile  string
	)
	flag.IntVar(&listenPort, "port", 8081, "port to listen on")
	flag.BoolVar(&useTLS, "tls", false, "enable TLS (auto-generates self-signed cert if --cert/--key not provided)")
	flag.StringVar(&certFile, "cert", "", "path to TLS certificate file")
	flag.StringVar(&keyFile, "key", "", "path to TLS private key file")
	flag.StringVar(&dataDir, "data-dir", "", "path to directory with test data (required)")
	flag.StringVar(&dataFormat, "data-format", "csv", "format of test data files: csv or json")
	flag.StringVar(&simulationSecret, "simulation-secret", "", "secret for /testing/simulate-shipping endpoint")
	flag.Parse()

	tlsEnabled = useTLS
	merchantName = generateMerchantName()

	if dataDir == "" {
		log.Fatalf("--data-dir is required")
	}
	if err := loadFlowerShopData(dataDir, dataFormat); err != nil {
		log.Fatalf("Failed to load data from %s: %v", dataDir, err)
	}
	log.Printf("Loaded %d products from %s", len(catalog), dataDir)

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
			Services: map[ucp.UCPService]model.UCPServiceEntry{
				ucp.ServiceShopping: {
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
