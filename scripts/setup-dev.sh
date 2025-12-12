#!/bin/bash
# Setup dev environment with Podman

set -euo pipefail

echo "Setting up Wili dev environment..."

# Install Air for hot-reload
if [ ! -f "$(go env GOPATH)/bin/air" ]; then
  echo "Installing Air for hot-reload..."
  curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
else
  echo "Air already installed"
fi

# Install Node.js and pnpm if not available
if ! command -v node &> /dev/null; then
  echo "Installing Node.js..."
  curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
  sudo apt-get install -y nodejs
else
  echo "Node.js already installed"
fi

if ! command -v pnpm &> /dev/null; then
  echo "Installing pnpm..."
  sudo npm install -g pnpm
else
  echo "pnpm already installed"
fi

# Create network
podman network create wili-dev-network 2>/dev/null || true

# Start PostgreSQL if not exists
if ! podman container exists wili-postgres-dev; then
  podman run -d \
    --name wili-postgres-dev \
    --network wili-dev-network \
    -p 5432:5432 \
    -e POSTGRES_DB=wili_dev \
    -e POSTGRES_USER=wili \
    -e POSTGRES_PASSWORD=wili_dev_password \
    -v wili-postgres-data:/var/lib/postgresql/data \
    docker.io/postgres:15-alpine
else
  echo "PostgreSQL container already exists"
fi

# Start MongoDB if not exists
if ! podman container exists wili-mongodb-dev; then
  podman run -d \
    --name wili-mongodb-dev \
    --network wili-dev-network \
    -p 27017:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=wili \
    -e MONGO_INITDB_ROOT_PASSWORD=wili_dev_password \
    -e MONGO_INITDB_DATABASE=wili_dev \
    -v wili-mongodb-data:/data/db \
    docker.io/mongo:7
else
  echo "MongoDB container already exists"
fi

# Create user-service .env
cat > backend/services/user/.env << 'EOF'
DATABASE_URL=postgres://wili:wili_dev_password@localhost:5432/wili_dev?sslmode=disable

JWT_SECRET=dev_jwt_signing_key_change_in_production
JWT_EXPIRY_HOURS=24

YANDEX_CLIENT_ID=your_yandex_client_id_here
YANDEX_CLIENT_SECRET=your_yandex_client_secret_here

USER_SERVICE_PORT=8080
EOF

# Create wishlist-service .env
cat > backend/services/wishlist/.env << 'EOF'
MONGODB_URI=mongodb://wili:wili_dev_password@localhost:27017/wili_dev?authSource=admin
DATABASE_NAME=wili_dev

USER_SERVICE_URL=http://localhost:8080
WISHLIST_SERVICE_PORT=8081
EOF

# Create frontend .env
cat > frontend/.env << 'EOF'
PUBLIC_USER_API_BASE_URL=http://localhost:8080
PUBLIC_WISHLIST_API_BASE_URL=http://localhost:8081
PUBLIC_YANDEX_CLIENT_ID=your_yandex_client_id_here
PUBLIC_APP_URL=http://localhost:5173
EOF

echo "✅ Done! Databases running on ports 5432 and 27017"
echo "Update Yandex OAuth credentials in .env files"
echo ""
echo "To run services with hot-reload:"
echo "  ./backend/services/user/dev.sh"
echo "  ./backend/services/wishlist/dev.sh"
echo ""
echo "To run frontend:"
echo "  cd frontend && pnpm install && pnpm dev"
