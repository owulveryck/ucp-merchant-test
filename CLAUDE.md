# CLAUDE.md

## Project Overview

UCP merchant test server written in Go. Implements the Universal Commerce Protocol (UCP) Shopping Service with both REST and MCP (JSON-RPC) transports. Designed to pass the UCP conformance test suite (60 tests across 13 files).

## Build & Run

```bash
# Build
go build ./sample_implementation

# Run with flower shop test data
go run ./sample_implementation --port 8182 \
  --data-dir /path/to/conformance/test_data/flower_shop \
  --simulation-secret super-secret-sim-key

# Run Go tests
go test -count=1 ./...

# Run conformance tests (from the conformance directory)
python3 <test_file>.py \
  --server_url=http://localhost:8182 \
  --simulation_secret=super-secret-sim-key \
  --conformance_input=test_data/flower_shop/conformance_input.json \
  --test_data_dir=test_data/flower_shop
```

## Key Architecture

The merchant server binary lives under `sample_implementation/`. The `merchant.Merchant` interface, sentinel errors, and transport adapters (REST, MCP) are in `internal/merchant/`. Business logic sub-packages are in `internal/merchant/` as well.

- **REST transport** (`internal/merchant/transport/rest/`): HTTP handlers for checkout sessions, orders, simulation.
- **MCP transport** (`internal/merchant/transport/mcp/`): JSON-RPC 2.0 tool handlers.
- **Models** (`internal/model/`): UCP data types (`model.Checkout`, `model.Order`, etc.). All source files use qualified `model.X` names.
- **Data** (`data.go`): CSV loading for flower shop dataset. Global `shopData` variable.
- **Merchant interface** (`internal/merchant/merchant.go`): `Cataloger`, `Carter`, `Checkouter`, `Orderer` interfaces.
- **Internal packages**: `discount`, `fulfillment`, `payment`, `pricing` under `internal/merchant/` contain the business logic.

## Conventions

- After modifying any `.go` file, run `goimports -w <file>` to fix imports.
- Module: `github.com/owulveryck/ucp-merchant-test`, Go 1.22, no external dependencies.
- Error responses use JSON format: `{"detail": "message"}`.
- UCP version is `2026-01-11` everywhere (discovery, checkout `ucp` field, order `ucp` field).
- Totals types must be one of: `items_discount`, `subtotal`, `discount`, `fulfillment`, `tax`, `fee`, `total`. Never use `shipping`.
- Discount amounts in totals must be >= 0 (positive value, not negative).
- The `ucp` field in checkout/order responses is an object `{"version": "2026-01-11", "capabilities": []}`, not a string.
- Links use `type` field (e.g., `"application/json"`), not `rel`.
- `payment` is required (not optional) in checkout responses.

## Test Data Location

The conformance test data lives at:
`/path/to/Universal-Commerce-Protocol/conformance/test_data/flower_shop/`

Key test data facts:
- 6 products, `gardenias` is out of stock (quantity 0)
- Discount codes: `10OFF` (10%), `WELCOME20` (20%), `FIXED500` ($5 fixed)
- 3 customers, `cust_1` (John Doe) has 2 addresses, `cust_3` (Jane Doe) has none
- Payment tokens: `success_token` -> 200, `fail_token` -> 402
- Free shipping: orders >= $100 subtotal, or containing `bouquet_roses`
