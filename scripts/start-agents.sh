#!/bin/bash
# Start all A2A agents

set -e

cd "$(dirname "$0")/.."

echo "🚀 Starting A2A Agents..."
echo ""

# Check if binaries exist
if [ ! -f "bin/customer-growth-agent" ] || [ ! -f "bin/competitiveness-agent" ] || [ ! -f "bin/agents-dashboard" ]; then
    echo "❌ Binaries not found. Building..."
    go build -o bin/customer-growth-agent ./cmd/customer-growth-agent
    go build -o bin/competitiveness-agent ./cmd/competitiveness-agent
    go build -o bin/agents-dashboard ./cmd/agents-dashboard
    echo "✅ Build complete"
    echo ""
fi

# Start Customer Growth Agent
echo "Starting Customer Growth Agent on port 9001..."
bin/customer-growth-agent --port 9001 > logs/customer-growth.log 2>&1 &
CG_PID=$!
echo $CG_PID > /tmp/customer-growth.pid
echo "  PID: $CG_PID"

# Start Competitiveness Agent
echo "Starting Competitiveness Agent on port 9002..."
bin/competitiveness-agent --port 9002 > logs/competitiveness.log 2>&1 &
COMP_PID=$!
echo $COMP_PID > /tmp/competitiveness.pid
echo "  PID: $COMP_PID"

# Start Dashboard
echo "Starting Agents Dashboard on port 8080..."
bin/agents-dashboard --port 8080 \
  --customer-growth-url http://localhost:9001 \
  --competitiveness-url http://localhost:9002 > logs/agents-dashboard.log 2>&1 &
DASH_PID=$!
echo $DASH_PID > /tmp/agents-dashboard.pid
echo "  PID: $DASH_PID"

echo ""
echo "✅ All agents started!"
echo ""
echo "📍 Endpoints:"
echo "  • Dashboard:           http://localhost:8080"
echo "  • Customer Growth:     http://localhost:9001"
echo "  • Competitiveness:     http://localhost:9002"
echo ""
echo "📝 Logs:"
echo "  • tail -f logs/customer-growth.log"
echo "  • tail -f logs/competitiveness.log"
echo "  • tail -f logs/agents-dashboard.log"
echo ""
echo "🛑 To stop: ./scripts/stop-agents.sh"
echo ""

# Wait a moment for agents to start
sleep 2

# Health check
echo "🔍 Health check..."
if curl -s http://localhost:9001/health > /dev/null; then
    echo "  ✅ Customer Growth: healthy"
else
    echo "  ❌ Customer Growth: failed"
fi

if curl -s http://localhost:9002/health > /dev/null; then
    echo "  ✅ Competitiveness: healthy"
else
    echo "  ❌ Competitiveness: failed"
fi

echo ""
echo "🎉 Ready! Open http://localhost:8080 in your browser"
