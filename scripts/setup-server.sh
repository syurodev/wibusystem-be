#!/bin/bash
# =============================================================================
# Setup Script cho VPS Debian 12 - WibuSystem Backend
# Chạy: curl -fsSL <url> | bash hoặc bash setup-server.sh
# =============================================================================

set -e

echo "=========================================="
echo "  WibuSystem Backend - Server Setup"
echo "  Debian 12 | Docker | Docker Compose"
echo "=========================================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Vui lòng chạy với quyền root (sudo)${NC}"
    exit 1
fi

echo -e "${YELLOW}[1/6] Cập nhật hệ thống...${NC}"
apt-get update -y
apt-get upgrade -y

echo -e "${YELLOW}[2/6] Cài đặt các packages cần thiết...${NC}"
apt-get install -y \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    htop \
    vim \
    ufw \
    fail2ban

echo -e "${YELLOW}[3/6] Cài đặt Docker...${NC}"
# Xóa docker cũ nếu có
apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true

# Add Docker GPG key
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

# Add Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update -y
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Start Docker
systemctl enable docker
systemctl start docker

echo -e "${YELLOW}[4/6] Cấu hình Firewall (UFW)...${NC}"
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp   # HTTP
ufw allow 443/tcp  # HTTPS
ufw allow 8080/tcp # Backend API (optional, có thể bỏ nếu dùng Cloudflare proxy)
echo "y" | ufw enable

echo -e "${YELLOW}[5/6] Cấu hình Fail2ban...${NC}"
systemctl enable fail2ban
systemctl start fail2ban

echo -e "${YELLOW}[6/6] Tạo thư mục project...${NC}"
mkdir -p /opt/wibusystem
mkdir -p /opt/wibusystem/configs/keys
cd /opt/wibusystem

# Tạo file .env rỗng
touch .env
chmod 600 .env

echo ""
echo -e "${GREEN}=========================================="
echo "  ✓ Setup hoàn tất!"
echo "==========================================${NC}"
echo ""
echo "Các bước tiếp theo:"
echo "1. Copy .env.example -> .env và cập nhật các giá trị"
echo "   cp /opt/wibusystem/.env.example /opt/wibusystem/.env"
echo "   nano /opt/wibusystem/.env"
echo ""
echo "2. Login Docker Hub:"
echo "   docker login"
echo ""
echo "3. Pull và chạy containers (sau khi setup GitHub Actions):"
echo "   cd /opt/wibusystem"
echo "   docker compose -f docker-compose.production.yml up -d"
echo ""
echo "Server info:"
echo "  - Docker: $(docker --version)"
echo "  - Docker Compose: $(docker compose version)"
echo "  - Firewall: UFW enabled (22, 80, 443, 8080)"
echo ""
