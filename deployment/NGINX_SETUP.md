# Nginx Setup for ligneous.org

This guide explains how to set up nginx as a reverse proxy for the ligneous-gedcom API on ligneous.org.

## Prerequisites

1. **Server with nginx installed**
   ```bash
   sudo apt update
   sudo apt install nginx
   ```

2. **Domain DNS configured**
   - Point `ligneous.org` and `www.ligneous.org` to your server's IP address
   - Verify with: `dig ligneous.org` or `nslookup ligneous.org`

3. **API server running**
   - The API should be running on `localhost:8090`
   - Use systemd service (see `API_SERVICE_SETUP.md`) or run manually

4. **Firewall configured**
   ```bash
   sudo ufw allow 'Nginx Full'
   sudo ufw allow OpenSSH
   sudo ufw enable
   ```

## Step 1: Install SSL Certificate (Let's Encrypt)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate (nginx plugin will automatically configure)
sudo certbot --nginx -d ligneous.org -d www.ligneous.org

# Test auto-renewal
sudo certbot renew --dry-run
```

**Note**: Certbot will automatically modify your nginx config. We'll replace it with our optimized config in the next step.

## Step 2: Install Nginx Configuration

```bash
# Copy the configuration file
sudo cp deployment/nginx-ligneous.org.conf /etc/nginx/sites-available/ligneous.org

# Create symlink to enable the site
sudo ln -s /etc/nginx/sites-available/ligneous.org /etc/nginx/sites-enabled/

# Remove default site (optional)
sudo rm /etc/nginx/sites-enabled/default

# Test nginx configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

## Step 3: Verify Setup

1. **Check nginx status:**
   ```bash
   sudo systemctl status nginx
   ```

2. **Test HTTP redirect:**
   ```bash
   curl -I http://ligneous.org
   # Should return 301 redirect to HTTPS
   ```

3. **Test HTTPS API:**
   ```bash
   curl https://ligneous.org/api/v1/health
   # Should return JSON health check response
   ```

4. **Test file upload:**
   ```bash
   curl -X POST https://ligneous.org/api/v1/files \
     -F "file=@testdata/xavier.ged" \
     -F "name=xavier.ged"
   ```

## Step 4: Configure Log Rotation (Optional)

Nginx logs are automatically rotated, but you can verify:

```bash
# Check log rotation config
cat /etc/logrotate.d/nginx

# View logs
sudo tail -f /var/log/nginx/ligneous.org.access.log
sudo tail -f /var/log/nginx/ligneous.org.error.log
```

## Configuration Details

### Upstream Backend
- **Server**: `127.0.0.1:8090` (local API server)
- **Keepalive**: 32 connections (for connection pooling)

### SSL/TLS
- **Protocols**: TLSv1.2, TLSv1.3
- **Ciphers**: Modern, secure cipher suites
- **HSTS**: Enabled with 1-year max-age

### Security Headers
- `Strict-Transport-Security`: Forces HTTPS
- `X-Frame-Options`: Prevents clickjacking
- `X-Content-Type-Options`: Prevents MIME sniffing
- `X-XSS-Protection`: XSS protection
- `Referrer-Policy`: Controls referrer information

### File Upload Limits
- **Max body size**: 50MB (configurable)
- **Buffer size**: 128KB

### CORS
- Currently allows all origins (`*`)
- For production, restrict to specific domains:
  ```nginx
  add_header Access-Control-Allow-Origin "https://yourdomain.com" always;
  ```

## Troubleshooting

### Nginx won't start
```bash
# Check configuration syntax
sudo nginx -t

# Check error logs
sudo tail -f /var/log/nginx/error.log
```

### API not responding
```bash
# Check if API server is running
curl http://localhost:8090/health

# Check nginx error logs
sudo tail -f /var/log/nginx/ligneous.org.error.log

# Check API server logs
journalctl -u ligneous-api -f
```

### SSL certificate issues
```bash
# Check certificate status
sudo certbot certificates

# Renew certificate manually
sudo certbot renew

# Check certificate expiration
openssl x509 -in /etc/letsencrypt/live/ligneous.org/cert.pem -noout -dates
```

### 502 Bad Gateway
- API server is not running or not accessible
- Check firewall rules
- Verify upstream server in nginx config

### 413 Request Entity Too Large
- Increase `client_max_body_size` in nginx config
- Restart nginx after changes

## Performance Tuning

### Increase worker connections
Edit `/etc/nginx/nginx.conf`:
```nginx
worker_processes auto;
worker_connections 1024;
```

### Enable gzip compression
Add to server block:
```nginx
gzip on;
gzip_vary on;
gzip_proxied any;
gzip_comp_level 6;
gzip_types text/plain text/css application/json application/javascript text/xml application/xml application/xml+rss text/javascript;
```

### Enable caching (for static responses)
```nginx
location /api/v1/files {
    proxy_pass http://ligneous_api;
    proxy_cache my_cache;
    proxy_cache_valid 200 1h;
    proxy_cache_use_stale error timeout updating http_500 http_502 http_503 http_504;
}
```

## Monitoring

### Set up log monitoring
```bash
# Install goaccess for real-time log analysis
sudo apt install goaccess

# View real-time stats
sudo goaccess /var/log/nginx/ligneous.org.access.log --log-format=COMBINED
```

### Set up uptime monitoring
- Use services like UptimeRobot, Pingdom, or StatusCake
- Monitor: `https://ligneous.org/api/v1/health`

## Backup Configuration

```bash
# Backup nginx config
sudo cp /etc/nginx/sites-available/ligneous.org /etc/nginx/sites-available/ligneous.org.backup

# Backup SSL certificates
sudo tar -czf /backup/letsencrypt-$(date +%Y%m%d).tar.gz /etc/letsencrypt/
```

## Next Steps

1. Set up API service with systemd (see `API_SERVICE_SETUP.md`)
2. Configure monitoring and alerts
3. Set up automated backups
4. Configure rate limiting (if needed)
5. Set up CDN (if needed for static assets)

