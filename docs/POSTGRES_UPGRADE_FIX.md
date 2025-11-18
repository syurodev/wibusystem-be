# PostgreSQL Upgrade Fix

## 🔴 Lỗi

```
FATAL: database files are incompatible with server
This is usually the result of upgrading the Docker image without
upgrading the underlying database using "pg_upgrade"
```

## 🎯 Nguyên Nhân

Docker volume `system_data` chứa data từ **PostgreSQL version cũ** (17 hoặc thấp hơn), không tương thích với **PostgreSQL 18**.

---

## ✅ GIẢI PHÁP - Chọn 1 trong 2

### Option 1: Automated Script (RECOMMENDED)

**Dùng khi**: Development, không cần data cũ

```bash
./scripts/fix-postgres-upgrade.sh
```

Script sẽ tự động:
1. ✅ Stop containers
2. ✅ Remove old PostgreSQL volume
3. ✅ Start PostgreSQL 18 fresh
4. ✅ Run migrations
5. ✅ Verify setup

**Time**: ~2 phút

---

### Option 2: Manual Steps

**Step 1: Stop containers**
```bash
docker-compose down
```

**Step 2: Remove old volume**
```bash
docker volume rm wibusystem-backend_system_data
```

**Step 3: Start PostgreSQL**
```bash
docker-compose up -d system_dev
```

**Step 4: Wait for ready (20 seconds)**
```bash
sleep 20
```

**Step 5: Verify PostgreSQL**
```bash
docker-compose exec system_dev psql -U system_dev -d system_dev -c "SELECT version();"
```

Should show: `PostgreSQL 18.0`

**Step 6: Run migrations**
```bash
# Install migrate tool if not already
make migrate-install

# Run migrations
make migrate-up
```

**Step 7: Verify tables created**
```bash
make db-shell
```

Then in psql:
```sql
\dt identify.*
\dt catalog.*
\dt community.*
\dt payment.*
```

Should see all tables from migrations.

**Step 8: Start all services**
```bash
docker-compose up -d
```

---

## 🔍 Verification

```bash
# Check PostgreSQL version
docker-compose exec system_dev psql -U system_dev -d system_dev -c "SELECT version();"

# Check tables exist
docker-compose exec system_dev psql -U system_dev -d system_dev -c "\dt identify.*"

# Check migrations applied
make migrate-version
```

---

## 📊 Quick Comparison

| Method | Time | Data Preserved | Difficulty |
|--------|------|----------------|------------|
| **Script** | 2 min | ❌ No | ⭐ Easy |
| **Manual** | 5 min | ❌ No | ⭐⭐ Medium |
| **pg_upgrade** | 30+ min | ✅ Yes | ⭐⭐⭐⭐⭐ Expert |

---

## ⚠️ Important Notes

### Data Loss
**BOTH options will DELETE all PostgreSQL data!**

If you need to preserve data, you must:
1. Backup data before upgrade
2. Use `pg_upgrade` (complex, see official docs)
3. Or export/import manually

### Development vs Production

**Development** (recommended for this case):
- Use automated script
- Quick and clean
- Re-run migrations
- Seed test data

**Production** (not covered here):
- Must use `pg_upgrade`
- Must test in staging first
- Must have backup strategy
- See: https://www.postgresql.org/docs/current/pgupgrade.html

---

## 🚀 After Fix - Next Steps

### 1. Generate OAuth2 Keys
```bash
mkdir -p configs/keys
openssl genrsa -out configs/keys/private_key.pem 2048
openssl rsa -in configs/keys/private_key.pem -pubout -out configs/keys/public_key.pem
```

### 2. Update .env
```bash
# Check .env has these (minimum 32 chars for HMAC secret)
OAUTH2_ISSUER=http://localhost:8080
OAUTH2_PRIVATE_KEY_PATH=configs/keys/private_key.pem
OAUTH2_KEY_ID=oauth2-key-1
OAUTH2_HMAC_SECRET=your-secret-at-least-32-characters-long
```

### 3. Seed Test Data (Optional)
Create users, OAuth2 clients, etc.

### 4. Run Application
```bash
make run
```

### 5. Test
```bash
# Health check
curl http://localhost:8080/health

# Check database connection
docker logs <app-container> | grep -i "database"
```

---

## 🐛 Troubleshooting

### Issue: Script hangs at "Waiting for PostgreSQL"

**Check logs**:
```bash
docker logs system_dev
```

**Common causes**:
- Port 5432 already in use
- Permission issues with volume
- Docker daemon issues

**Fix**:
```bash
# Kill any process using port 5432
sudo lsof -ti:5432 | xargs kill -9

# Restart Docker
sudo systemctl restart docker  # Linux
# or restart Docker Desktop on Mac/Windows

# Try again
./scripts/fix-postgres-upgrade.sh
```

### Issue: Migrations fail

**Check migrate tool installed**:
```bash
make migrate-install
```

**Check database reachable**:
```bash
docker-compose exec system_dev psql -U system_dev -d system_dev -c "SELECT 1;"
```

**Check migrations directory**:
```bash
ls -la migrations/
```

**Manual migration**:
```bash
# Force to version 0 (clean slate)
make migrate-force VERSION=0

# Re-run all migrations
make migrate-up
```

### Issue: "role does not exist" error

**Create role manually**:
```bash
docker-compose exec system_dev psql -U postgres -c "CREATE ROLE system_dev WITH LOGIN PASSWORD 'system_dev';"
docker-compose exec system_dev psql -U postgres -c "CREATE DATABASE system_dev OWNER system_dev;"
```

---

## 📚 Related Docs

- [PostgreSQL Upgrade Guide](https://www.postgresql.org/docs/current/pgupgrade.html)
- [Docker PostgreSQL Image](https://hub.docker.com/_/postgres)
- [golang-migrate Docs](https://github.com/golang-migrate/migrate)

---

## ❓ FAQ

**Q: Will I lose all my data?**
A: Yes, if using the script or manual steps. Use `pg_upgrade` to preserve data (advanced).

**Q: Can I downgrade to PostgreSQL 17?**
A: Yes, change `image: postgres:17` in docker-compose.yml and restart.

**Q: How to prevent this in the future?**
A: Use named migrations and backup data before upgrading PostgreSQL version.

**Q: What if I need my data back?**
A: If you have a backup, restore it after fix. Otherwise, data is lost.

**Q: Is this safe for production?**
A: **NO!** This is for development only. Production needs proper upgrade procedure.
