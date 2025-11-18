# Quick Start Guide - WibuSystem Backend

## 🚀 Start Everything (Recommended)

Một lệnh để khởi động toàn bộ stack (PostgreSQL, Redis, Application, Logging):

```bash
./scripts/start-full-stack.sh
```

Script này sẽ:
- ✅ Kiểm tra và tạo OAuth2 keys
- ✅ Khởi động PostgreSQL & Redis
- ✅ Chạy database migrations
- ✅ Build và start application
- ✅ Khởi động Loki, Promtail, Grafana
- ✅ Verify tất cả services

**Thời gian:** ~2-3 phút (lần đầu build image)

---

## 📊 Access Points

Sau khi chạy script:

| Service | URL | Credentials |
|---------|-----|-------------|
| **Application** | http://localhost:8080 | - |
| **Health Check** | http://localhost:8080/health | - |
| **Grafana** | http://localhost:5555 | admin/admin |
| **Loki API** | http://localhost:3100 | - |
| **PostgreSQL** | localhost:5432 | system_dev/system_dev |
| **Redis** | localhost:6379 | (password in .env) |

---

## 🔍 Test Logging System

### 1. Generate Logs
```bash
# Make some requests to generate logs
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/some-endpoint
```

### 2. Query Logs via Loki API
```bash
# Wait 10 seconds for log ingestion, then query
curl -G http://localhost:3100/loki/api/v1/query \
  --data-urlencode 'query={container_name="wibusystem-be"}' | jq
```

### 3. View in Grafana
1. Mở http://localhost:5555
2. Login: admin/admin
3. Vào **Explore** (icon la bàn bên trái)
4. Chọn datasource **Loki**
5. Query: `{container_name="wibusystem-be"}`
6. Click **Run query**

---

## 🛠️ Common Commands

### Development
```bash
# Run application locally (không dùng Docker)
make run

# Build application
make build

# Run tests
make test

# Format code
make fmt
```

### Docker
```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# View logs
docker compose logs -f app        # Application logs
docker compose logs -f system_dev # PostgreSQL logs
docker compose logs -f            # All logs

# Rebuild application
docker compose up -d --build app

# Restart a service
docker compose restart app
```

### Database
```bash
# Open PostgreSQL shell
make db-shell

# Run migrations
make migrate-up

# Check migration version
make migrate-version

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create NAME=add_users_table
```

### Logging
```bash
# View application logs
docker logs -f wibusystem-be

# Debug logging issues
./scripts/debug-logging.sh

# Restart logging stack only
docker compose restart loki promtail grafana
```

---

## 🐛 Troubleshooting

### PostgreSQL 18 Upgrade Error

**Lỗi:**
```
database files are incompatible with server
```

**Fix:**
```bash
./scripts/fix-postgres-upgrade.sh
```

Xem thêm: [docs/POSTGRES_UPGRADE_FIX.md](docs/POSTGRES_UPGRADE_FIX.md)

---

### No Logs in Grafana

**Kiểm tra:**
```bash
# 1. Check if application is running
docker ps | grep wibusystem-be

# 2. Check if logs are in JSON format
docker logs wibusystem-be | head -1 | jq

# 3. Check Promtail is sending logs
curl http://localhost:9080/metrics | grep promtail_sent_entries_total

# 4. Run diagnostic
./scripts/debug-logging.sh
```

**Fix:**
```bash
# Restart application to ensure JSON logging
docker compose restart app

# Wait 10 seconds, then check Grafana
```

Xem thêm: [docs/LOGGING_TROUBLESHOOTING.md](docs/LOGGING_TROUBLESHOOTING.md)

---

### Application Won't Start

**Check logs:**
```bash
docker logs wibusystem-be
```

**Common issues:**

1. **Database not ready**
   ```bash
   # Wait for PostgreSQL to be healthy
   docker compose up -d system_dev
   sleep 20
   docker compose up -d app
   ```

2. **Migrations not run**
   ```bash
   make migrate-up
   docker compose restart app
   ```

3. **Environment variables missing**
   - Check `.env` file exists
   - Verify `docker-compose.override.yml` has correct values

---

## 📚 Documentation

- [Logging System Guide](docs/LOGGING.md) - Chi tiết logging system
- [Logging Quickstart](docs/LOGGING_QUICKSTART.md) - 5-phút setup guide
- [PostgreSQL Upgrade Fix](docs/POSTGRES_UPGRADE_FIX.md) - Fix PG18 issues
- [Logging Troubleshooting](docs/LOGGING_TROUBLESHOOTING.md) - Debug guide

---

## 🎯 Next Steps

Sau khi chạy `./scripts/start-full-stack.sh`:

1. **Verify application works:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check logs in Grafana:**
   - http://localhost:5555
   - Explore → Loki → `{container_name="wibusystem-be"}`

3. **Integrate logging into code:**
   - See [docs/LOGGING_QUICKSTART.md](docs/LOGGING_QUICKSTART.md)
   - Add audit logging for OAuth2 events
   - Add performance logging for slow operations

4. **Set up OAuth2:**
   - Keys already generated in `configs/keys/`
   - Update `.env` with production secrets
   - Test with `./scripts/test_oauth2_flow.sh`

5. **Develop features:**
   ```bash
   # Make changes to code
   # Rebuild and restart
   docker compose up -d --build app

   # View logs in real-time
   docker logs -f wibusystem-be
   ```

---

## 💡 Tips

- **JSON Logging:** Application MUST output JSON logs for Loki to parse properly
- **Log Levels:** Use appropriate levels (DEBUG, INFO, WARN, ERROR) for filtering
- **Structured Logging:** Include request_id, user_id in log context for tracing
- **Performance:** Monitor query durations and add indexes as needed
- **Security:** Never log passwords, tokens, or sensitive data

---

## 🆘 Need Help?

1. Check documentation in `docs/` directory
2. Run diagnostic: `./scripts/debug-logging.sh`
3. View logs: `docker compose logs -f`
4. Check container status: `docker compose ps`

---

**Happy coding! 🚀**
