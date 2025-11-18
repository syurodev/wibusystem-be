# Logging Setup Options - Choose Your Path

## Vấn Đề Hiện Tại

✅ Loki: Running (port 3100)
❌ Promtail: Running nhưng không có port 9080 (ĐÃ FIX)
❌ Application: **KHÔNG CÓ CONTAINER** → Promtail không có gì để scrape!

---

## 🎯 3 Options Để Chạy Application Với Logging

### Option 1: Run App Trong Docker (RECOMMENDED)

**Ưu điểm**:
- Promtail tự động scrape logs từ Docker containers
- Không cần config thêm
- Production-like environment

**Bước 1: Restart Promtail với port mới**

```bash
./scripts/restart-logging.sh
```

**Bước 2: Copy docker-compose override**

```bash
cp docker-compose.override.example.yml docker-compose.override.yml
```

**Bước 3: Update environment variables trong override file**

Edit `docker-compose.override.yml`:
```yaml
services:
  wibusystem-be:
    environment:
      - OAUTH2_ISSUER=http://localhost:8080
      - OAUTH2_HMAC_SECRET=your-actual-secret-here  # Change this!
      - OAUTH2_PRIVATE_KEY_PATH=/app/configs/keys/private_key.pem
```

**Bước 4: Tạo RSA keys nếu chưa có**

```bash
mkdir -p configs/keys
openssl genrsa -out configs/keys/private_key.pem 2048
openssl rsa -in configs/keys/private_key.pem -pubout -out configs/keys/public_key.pem
```

**Bước 5: Start application container**

```bash
docker-compose up -d wibusystem-be
```

**Bước 6: Verify logs**

```bash
# Check container is running
docker ps | grep wibusystem-be

# Check logs are JSON
docker logs wibusystem-be --tail 5

# Check Promtail is collecting
curl http://localhost:9080/targets | jq '.activeTargets[] | select(.labels.container_name == "wibusystem-be")'

# Check logs in Grafana
# Go to http://localhost:5555 → Explore → Query: {container_name="wibusystem-be"}
```

---

### Option 2: Run App Locally + File-Based Logging

**Ưu điểm**:
- Faster development (no rebuild)
- Easier debugging

**Nhược điểm**:
- Cần config thêm file scraping

**Bước 1: Restart Promtail**

```bash
./scripts/restart-logging.sh
```

**Bước 2: Configure app để write logs ra file**

Update `cmd/server/main.go`:

```go
package main

import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func main() {
    // ... load config ...

    // Create logs directory
    os.MkdirAll("logs", 0755)

    // Configure Zap to write to both stdout AND file
    file, _ := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

    fileWriter := zapcore.AddSync(file)
    consoleWriter := zapcore.AddSync(os.Stdout)

    config := zap.NewProductionConfig()
    config.Encoding = "json"

    core := zapcore.NewTee(
        zapcore.NewCore(
            zapcore.NewJSONEncoder(config.EncoderConfig),
            fileWriter,
            zapcore.InfoLevel,
        ),
        zapcore.NewCore(
            zapcore.NewJSONEncoder(config.EncoderConfig),
            consoleWriter,
            zapcore.InfoLevel,
        ),
    )

    appLogger := zap.New(core)
    defer appLogger.Sync()

    // ... rest of your code ...
}
```

**Bước 3: Update Promtail config để scrape files**

Edit `promtail-config.yml`, thêm vào `scrape_configs`:

```yaml
scrape_configs:
  - job_name: docker
    # ... existing docker config ...

  # NEW: File-based scraping
  - job_name: local-app
    static_configs:
      - targets:
          - localhost
        labels:
          job: wibusystem-be
          environment: development
          __path__: /logs/app.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: ts
            message: msg
            category: category
      - timestamp:
          source: timestamp
          format: RFC3339Nano
      - labels:
          level:
          category:
```

**Bước 4: Update docker-compose.yml Promtail volumes**

```yaml
promtail:
  volumes:
    - /var/run/docker.sock:/var/run/docker.sock
    - ./promtail-config.yml:/etc/promtail/promtail-config.yml
    - ./logs:/logs:ro  # ADD THIS LINE
```

