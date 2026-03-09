# CLAUDE.md

## Project Overview

UCP merchant test server written in Go. Implements the Universal Commerce Protocol (UCP) Shopping Service with both REST and MCP (JSON-RPC) transports. Designed to pass the UCP conformance test suite (60 tests across 13 files).

## Build & Run

```bash
# Build
go build .

# Run with flower shop test data
go run . --port 8182 \
  --data-dir /path/to/conformance/test_data/flower_shop \
  --simulation-secret super-secret-sim-key

# Run conformance tests (from the conformance directory)
python3 <test_file>.py \
  --server_url=http://localhost:8182 \
  --simulation_secret=super-secret-sim-key \
  --conformance_input=test_data/flower_shop/conformance_input.json \
  --test_data_dir=test_data/flower_shop
```

## Key Architecture

- **REST layer** (`rest.go`): All checkout session and order CRUD. Routes registered in `main.go`.
- **Models** (`models.go`): Go structs matching UCP SDK Pydantic models. Prefix `Rest` on type names.
- **Data** (`data.go`): CSV loading for flower shop dataset. Global `shopData` variable.
- **MCP layer** (`handlers.go`): JSON-RPC tool handlers (separate from REST stores).
- Two separate in-memory stores: `restCheckouts`/`restOrders` (REST) and `checkouts`/`orders` (MCP).

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
