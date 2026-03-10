# UCP Merchant Test Server

[![UCP Conformance Tests](https://github.com/owulveryck/ucp-merchant-test/actions/workflows/conformance.yml/badge.svg)](https://github.com/owulveryck/ucp-merchant-test/actions/workflows/conformance.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/owulveryck/ucp-merchant-test.svg)](https://pkg.go.dev/github.com/owulveryck/ucp-merchant-test)

A Go-based merchant server implementing the [Universal Commerce Protocol (UCP)](https://ucp.dev) Shopping Service. Supports both MCP (JSON-RPC) and REST API transports. Passes all 60 UCP conformance tests.

## Quick Start

```bash
go run ./sample_implementation --port 8182 \
  --data-dir path/to/test_data/flower_shop \
  --simulation-secret super-secret-sim-key
```

Server starts on `http://localhost:8182`:
- **REST endpoint**: `http://localhost:8182/shopping-api`
- **MCP endpoint**: `http://localhost:8182/mcp`
- **UCP discovery**: `http://localhost:8182/.well-known/ucp`
- **Dashboard**: `http://localhost:8182/`

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8081` | HTTP listen port |
| `--data-dir` | _(none)_ | Path to CSV test data directory (flower shop dataset) |
| `--simulation-secret` | _(random UUID)_ | Secret for the `/testing/simulate-shipping/` endpoint |
| `--tls` | `false` | Enable TLS with self-signed certificate |
| `--cert` | _(none)_ | TLS certificate file |
| `--key` | _(none)_ | TLS key file |
| `--db` | _(none)_ | JSON file with custom product catalog |

## REST API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/shopping-api/checkout-sessions` | Create a checkout session |
| `GET` | `/shopping-api/checkout-sessions/{id}` | Retrieve a checkout session |
| `PUT` | `/shopping-api/checkout-sessions/{id}` | Update a checkout session |
| `POST` | `/shopping-api/checkout-sessions/{id}/complete` | Complete checkout and place order |
| `POST` | `/shopping-api/checkout-sessions/{id}/cancel` | Cancel a checkout session |
| `GET` | `/orders/{id}` | Retrieve an order |
| `PUT` | `/orders/{id}` | Update an order |
| `POST` | `/testing/simulate-shipping/{id}` | Simulate shipping (requires `Simulation-Secret` header) |

## Test Data

The `--data-dir` flag points to a directory containing CSV files for the flower shop dataset:

- `products.csv` - Product catalog (id, title, price, image_url)
- `inventory.csv` - Stock quantities per product
- `customers.csv` - Customer records (id, name, email)
- `addresses.csv` - Customer addresses
- `payment_instruments.csv` - Payment instruments (id, type, brand, last_digits, token, handler_id)
- `discounts.csv` - Discount codes (percentage and fixed amount)
- `shipping_rates.csv` - Shipping rates by country and service level
- `promotions.csv` - Free shipping promotions
- `conformance_input.json` - Reference item info for conformance tests

## Running Conformance Tests

```bash
# Start the merchant
go run ./sample_implementation --port 8182 \
  --data-dir /path/to/conformance/test_data/flower_shop \
  --simulation-secret super-secret-sim-key

# Run all conformance tests
cd /path/to/Universal-Commerce-Protocol/conformance
for test_file in *_test.py; do
  python3 "$test_file" \
    --server_url=http://localhost:8182 \
    --simulation_secret=super-secret-sim-key \
    --conformance_input=test_data/flower_shop/conformance_input.json \
    --test_data_dir=test_data/flower_shop
done
```

All 60 tests across 13 test files should pass:

| Test File | Tests |
|-----------|-------|
| protocol_test.py | 3 |
| checkout_lifecycle_test.py | 11 |
| validation_test.py | 6 |
| business_logic_test.py | 8 |
| fulfillment_test.py | 11 |
| order_test.py | 4 |
| idempotency_test.py | 4 |
| webhook_test.py | 3 |
| simulation_url_security_test.py | 3 |
| binding_test.py | 1 |
| invalid_input_test.py | 3 |
| card_credential_test.py | 1 |
| ap2_test.py | 1 |

## Features

- **UCP Discovery** at `/.well-known/ucp` (version `2026-01-11`)
- **Checkout lifecycle** with status transitions: `incomplete` -> `completed` / `canceled`
- **Hierarchical fulfillment**: methods -> destinations -> groups -> options
- **Address injection** for known customers (lookup by email)
- **Dynamic shipping options** based on destination country
- **Free shipping promotions** (subtotal threshold and eligible item matching)
- **Discount codes** (percentage and fixed amount, sequential application)
- **Payment processing** with token-based success/failure
- **Idempotency keys** with SHA-256 payload hashing and conflict detection
- **Webhooks** for `order_placed` and `order_shipped` events
- **Simulation endpoint** for testing shipping flows
- **Version negotiation** via `UCP-Agent` header
- **Buyer consent** (marketing, analytics, sale_of_data)
- **MCP transport** with JSON-RPC 2.0 for tool-based interactions
- **OAuth2 server** for identity linking
- **SSE dashboard** for real-time event monitoring

## Running with TLS

```bash
# Self-signed certificate
go run ./sample_implementation --tls

# With mkcert (trusted local cert)
mkcert localhost 127.0.0.1
go run ./sample_implementation --tls --cert localhost+1.pem --key localhost+1-key.pem
```

## Project Structure

```
sample_implementation/          # UCP merchant server binary
  main.go                       # Server setup, routes, UCP discovery
  merchant_impl.go              # merchant.Merchant implementation
  data.go                       # CSV/JSON data loading
  catalog.go / catalog_impl.go  # Product catalog
  dashboard.go                  # SSE dashboard

internal/
  merchant/                     # Merchant interface and transport adapters
    merchant.go                 # Merchant interface (Cataloger, Carter, Checkouter, Orderer)
    errors.go                   # Sentinel errors for transport mapping
    transport/rest/             # UCP REST transport (HTTP handlers)
    transport/mcp/              # UCP MCP transport (JSON-RPC 2.0 handlers)
    discount/                   # Discount code application logic
    fulfillment/                # Fulfillment parsing and shipping options
    payment/                    # Payment and buyer parsing
    pricing/                    # Line item and totals calculation
  model/                        # UCP data models
  auth/                         # OAuth2 server
  catalog/                      # Catalog interface
  idempotency/                  # Idempotency key tracking
  webhook/                      # Webhook dispatch
  event/                        # SSE event hub
  store/                        # Store interface
  config/                       # Configuration types
```