**Bước 5: Restart Promtail**

```bash
docker-compose restart promtail
```

**Bước 6: Run app locally**

```bash
go run ./cmd/server/main.go
```

**Bước 7: Verify**

```bash
# Check log file is created
ls -lh logs/app.log

# Check logs are JSON
tail -1 logs/app.log | jq .

# Check Promtail targets
curl http://localhost:9080/targets | jq '.activeTargets[] | select(.labels.job == "wibusystem-be")'

# Check in Grafana
# Query: {job="wibusystem-be"}
```

---

### Option 3: Hybrid (Docker Services + Local App via Stdout)

**Chỉ dùng cho testing nhanh**

**Bước 1: Run app locally với JSON logs**

```bash
# Make sure logger outputs JSON
go run ./cmd/server/main.go 2>&1 | tee logs/app.log
```

**Bước 2: Use a log forwarder**

Install và chạy Promtail locally:

```bash
# Download Promtail binary
wget https://github.com/grafana/loki/releases/download/v3.0.0/promtail-linux-amd64.zip
unzip promtail-linux-amd64.zip

# Create local config
cat > promtail-local.yml <<EOF
server:
  http_listen_port: 9081

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  - job_name: local-app
    static_configs:
      - targets:
          - localhost
        labels:
          job: wibusystem-be-local
          __path__: $PWD/logs/app.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            ts: ts
            msg: msg
      - labels:
          level:
EOF

# Run Promtail
./promtail-linux-amd64 -config.file=promtail-local.yml
```

---

## 🚀 Quick Start (Recommended Path)

**Nếu bạn muốn setup nhanh nhất**:

```bash
# 1. Fix Promtail port
./scripts/restart-logging.sh

# 2. Test với một container đơn giản
docker run -d \
  --name test-logger \
  --network wibusystem-backend \
  alpine sh -c 'while true; do echo "{\"level\":\"info\",\"ts\":\"$(date -Iseconds)\",\"msg\":\"Test log from Alpine\",\"category\":\"test\"}"; sleep 2; done'

# 3. Wait 10 seconds for Promtail to pick up

# 4. Check Grafana
# Go to http://localhost:5555 → Explore
# Query: {container_name="test-logger"}
# You should see logs!

# 5. Stop test container
docker stop test-logger && docker rm test-logger

# 6. Now add your real application using Option 1 above
```

---

## ✅ Verification Checklist

Run these commands to verify setup:

```bash
# 1. All services running?
docker-compose ps

# 2. Promtail accessible?
curl http://localhost:9080/ready

# 3. Promtail has targets?
curl http://localhost:9080/targets | jq '.activeTargets | length'
# Should be > 0

# 4. Promtail sending data?
curl http://localhost:9080/metrics | grep promtail_sent_entries_total
# Should have values > 0

# 5. Loki receiving data?
curl -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={job=~".+"}' \
  --data-urlencode 'limit=1' | jq '.data.result | length'
# Should be > 0

# 6. Check in Grafana Explore
# http://localhost:5555 → Explore → {container_name=~".+"}
```

---

## 🐛 Troubleshooting

### Issue: Promtail targets = 0

**Check Docker socket permission**:
```bash
ls -l /var/run/docker.sock
# Should allow docker group to read

# If not, add to docker group (Linux)
sudo usermod -aG docker $USER
# Then logout/login
```

### Issue: Logs not appearing in Loki

**Check log format**:
```bash
docker logs <container-name> --tail 1 | jq .
# Should parse as valid JSON
```

### Issue: Can't access Promtail port 9080

**Restart with new compose file**:
```bash
docker-compose down
docker-compose up -d
```

---

## 📊 Current Status After Fix

```
✅ docker-compose.yml updated with Promtail port mapping
✅ Dockerfile created for containerized app
✅ docker-compose.override.example.yml ready
✅ Restart script created

⚠️ NEXT: Choose an option above and implement it
```

---

## Need Help?

1. Start with the Quick Start test container
2. If that works, you know Promtail → Loki → Grafana is working
3. Then add your actual application using Option 1, 2, or 3
4. Run verification checklist
5. If stuck, run `./scripts/debug-logging.sh` and share output
