# Multi-Agent Shopping Demo

A multi-agent shopping demo showcasing UCP and A2A protocols. Three merchant instances, a Shopping Graph for cross-merchant search, a Gemini-powered Client Agent, and an Observability Hub.

## Architecture

```
Client Agent (Gemini) ──> Shopping Graph ──> Merchant A (SuperShop :8182)
                     │                  ├──> Merchant B (MegaMart  :8183)
                     │                  └──> Merchant C (BudgetBuy :8184)
                     └──> Obs Hub (:9002)
```

- **Merchants**: Three instances of the UCP merchant server with different product catalogs and pricing
- **Shopping Graph** (:9000): Polls merchants via A2A, indexes products, provides cross-merchant search with Jaccard similarity matching
- **Client Agent**: Gemini-powered agent that searches, compares prices, creates checkouts, applies discounts, and completes the cheapest order
- **Observability Hub** (:9002): Real-time event dashboard with SSE

## Prerequisites

```bash
# Vertex AI auth
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT=your-project-id

# Optional
export GOOGLE_CLOUD_LOCATION=us-central1  # default
```

## Quick Start

```bash
demo/scripts/run_demo.sh
```

Or step by step:

```bash
# Build
go build -o demo/bin/merchant ./sample_implementation
go build -o demo/bin/shopping-graph ./demo/cmd/shopping-graph
go build -o demo/bin/obs-hub ./demo/cmd/obs-hub
go build -o demo/bin/client ./demo/cmd/client

# Start services (in separate terminals)
demo/bin/obs-hub --port 9002
demo/bin/merchant --port 8182 --data-dir demo/data/merchant_a --data-format json --merchant-name SuperShop
demo/bin/merchant --port 8183 --data-dir demo/data/merchant_b --data-format json --merchant-name MegaMart
demo/bin/merchant --port 8184 --data-dir demo/data/merchant_c --data-format json --merchant-name BudgetBuy
demo/bin/shopping-graph --port 9000 --config demo/config/shopping_graph.yaml --obs-url http://localhost:9002

# Run client agent
demo/bin/client --graph-url http://localhost:9000 --obs-url http://localhost:9002
```

## Example Interaction

```
> Find me wireless headphones at the best price

Searching for "wireless headphones" across all merchants...

Found 3 results:
1. BudgetBuy - Wireless Audio Headphones - $79.99 (discount hints: BUDGET20, SAVE5)
2. MegaMart - Bluetooth Headset Pro - $84.99 (discount hints: MEGA10)
3. SuperShop - Wireless Headphones - $89.99 (discount hints: SAVE10, WELCOME15)

Creating checkouts at BudgetBuy and SuperShop...
Applying discount BUDGET20 at BudgetBuy... 20% off!
Applying discount WELCOME15 at SuperShop... 15% off!

Price comparison:
  BudgetBuy:  $63.99 + $5.99 shipping = $69.98
  SuperShop:  $76.49 + $0.00 shipping = $76.49 (free shipping over $100 not met)

Completing order at BudgetBuy (cheaper by $6.51)
Cancelling checkout at SuperShop

Order placed! Order ID: ord_xxx
```

## Merchant Data

| Merchant | Port | Products | Discount Codes | Free Shipping |
|----------|------|----------|---------------|---------------|
| SuperShop | 8182 | 6 (1 OOS) | SAVE10 (10%), WELCOME15 (15%) | >= $100 |
| MegaMart | 8183 | 5 (1 OOS) | MEGA10 (10%) | >= $150 |
| BudgetBuy | 8184 | 5 (1 OOS) | BUDGET20 (20%), SAVE5 ($5 off) | >= $80 |
