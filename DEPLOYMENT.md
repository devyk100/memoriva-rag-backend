# Memoriva Backend - Systemd Deployment Guide

This guide explains how to deploy the Memoriva RAG Backend as a systemd daemon on your VPS.

## Prerequisites

- Go 1.19+ installed on your VPS
- PostgreSQL database running
- Root/sudo access on your VPS

## Quick Deployment

1. **Make the deployment script executable:**
   ```bash
   chmod +x deploy.sh
   ```

2. **Run the deployment script:**
   ```bash
   sudo ./deploy.sh
   ```

3. **Configure your environment:**
   ```bash
   sudo nano /opt/memoriva-backend/.env
   ```
   Update the values according to your setup:
   ```env
   DATABASE_URL=postgresql://postgres:your_password@localhost:5432/memoriva?sslmode=disable
   DEEPSEEK_API_KEY=your_actual_deepseek_api_key
   OPENAI_API_KEY=your_actual_openai_api_key
   PORT=8080
   ```

4. **Restart the service:**
   ```bash
   sudo systemctl restart memoriva-backend
   ```

## Manual Deployment Steps

If you prefer to deploy manually:

### 1. Build the Application
```bash
go build -o memoriva-backend .
```

### 2. Create Service User
```bash
sudo useradd --system --shell /bin/false --home-dir /opt/memoriva-backend --create-home memoriva
```

### 3. Create Deployment Directory
```bash
sudo mkdir -p /opt/memoriva-backend/{uploads,logs}
```

### 4. Copy Files
```bash
sudo cp memoriva-backend /opt/memoriva-backend/
sudo cp .env.example /opt/memoriva-backend/.env
sudo cp memoriva-backend.service /etc/systemd/system/
```

### 5. Set Permissions
```bash
sudo chown -R memoriva:memoriva /opt/memoriva-backend
sudo chmod +x /opt/memoriva-backend/memoriva-backend
sudo chmod 600 /opt/memoriva-backend/.env
```

### 6. Configure and Start Service
```bash
sudo systemctl daemon-reload
sudo systemctl enable memoriva-backend
sudo systemctl start memoriva-backend
```

## Service Management

### Check Service Status
```bash
sudo systemctl status memoriva-backend
```

### View Logs
```bash
# Real-time logs
sudo journalctl -u memoriva-backend -f

# Recent logs
sudo journalctl -u memoriva-backend -n 50
```

### Restart Service
```bash
sudo systemctl restart memoriva-backend
```

### Stop Service
```bash
sudo systemctl stop memoriva-backend
```

### Disable Service
```bash
sudo systemctl disable memoriva-backend
```

## Configuration

### Environment Variables

The service reads configuration from `/opt/memoriva-backend/.env`:

- `DATABASE_URL`: PostgreSQL connection string
- `DEEPSEEK_API_KEY`: Your DeepSeek API key
- `OPENAI_API_KEY`: Your OpenAI API key
- `PORT`: Server port (default: 8080)

### Service Configuration

The systemd service is configured with:
- **User**: `memoriva` (dedicated service user)
- **Working Directory**: `/opt/memoriva-backend`
- **Auto-restart**: Yes (on failure)
- **Memory Limit**: 512MB
- **Security**: Restricted file system access

## Troubleshooting

### Service Won't Start
1. Check logs: `sudo journalctl -u memoriva-backend -n 50`
2. Verify .env file: `sudo cat /opt/memoriva-backend/.env`
3. Check file permissions: `ls -la /opt/memoriva-backend/`
4. Test binary manually: `sudo -u memoriva /opt/memoriva-backend/memoriva-backend`

### Database Connection Issues
1. Verify PostgreSQL is running: `sudo systemctl status postgresql`
2. Check database URL in .env file
3. Test database connection manually

### Permission Issues
```bash
sudo chown -R memoriva:memoriva /opt/memoriva-backend
sudo chmod +x /opt/memoriva-backend/memoriva-backend
```

### High Memory Usage
The service is limited to 512MB RAM. If you need more:
1. Edit `/etc/systemd/system/memoriva-backend.service`
2. Change `MemoryMax=512M` to desired value
3. Reload: `sudo systemctl daemon-reload`
4. Restart: `sudo systemctl restart memoriva-backend`

## Health Check

The service provides a health check endpoint:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "memoriva-rag-backend"
}
```

## Updates

To update the application:

1. **Stop the service:**
   ```bash
   sudo systemctl stop memoriva-backend
   ```

2. **Build new version:**
   ```bash
   go build -o memoriva-backend .
   ```

3. **Replace binary:**
   ```bash
   sudo cp memoriva-backend /opt/memoriva-backend/
   sudo chown memoriva:memoriva /opt/memoriva-backend/memoriva-backend
   sudo chmod +x /opt/memoriva-backend/memoriva-backend
   ```

4. **Start service:**
   ```bash
   sudo systemctl start memoriva-backend
   ```

## Security Notes

- The service runs as a non-privileged user (`memoriva`)
- File system access is restricted
- Environment file permissions are set to 600 (owner read/write only)
- The service has memory limits to prevent resource exhaustion

## File Locations

- **Binary**: `/opt/memoriva-backend/memoriva-backend`
- **Configuration**: `/opt/memoriva-backend/.env`
- **Uploads**: `/opt/memoriva-backend/uploads/`
- **Logs**: Available via `journalctl`
- **Service File**: `/etc/systemd/system/memoriva-backend.service`
