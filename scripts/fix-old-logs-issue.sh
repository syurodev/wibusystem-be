#!/bin/bash
# Fix old logs issue - Restart containers to generate fresh logs

echo "==================================================="
echo "FIX OLD LOGS TIMESTAMP ISSUE"
echo "==================================================="
echo ""

echo "Issue: Loki is rejecting logs with old timestamps from"
echo "containers that have been running since yesterday."
echo ""
echo "Solution: Restart all containers to generate fresh logs."
echo ""

read -p "Continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Aborted."
    exit 1
fi

echo ""
echo "1. Stopping all containers..."
docker-compose down

echo ""
echo "2. Starting services..."
docker-compose up -d

echo ""
echo "3. Waiting for services to be ready (15 seconds)..."
sleep 15

echo ""
echo "==================================================="
echo "VERIFICATION"
echo "==================================================="
echo ""

echo "✓ Loki status:"
curl -s http://localhost:3100/ready && echo " - Ready ✓" || echo " - Not ready ✗"

echo ""
echo "✓ Promtail status:"
curl -s http://localhost:9080/ready && echo " - Ready ✓" || echo " - Not ready ✗"

echo ""
echo "✓ Promtail targets (should see containers):"
curl -s http://localhost:9080/targets 2>/dev/null | grep -o '"container_name":"[^"]*"' | sort -u || echo "  Waiting for targets..."

echo ""
echo "✓ Wait 30 seconds for logs to flow..."
sleep 30

echo ""
echo "✓ Check dropped entries (should be 0):"
curl -s http://localhost:9080/metrics 2>/dev/null | grep 'promtail_dropped_entries_total{.*reason="ingester_error"'

echo ""
echo "✓ Check sent entries (should be > 0):"
curl -s http://localhost:9080/metrics 2>/dev/null | grep 'promtail_sent_entries_total'

echo ""
echo "==================================================="
echo "NEXT STEPS"
echo "==================================================="
echo ""
echo "1. Run test container:"
echo "   docker run -d --name test-logger --network wibusystem-backend \\"
echo "     alpine sh -c 'while true; do echo \"{\\\"level\\\":\\\"info\\\",\\\"ts\\\":\\\"\$(date -Iseconds)\\\",\\\"msg\\\":\\\"Test log\\\",\\\"category\\\":\\\"test\\\"}\"; sleep 2; done'"
echo ""
echo "2. Wait 10 seconds, then check Grafana:"
echo "   http://localhost:5555 → Explore → {container_name=\"test-logger\"}"
echo ""
echo "3. Clean up test:"
echo "   docker stop test-logger && docker rm test-logger"
echo ""
