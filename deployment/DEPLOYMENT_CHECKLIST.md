# Deployment Checklist for ligneous.org

Use this checklist to ensure a complete and secure deployment of the ligneous-gedcom API.

## Pre-Deployment

- [ ] **Code Review**
  - [ ] All tests passing
  - [ ] Code reviewed and approved
  - [ ] No security vulnerabilities

- [ ] **Build**
  - [ ] API binary built and tested
  - [ ] Binary copied to server
  - [ ] Binary has correct permissions

- [ ] **Database**
  - [ ] PostgreSQL installed and running
  - [ ] Database created
  - [ ] User created with appropriate permissions
  - [ ] Connection string tested
  - [ ] Database backups configured

- [ ] **File System**
  - [ ] Storage directory created (`/var/lib/ligneous/storage`)
  - [ ] Log directory created (`/var/log/ligneous`)
  - [ ] Permissions set correctly
  - [ ] Disk space sufficient

## Server Setup

- [ ] **System User**
  - [ ] `ligneous` user created
  - [ ] User has no shell access
  - [ ] User owns required directories

- [ ] **Firewall**
  - [ ] UFW or firewalld configured
  - [ ] SSH access allowed
  - [ ] HTTP (80) and HTTPS (443) allowed
  - [ ] API port (8090) blocked from external access

- [ ] **System Updates**
  - [ ] System packages updated
  - [ ] Security patches applied
  - [ ] Automatic updates configured (optional)

## API Service

- [ ] **systemd Service**
  - [ ] Service file created (`/etc/systemd/system/ligneous-api.service`)
  - [ ] Environment variables configured
  - [ ] Service enabled on boot
  - [ ] Service started and running
  - [ ] Logs accessible via journalctl

- [ ] **Health Check**
  - [ ] Health endpoint responding (`/health`)
  - [ ] Service restarts on failure
  - [ ] Monitoring configured

## Nginx Configuration

- [ ] **SSL/TLS**
  - [ ] Let's Encrypt certificate obtained
  - [ ] Certificate auto-renewal configured
  - [ ] SSL configuration secure (TLS 1.2+)
  - [ ] HSTS header enabled

- [ ] **Reverse Proxy**
  - [ ] Nginx configuration installed
  - [ ] Upstream backend configured correctly
  - [ ] Proxy headers set correctly
  - [ ] File upload size limits appropriate

- [ ] **Security Headers**
  - [ ] HSTS configured
  - [ ] X-Frame-Options set
  - [ ] X-Content-Type-Options set
  - [ ] X-XSS-Protection set
  - [ ] Referrer-Policy set

- [ ] **CORS**
  - [ ] CORS headers configured
  - [ ] Allowed origins restricted (not `*` in production)
  - [ ] Preflight requests handled

- [ ] **Testing**
  - [ ] HTTP redirects to HTTPS
  - [ ] HTTPS endpoints accessible
  - [ ] API endpoints responding
  - [ ] File uploads working

## DNS Configuration

- [ ] **Domain Setup**
  - [ ] `ligneous.org` A record points to server IP
  - [ ] `www.ligneous.org` A record points to server IP
  - [ ] DNS propagation verified
  - [ ] TTL values appropriate

## Monitoring & Logging

- [ ] **Logging**
  - [ ] Nginx access logs configured
  - [ ] Nginx error logs configured
  - [ ] API service logs accessible
  - [ ] Log rotation configured

- [ ] **Monitoring**
  - [ ] Uptime monitoring configured
  - [ ] Health check endpoint monitored
  - [ ] Error rate monitoring
  - [ ] Response time monitoring

- [ ] **Alerts**
  - [ ] Service down alerts
  - [ ] High error rate alerts
  - [ ] SSL certificate expiration alerts
  - [ ] Disk space alerts

## Security

- [ ] **Access Control**
  - [ ] SSH key-based authentication only
  - [ ] Root login disabled
  - [ ] Fail2ban configured (optional)
  - [ ] API rate limiting considered

- [ ] **Backups**
  - [ ] Database backups automated
  - [ ] Configuration files backed up
  - [ ] SSL certificates backed up
  - [ ] Backup restoration tested

- [ ] **Updates**
  - [ ] Automatic security updates configured
  - [ ] Update schedule defined
  - [ ] Rollback plan documented

## Performance

- [ ] **Optimization**
  - [ ] Nginx worker processes configured
  - [ ] Connection limits appropriate
  - [ ] Gzip compression enabled (if applicable)
  - [ ] Caching configured (if applicable)

- [ ] **Load Testing**
  - [ ] Basic load test performed
  - [ ] File upload limits tested
  - [ ] Concurrent request handling verified

## Documentation

- [ ] **Documentation Complete**
  - [ ] Deployment guide documented
  - [ ] Service management documented
  - [ ] Troubleshooting guide available
  - [ ] Runbook for common issues

## Post-Deployment

- [ ] **Verification**
  - [ ] All endpoints tested
  - [ ] File upload tested
  - [ ] Background graph saving verified
  - [ ] Graph persistence verified
  - [ ] Error handling tested

- [ ] **Monitoring**
  - [ ] Service running stable for 24 hours
  - [ ] No unexpected errors in logs
  - [ ] Performance metrics acceptable
  - [ ] SSL certificate auto-renewal working

## Rollback Plan

- [ ] **Rollback Prepared**
  - [ ] Previous version backed up
  - [ ] Rollback procedure documented
  - [ ] Database migration rollback tested (if applicable)

## Sign-Off

- [ ] **Deployment Approved**
  - [ ] Technical lead approval
  - [ ] Security review completed
  - [ ] Performance verified
  - [ ] Documentation complete

---

**Deployment Date**: _______________

**Deployed By**: _______________

**Approved By**: _______________

**Notes**:
_________________________________________________
_________________________________________________
_________________________________________________




