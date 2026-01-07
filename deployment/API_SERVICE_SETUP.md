# API Service Setup with systemd

> **⚠️ Note**: The REST API has been moved to a separate repository: [ligneous-gedcom-api](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-api).  
> Please refer to the [deployment guide](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-api/blob/main/deployment/DEPLOYMENT.md) in the new repository for the latest setup instructions.

This guide explains how to set up the ligneous-gedcom API as a systemd service that runs automatically on boot.

## Prerequisites

1. **API binary built**
   ```bash
   cd /apps/ligneous-gedcom-api
   ./scripts/build.sh
   # Or manually:
   go build -o /usr/local/bin/ligneous-api-server ./cmd/api
   ```

2. **Environment variables configured**
   - `DATABASE_URL`: PostgreSQL connection string
   - `STORAGE_DIR`: Directory for file storage (default: `/var/lib/ligneous/storage`)
   - `PORT`: API server port (default: 8090)

## Step 1: Create System User

```bash
# Create system user for the API service
sudo useradd -r -s /bin/false -d /var/lib/ligneous ligneous

# Create directories
sudo mkdir -p /var/lib/ligneous/storage
sudo mkdir -p /var/log/ligneous
sudo chown -R ligneous:ligneous /var/lib/ligneous
sudo chown -R ligneous:ligneous /var/log/ligneous
```

## Step 2: Create systemd Service File

Create `/etc/systemd/system/ligneous-api.service`:

```ini
[Unit]
Description=Ligneous GEDCOM API Server
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=ligneous
Group=ligneous
WorkingDirectory=/var/lib/ligneous

# Environment variables
Environment="DATABASE_URL=postgres://ligneous_user:password@localhost:5432/ligneous_graphs?sslmode=disable"
Environment="STORAGE_DIR=/var/lib/ligneous/storage"
Environment="PORT=8090"

# Executable
ExecStart=/usr/local/bin/ligneous-api-server -port 8090
ExecReload=/bin/kill -HUP $MAINPID

# Restart policy
Restart=always
RestartSec=5

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/ligneous /var/log/ligneous

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ligneous-api

[Install]
WantedBy=multi-user.target
```

## Step 3: Install Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable ligneous-api

# Start the service
sudo systemctl start ligneous-api

# Check status
sudo systemctl status ligneous-api
```

## Step 4: Verify Service

```bash
# Check if service is running
sudo systemctl is-active ligneous-api

# Check logs
sudo journalctl -u ligneous-api -f

# Test API endpoint
curl http://localhost:8090/health
```

## Step 5: Configure Logging

### View logs
```bash
# Real-time logs
sudo journalctl -u ligneous-api -f

# Last 100 lines
sudo journalctl -u ligneous-api -n 100

# Logs since today
sudo journalctl -u ligneous-api --since today

# Logs with timestamps
sudo journalctl -u ligneous-api -o short-precise
```

### Persistent logging
To make logs persistent across reboots, edit `/etc/systemd/journald.conf`:
```ini
[Journal]
Storage=persistent
```

Then restart:
```bash
sudo systemctl restart systemd-journald
```

## Service Management

### Start/Stop/Restart
```bash
sudo systemctl start ligneous-api
sudo systemctl stop ligneous-api
sudo systemctl restart ligneous-api
sudo systemctl reload ligneous-api  # Sends HUP signal
```

### Enable/Disable on boot
```bash
sudo systemctl enable ligneous-api
sudo systemctl disable ligneous-api
```

### Check status
```bash
sudo systemctl status ligneous-api
```

## Environment Variables

### Option 1: In service file (recommended)
Edit `/etc/systemd/system/ligneous-api.service` and add:
```ini
Environment="DATABASE_URL=postgres://..."
Environment="STORAGE_DIR=/var/lib/ligneous/storage"
```

### Option 2: Environment file
Create `/etc/ligneous/api.env`:
```bash
DATABASE_URL=postgres://ligneous_user:password@localhost:5432/ligneous_graphs?sslmode=disable
STORAGE_DIR=/var/lib/ligneous/storage
PORT=8090
```

Then in service file:
```ini
EnvironmentFile=/etc/ligneous/api.env
```

**Note**: Make sure to set proper permissions:
```bash
sudo chmod 600 /etc/ligneous/api.env
sudo chown root:root /etc/ligneous/api.env
```

## Health Checks

### systemd health check
Add to service file:
```ini
[Service]
ExecStartPre=/bin/sh -c 'until pg_isready -h localhost -p 5432; do sleep 1; done'
```

### External health check script
Create `/usr/local/bin/ligneous-healthcheck.sh`:
```bash
#!/bin/bash
if curl -f http://localhost:8090/health > /dev/null 2>&1; then
    exit 0
else
    exit 1
fi
```

Make executable:
```bash
sudo chmod +x /usr/local/bin/ligneous-healthcheck.sh
```

## Monitoring

### Set up monitoring alerts
```bash
# Check if service is running
systemctl is-active ligneous-api || echo "Service is down!"

# Check response time
curl -o /dev/null -s -w "%{time_total}\n" http://localhost:8090/health
```

### Integration with monitoring tools
- **Prometheus**: Expose metrics endpoint
- **Grafana**: Dashboard for metrics
- **Alertmanager**: Alert on service failures

## Troubleshooting

### Service won't start
```bash
# Check service status
sudo systemctl status ligneous-api

# Check logs
sudo journalctl -u ligneous-api -n 50

# Check if port is in use
sudo netstat -tlnp | grep 8090
```

### Permission issues
```bash
# Check directory permissions
ls -la /var/lib/ligneous
ls -la /var/log/ligneous

# Fix permissions
sudo chown -R ligneous:ligneous /var/lib/ligneous
sudo chown -R ligneous:ligneous /var/log/ligneous
```

### Database connection issues
```bash
# Test database connection
psql "$DATABASE_URL" -c "SELECT 1;"

# Check PostgreSQL is running
sudo systemctl status postgresql
```

### Service keeps restarting
```bash
# Check restart count
systemctl show ligneous-api | grep RestartCount

# Check logs for errors
sudo journalctl -u ligneous-api --since "5 minutes ago" | grep -i error
```

## Backup and Recovery

### Backup service configuration
```bash
sudo cp /etc/systemd/system/ligneous-api.service /backup/ligneous-api.service.backup
```

### Restore service
```bash
sudo cp /backup/ligneous-api.service.backup /etc/systemd/system/ligneous-api.service
sudo systemctl daemon-reload
sudo systemctl restart ligneous-api
```

## Security Considerations

1. **Run as non-root user** (already configured)
2. **Limit file system access** (ProtectSystem, ProtectHome)
3. **No new privileges** (NoNewPrivileges)
4. **Resource limits** (LimitNOFILE, LimitNPROC)
5. **Secure environment variables** (use EnvironmentFile with 600 permissions)

## Next Steps

1. Set up nginx reverse proxy (see `NGINX_SETUP.md`)
2. Configure SSL/TLS certificates
3. Set up monitoring and alerts
4. Configure automated backups
5. Set up log rotation




