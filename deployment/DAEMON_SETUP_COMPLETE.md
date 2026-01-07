# API Daemon Setup - Complete Documentation

## Overview

We successfully converted the ligneous-gedcom API server into a **systemd daemon** (system service) that runs automatically in the background, starts on boot, and automatically restarts on failure.

## What is a Daemon?

A **daemon** (or service) is a background process that runs continuously, independent of user sessions. In Linux, systemd manages these services, providing:

- **Automatic startup** on system boot
- **Automatic restart** on failure
- **Process management** and monitoring
- **Logging** via journald
- **Resource limits** and security controls
- **Dependency management** (e.g., wait for PostgreSQL before starting)

## What We Accomplished

### 1. Built and Installed the API Binary

**Location**: `/usr/local/bin/ligneous-api-server`

> **⚠️ Note**: The API has been moved to a separate repository. Build from the new location:
> ```bash
> cd /apps/ligneous-gedcom-api
> ./scripts/build.sh
> ```

```bash
# Built from source (old location - now use ligneous-gedcom-api repository)
cd /apps/ligneous-gedcom-api
go build -o /usr/local/bin/ligneous-api-server ./cmd/api

# Installed to system PATH
sudo cp /tmp/ligneous-api-server /usr/local/bin/ligneous-api-server
sudo chmod +x /usr/local/bin/ligneous-api-server
```

**Binary Size**: 29MB  
**Architecture**: Linux amd64

### 2. Created System User

**User**: `ligneous`  
**Type**: System user (no shell, no login)

```bash
sudo useradd -r -s /bin/false -d /var/lib/ligneous ligneous
```

**Purpose**: Run the API with minimal privileges (security best practice)

### 3. Created Required Directories

**Storage Directory**: `/var/lib/ligneous/storage`
- Stores uploaded GEDCOM files
- Stores BadgerDB graph data
- Owned by `ligneous` user

**Log Directory**: `/var/log/ligneous`
- Available for application logs (if needed)
- Currently using systemd journal instead

**Permissions**:
```bash
sudo chown -R ligneous:ligneous /var/lib/ligneous
sudo chown -R ligneous:ligneous /var/log/ligneous
```

### 4. Created systemd Service File

**Location**: `/etc/systemd/system/ligneous-api.service`

**Key Features**:

#### Service Configuration
- **Type**: `simple` - runs as a single process
- **User**: `ligneous` - runs with minimal privileges
- **Working Directory**: `/var/lib/ligneous`

#### Environment Variables
```ini
Environment="DATABASE_URL=postgres://ligneous_test:test_password@localhost:5432/ligneous_graphs_test?sslmode=disable"
Environment="STORAGE_DIR=/var/lib/ligneous/storage"
Environment="PORT=8090"
```

#### Auto-Restart Policy
- **Restart**: `always` - restarts on any failure
- **RestartSec**: `5` - waits 5 seconds before restarting

#### Resource Limits
- **LimitNOFILE**: 65536 - maximum open file descriptors
- **LimitNPROC**: 4096 - maximum processes

#### Security Hardening
- **NoNewPrivileges**: `true` - cannot gain additional privileges
- **PrivateTmp**: `true` - private /tmp directory
- **ProtectSystem**: `strict` - read-only filesystem (except allowed paths)
- **ProtectHome**: `true` - no access to /home directories
- **ReadWritePaths**: Only `/var/lib/ligneous` and `/var/log/ligneous`

#### Logging
- **StandardOutput**: `journal` - logs to systemd journal
- **StandardError**: `journal` - errors to systemd journal
- **SyslogIdentifier**: `ligneous-api` - tag for log filtering

#### Dependencies
- **After**: `network.target postgresql.service` - starts after network and PostgreSQL
- **Requires**: `postgresql.service` - requires PostgreSQL to be running

### 5. Installed and Activated the Service

```bash
# Install service file
sudo cp /tmp/ligneous-api.service /etc/systemd/system/ligneous-api.service

# Reload systemd to recognize new service
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable ligneous-api

# Start the service
sudo systemctl start ligneous-api
```

## Service Status

### Current Status
- **State**: `active (running)`
- **Enabled**: `yes` (starts on boot)
- **PID**: Running as process managed by systemd
- **Port**: 8090
- **Health**: Responding to `/health` endpoint

