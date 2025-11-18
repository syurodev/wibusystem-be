#!/bin/bash
# Script to restart logging services and verify setup

echo "==================================================="
echo "RESTARTING LOGGING SERVICES"
echo "==================================================="
echo ""

echo "1. Stopping Promtail..."
docker-compose stop promtail

echo "2. Starting Promtail with new config..."
docker-compose up -d promtail

echo "3. Waiting for Promtail to start..."
sleep 3

echo ""
echo "==================================================="
echo "VERIFICATION"
echo "==================================================="
echo ""

echo "✓ Docker Services Status:"
docker-compose ps promtail loki grafana

echo ""
echo "✓ Promtail Health Check:"
curl -s http://localhost:9080/ready && echo " - Promtail Ready ✓" || echo " - Promtail Not Ready ✗"

echo ""
echo "✓ Loki Health Check:"
curl -s http://localhost:3100/ready && echo " - Loki Ready ✓" || echo " - Loki Not Ready ✗"

echo ""
echo "✓ Promtail Targets:"
curl -s http://localhost:9080/targets | jq -r '.activeTargets[] | "  - Container: \(.labels.container_name // "N/A") | Job: \(.labels.job)"' 2>/dev/null || echo "  No active targets yet"

echo ""
echo "✓ Promtail Metrics (Entries Sent):"
curl -s http://localhost:9080/metrics 2>/dev/null | grep "promtail_sent_entries_total" | head -3 || echo "  Waiting for metrics..."

echo ""
echo "==================================================="
echo "NEXT STEPS"
echo "==================================================="
echo ""
echo "1. Check if you have any application containers running:"
echo "   docker ps --format 'table {{.Names}}\t{{.Image}}'"
echo ""
echo "2. If you don't have an app container, logs won't appear."
echo "   You need to either:"
echo "   - Run your Go app in a Docker container"
echo "   - Or run locally and configure file-based log scraping"
echo ""
echo "3. Test with a simple container:"
echo "   docker run --rm --name test-app --network wibusystem-backend alpine sh -c 'while true; do echo \"{\\\"level\\\":\\\"info\\\",\\\"msg\\\":\\\"test log\\\"}\"; sleep 2; done'"
echo ""
echo "4. Check Grafana Explore:"
echo "   http://localhost:5555 → Explore → Query: {container_name=~\".+\"}"
echo ""
