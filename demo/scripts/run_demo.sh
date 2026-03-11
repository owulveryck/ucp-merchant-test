#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
DEMO_DIR="$ROOT_DIR/demo"

MA=8182; MB=8183; MC=8184; GRAPH=9000; OBS=9002
SIM_SECRET="super-secret-sim-key"

echo "=== Building binaries ==="
go build -o "$DEMO_DIR/bin/merchant" "$ROOT_DIR/sample_implementation"
go build -o "$DEMO_DIR/bin/shopping-graph" "$DEMO_DIR/cmd/shopping-graph"
go build -o "$DEMO_DIR/bin/obs-hub" "$DEMO_DIR/cmd/obs-hub"
go build -o "$DEMO_DIR/bin/client" "$DEMO_DIR/cmd/client"
echo "Build complete."

PIDS=()
cleanup() {
    echo ""
    echo "=== Shutting down ==="
    for p in "${PIDS[@]}"; do
        kill "$p" 2>/dev/null || true
    done
    wait 2>/dev/null
    echo "Done."
}
trap cleanup EXIT

echo ""
echo "=== Starting Observability Hub (port $OBS) ==="
"$DEMO_DIR/bin/obs-hub" --port $OBS &
PIDS+=($!)

echo "=== Starting SuperShop (port $MA) ==="
"$DEMO_DIR/bin/merchant" --port $MA \
    --data-dir "$DEMO_DIR/data/merchant_a" \
    --data-format json \
    --merchant-name SuperShop \
    --simulation-secret "$SIM_SECRET" &
PIDS+=($!)

echo "=== Starting MegaMart (port $MB) ==="
"$DEMO_DIR/bin/merchant" --port $MB \
    --data-dir "$DEMO_DIR/data/merchant_b" \
    --data-format json \
    --merchant-name MegaMart \
    --simulation-secret "$SIM_SECRET" &
PIDS+=($!)

echo "=== Starting BudgetBuy (port $MC) ==="
"$DEMO_DIR/bin/merchant" --port $MC \
    --data-dir "$DEMO_DIR/data/merchant_c" \
    --data-format json \
    --merchant-name BudgetBuy \
    --simulation-secret "$SIM_SECRET" &
PIDS+=($!)

sleep 2

echo "=== Starting Shopping Graph (port $GRAPH) ==="
"$DEMO_DIR/bin/shopping-graph" --port $GRAPH \
    --config "$DEMO_DIR/config/shopping_graph.yaml" \
    --obs-url "http://localhost:$OBS" &
PIDS+=($!)

sleep 3

echo ""
echo "============================================"
echo "  Dashboard: http://localhost:$OBS"
echo "  Shopping Graph: http://localhost:$GRAPH"
echo "  SuperShop: http://localhost:$MA"
echo "  MegaMart: http://localhost:$MB"
echo "  BudgetBuy: http://localhost:$MC"
echo "============================================"
echo ""

if [ -z "${GOOGLE_CLOUD_PROJECT:-}" ]; then
    echo "WARNING: GOOGLE_CLOUD_PROJECT not set. Client agent requires Vertex AI."
    echo "Set it and run: demo/bin/client --graph-url http://localhost:$GRAPH --obs-url http://localhost:$OBS"
    echo ""
    echo "Press Ctrl+C to stop all services."
    wait
else
    "$DEMO_DIR/bin/client" \
        --graph-url "http://localhost:$GRAPH" \
        --obs-url "http://localhost:$OBS"
fi