### Verification
```bash
# Check service status
sudo systemctl status ligneous-api

# Check if active
sudo systemctl is-active ligneous-api

# Check if enabled
sudo systemctl is-enabled ligneous-api

# Test API health
curl http://localhost:8090/health
```

## Service Management Commands

### Start/Stop/Restart
```bash
# Start the service
sudo systemctl start ligneous-api

# Stop the service
sudo systemctl stop ligneous-api

# Restart the service
sudo systemctl restart ligneous-api

# Reload configuration (sends HUP signal)
sudo systemctl reload ligneous-api
```

### Enable/Disable on Boot
```bash
# Enable to start on boot
sudo systemctl enable ligneous-api

# Disable from starting on boot
sudo systemctl disable ligneous-api
```

### View Logs
```bash
# Real-time logs (follow)
sudo journalctl -u ligneous-api -f

# Last 50 lines
sudo journalctl -u ligneous-api -n 50

# Logs since today
sudo journalctl -u ligneous-api --since today

# Logs with timestamps
sudo journalctl -u ligneous-api -o short-precise

# Logs since specific time
sudo journalctl -u ligneous-api --since "2026-01-03 18:00:00"
```

## Service File Contents

The complete service file is located at `/etc/systemd/system/ligneous-api.service`:

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
Environment="DATABASE_URL=postgres://ligneous_test:test_password@localhost:5432/ligneous_graphs_test?sslmode=disable"
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

## How It Works

