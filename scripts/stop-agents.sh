#!/bin/bash
# Stop all A2A agents

echo "🛑 Stopping A2A Agents..."

if [ -f /tmp/customer-growth.pid ]; then
    PID=$(cat /tmp/customer-growth.pid)
    kill $PID 2>/dev/null && echo "  ✅ Customer Growth Agent stopped (PID: $PID)" || echo "  ⚠️  Customer Growth Agent not running"
    rm /tmp/customer-growth.pid
fi

if [ -f /tmp/competitiveness.pid ]; then
    PID=$(cat /tmp/competitiveness.pid)
    kill $PID 2>/dev/null && echo "  ✅ Competitiveness Agent stopped (PID: $PID)" || echo "  ⚠️  Competitiveness Agent not running"
    rm /tmp/competitiveness.pid
fi

if [ -f /tmp/agents-dashboard.pid ]; then
    PID=$(cat /tmp/agents-dashboard.pid)
    kill $PID 2>/dev/null && echo "  ✅ Agents Dashboard stopped (PID: $PID)" || echo "  ⚠️  Agents Dashboard not running"
    rm /tmp/agents-dashboard.pid
fi

echo ""
echo "✅ All agents stopped"
