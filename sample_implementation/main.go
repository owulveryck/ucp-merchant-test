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
	"time"

	"github.com/owulveryck/ucp-merchant-test/pkg/auth"
	"github.com/owulveryck/ucp-merchant-test/pkg/idempotency"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/a2a"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/discovery"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/mcp"
	"github.com/owulveryck/ucp-merchant-test/pkg/merchant/transport/rest"
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
var a2aServer *a2a.Server

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

	a2aServer = a2a.New(merchantInstance, oauthServer,
		a2a.WithMerchantName(merchantName),
		a2a.WithListenPort(func() int { return listenPort }),
		a2a.WithScheme(func() string { return scheme() }),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/events", handleSSE)
	mux.HandleFunc("/api/products", handleAPIProducts)

	// MCP transport
	mux.Handle("/mcp", mcpServer)

	// A2A transport
	mux.Handle("/a2a", a2aServer)
	mux.HandleFunc("/.well-known/agent-card.json", a2aServer.HandleAgentCard)

	// Discovery and OAuth
	disc := discovery.New(func() string {
		return fmt.Sprintf("%s://localhost:%d", scheme(), listenPort)
	})
	mux.HandleFunc("/.well-known/ucp", disc.HandleDiscovery)
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
	mux.HandleFunc("/specs/", disc.HandleSpecsAndSchemas)
	mux.HandleFunc("/schemas/", disc.HandleSpecsAndSchemas)

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
	if a2aServer != nil {
		a2aServer.Reset()
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
	if a2aServer != nil {
		a2aServer.Reset()
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
	flag.StringVar(&merchantName, "merchant-name", "", "merchant display name (random if empty)")
	flag.StringVar(&simulationSecret, "simulation-secret", "", "secret for /testing/simulate-shipping endpoint")
	flag.Parse()

	tlsEnabled = useTLS
	if merchantName == "" {
		merchantName = generateMerchantName()
	}

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
	log.Printf("A2A endpoint:  %s://localhost:%d/a2a", s, listenPort)
	log.Printf("Agent Card:    %s://localhost:%d/.well-known/agent-card.json", s, listenPort)
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

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
