#!/bin/bash
# Script để debug logging system

echo "==================================================="
echo "LOGGING SYSTEM DIAGNOSTIC"
echo "==================================================="
echo ""

echo "1. Checking Docker Services..."
echo "---------------------------------------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "loki|promtail|grafana|wibusystem"
echo ""

echo "2. Checking Loki Health..."
echo "---------------------------------------------------"
curl -s http://localhost:3100/ready || echo "❌ Loki not ready"
curl -s http://localhost:3100/metrics | grep -E "loki_ingester_chunks_created_total|loki_distributor_bytes_received_total" | head -5
echo ""

echo "3. Checking Promtail Targets..."
echo "---------------------------------------------------"
curl -s http://localhost:9080/targets | jq '.activeTargets[] | {job: .labels.job, container: .labels.container_name, lastScrape: .lastScrape}'
echo ""

echo "4. Checking Promtail Metrics..."
echo "---------------------------------------------------"
curl -s http://localhost:9080/metrics | grep -E "promtail_sent_entries_total|promtail_dropped_entries_total|promtail_read_bytes_total"
echo ""

echo "5. Sample Application Logs (last 5 lines)..."
echo "---------------------------------------------------"
docker logs --tail 5 wibusystem-be 2>&1
echo ""

echo "6. Check Log Format (first line)..."
echo "---------------------------------------------------"
docker logs --tail 1 wibusystem-be 2>&1 | jq . || echo "❌ Logs are not in JSON format"
echo ""

echo "7. Promtail Logs (last 10 lines)..."
echo "---------------------------------------------------"
docker logs --tail 10 promtail 2>&1
echo ""

echo "8. Loki Logs (last 10 lines)..."
echo "---------------------------------------------------"
docker logs --tail 10 loki 2>&1
echo ""

echo "9. Test Loki Query..."
echo "---------------------------------------------------"
curl -s -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={container_name=~".+"}' \
  --data-urlencode 'limit=5' | jq '.data.result | length'
echo " log streams found"
echo ""

echo "==================================================="
echo "DIAGNOSTIC COMPLETE"
echo "==================================================="