### Boot Sequence
1. System boots
2. systemd starts
3. Network becomes available
4. PostgreSQL starts
5. `ligneous-api` service starts (because it's enabled)
6. API server binds to port 8090
7. API is ready to accept requests

### Failure Recovery
1. API process crashes or exits
2. systemd detects failure
3. Waits 5 seconds (`RestartSec=5`)
4. Automatically restarts the service
5. Process continues running

### Logging Flow
1. API writes to stdout/stderr
2. systemd captures output
3. Logs stored in systemd journal
4. Accessible via `journalctl -u ligneous-api`

## Configuration Updates

### Changing Environment Variables

1. **Edit the service file**:
   ```bash
   sudo nano /etc/systemd/system/ligneous-api.service
   ```

2. **Update environment variables** (e.g., DATABASE_URL):
   ```ini
   Environment="DATABASE_URL=postgres://new_user:new_pass@host:5432/dbname?sslmode=disable"
   ```

3. **Reload and restart**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart ligneous-api
   ```

### Using Environment File (Alternative)

Instead of inline environment variables, you can use a file:

1. **Create environment file**:
   ```bash
   sudo nano /etc/ligneous/api.env
   ```
   ```bash
   DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable
   STORAGE_DIR=/var/lib/ligneous/storage
   PORT=8090
   ```

2. **Set permissions**:
   ```bash
   sudo chmod 600 /etc/ligneous/api.env
   sudo chown root:root /etc/ligneous/api.env
   ```

3. **Update service file**:
   ```ini
   EnvironmentFile=/etc/ligneous/api.env
   ```

4. **Reload and restart**:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl restart ligneous-api
   ```

## Troubleshooting

### Service Won't Start

**Check status**:
```bash
sudo systemctl status ligneous-api
```

**Check logs**:
```bash
sudo journalctl -u ligneous-api -n 50
```

**Common issues**:
- Port 8090 already in use → Stop other processes: `sudo pkill -f api-server`
- PostgreSQL not running → Start PostgreSQL: `sudo systemctl start postgresql`
- Permission issues → Check directory ownership: `ls -la /var/lib/ligneous`
- Database connection failed → Verify DATABASE_URL in service file

### Service Keeps Restarting

**Check restart count**:
```bash
systemctl show ligneous-api | grep RestartCount
```

**Check recent logs for errors**:
```bash
sudo journalctl -u ligneous-api --since "5 minutes ago" | grep -i error
```

**Common causes**:
- Database connection issues
- Invalid DATABASE_URL
- Port conflicts
- Permission problems

### Viewing Real-Time Logs

```bash
# Follow logs in real-time
sudo journalctl -u ligneous-api -f

# Filter for errors only
sudo journalctl -u ligneous-api -f | grep -i error
```

### Testing Service

```bash
# Check if service is running
sudo systemctl is-active ligneous-api

# Test health endpoint
curl http://localhost:8090/health

# Test file upload
curl -X POST http://localhost:8090/api/v1/files \
  -F "file=@testdata/xavier.ged" \
  -F "name=xavier.ged"
```

## Integration with Nginx

The API daemon runs on `localhost:8090` and is accessed through nginx reverse proxy:

```
Internet → nginx (port 443) → API daemon (port 8090)
```

**Nginx Configuration** (`/etc/nginx/sites-available/ligneous.org`):
```nginx
upstream ligneous_api {
    server 127.0.0.1:8090;
    keepalive 32;
}
```

**Benefits**:
- API not directly exposed to internet
- SSL/TLS termination at nginx
- Load balancing capability
- Request logging at nginx level

## Security Considerations

### What We Implemented

1. **Non-root execution**: Runs as `ligneous` user (no root privileges)
2. **Filesystem protection**: Read-only system, only specific writable paths
3. **No privilege escalation**: `NoNewPrivileges=true`
4. **Resource limits**: Prevents resource exhaustion
5. **Private temp**: Isolated temporary directory

### Additional Recommendations

1. **Firewall**: Ensure port 8090 is not exposed externally (only nginx should access it)
2. **Database credentials**: Use strong passwords, consider secrets management
3. **Log rotation**: Configure journald log limits
4. **Monitoring**: Set up alerts for service failures
5. **Backups**: Regular backups of `/var/lib/ligneous/storage`

## Monitoring

### Health Checks

**Endpoint**: `http://localhost:8090/health`

**Response**:
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2026-01-03T18:25:51Z"
  }
}
```

### Service Metrics

**Check service status**:
```bash
systemctl show ligneous-api --property=ActiveState,SubState,MainPID
```

**Check resource usage**:
```bash
systemctl status ligneous-api
# Shows CPU and memory usage
```

**Check restart count**:
```bash
systemctl show ligneous-api --property=RestartCount
```

## Backup and Recovery

### Service Configuration

**Backup**:
```bash
sudo cp /etc/systemd/system/ligneous-api.service /backup/ligneous-api.service.backup
```

**Restore**:
```bash
sudo cp /backup/ligneous-api.service.backup /etc/systemd/system/ligneous-api.service
sudo systemctl daemon-reload
sudo systemctl restart ligneous-api
```

### Data Backup

**Storage directory**:
```bash
sudo tar -czf /backup/ligneous-storage-$(date +%Y%m%d).tar.gz /var/lib/ligneous/storage
```

## Summary

### What We Created

✅ **Systemd daemon** (`ligneous-api.service`)  
✅ **System user** (`ligneous`)  
✅ **Installed binary** (`/usr/local/bin/ligneous-api-server`)  
✅ **Storage directories** (`/var/lib/ligneous/storage`)  
✅ **Auto-start on boot** (enabled)  
✅ **Auto-restart on failure** (configured)  
✅ **Secure execution** (non-root, hardened)  
✅ **Centralized logging** (systemd journal)  

### Benefits

1. **Reliability**: Automatic restart on failure
2. **Availability**: Starts automatically on boot
3. **Security**: Runs with minimal privileges
4. **Monitoring**: Integrated with systemd logging
5. **Management**: Standard Linux service management
6. **Integration**: Works seamlessly with nginx

### Next Steps

1. ✅ API daemon running
2. ✅ Nginx configured (from previous steps)
3. ⏭️ Test through nginx: `curl https://ligneous.org/api/v1/health`
4. ⏭️ Set up monitoring and alerts
5. ⏭️ Configure production database (if using test database)
6. ⏭️ Set up automated backups

## Files Created/Modified

| File | Purpose | Location |
|------|---------|----------|
| Service file | systemd service definition | `/etc/systemd/system/ligneous-api.service` |
| Binary | API server executable | `/usr/local/bin/ligneous-api-server` |
| Storage | File storage directory | `/var/lib/ligneous/storage` |
| Logs | Application logs | systemd journal (`journalctl -u ligneous-api`) |

## References

- **systemd Documentation**: https://www.freedesktop.org/software/systemd/man/systemd.service.html
- **systemd Security**: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
- **Journalctl Guide**: https://www.freedesktop.org/software/systemd/man/journalctl.html

---

**Setup Date**: January 3, 2026  
**Service Name**: `ligneous-api.service`  
**Status**: ✅ Active and Running

