#!/bin/bash

# Memoriva Backend Deployment Script
set -e

echo "🚀 Starting Memoriva Backend deployment..."

# Configuration
SERVICE_NAME="memoriva-backend"
DEPLOY_DIR="/opt/memoriva-backend"
SERVICE_USER="memoriva"
BINARY_NAME="memoriva-backend"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "Please run this script as root (use sudo)"
    exit 1
fi

# Create service user if it doesn't exist
if ! id "$SERVICE_USER" &>/dev/null; then
    print_status "Creating service user: $SERVICE_USER"
    useradd --system --shell /bin/false --home-dir $DEPLOY_DIR --create-home $SERVICE_USER
else
    print_status "Service user $SERVICE_USER already exists"
fi

# Create deployment directory
print_status "Creating deployment directory: $DEPLOY_DIR"
mkdir -p $DEPLOY_DIR
mkdir -p $DEPLOY_DIR/uploads
mkdir -p $DEPLOY_DIR/logs

# Build the Go application
print_status "Building Go application..."
go build -o $BINARY_NAME .

# Copy files to deployment directory
print_status "Copying files to deployment directory..."
cp $BINARY_NAME $DEPLOY_DIR/
cp .env.example $DEPLOY_DIR/.env
cp -r uploads/* $DEPLOY_DIR/uploads/ 2>/dev/null || true

# Set proper permissions
print_status "Setting file permissions..."
chown -R $SERVICE_USER:$SERVICE_USER $DEPLOY_DIR
chmod +x $DEPLOY_DIR/$BINARY_NAME
chmod 600 $DEPLOY_DIR/.env
chmod 755 $DEPLOY_DIR/uploads

# Install systemd service
print_status "Installing systemd service..."
cp $SERVICE_NAME.service /etc/systemd/system/
systemctl daemon-reload

# Enable and start the service
print_status "Enabling and starting the service..."
systemctl enable $SERVICE_NAME
systemctl restart $SERVICE_NAME

# Check service status
sleep 2
if systemctl is-active --quiet $SERVICE_NAME; then
    print_status "✅ Service is running successfully!"
    systemctl status $SERVICE_NAME --no-pager -l
else
    print_error "❌ Service failed to start. Check logs with: journalctl -u $SERVICE_NAME -f"
    exit 1
fi

echo ""
print_status "🎉 Deployment completed successfully!"
echo ""
print_warning "⚠️  IMPORTANT: Don't forget to:"
echo "   1. Edit $DEPLOY_DIR/.env with your actual configuration"
echo "   2. Restart the service after updating .env: sudo systemctl restart $SERVICE_NAME"
echo "   3. Check logs with: sudo journalctl -u $SERVICE_NAME -f"
echo "   4. Monitor service status: sudo systemctl status $SERVICE_NAME"
echo ""
print_status "Service will be available on port 8080 (or your configured PORT)"
