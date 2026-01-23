# MockGatehub HTTPS Setup Guide

## Problem

Mixed content error when loading Gatehub iframe:
```
Mixed Content: The page at 'https://interledger.test/personal-details' was loaded over HTTPS, 
but requested an insecure frame 'http://mockgatehub:8080/?bearer=...'. This request has been blocked; 
the content must be served over HTTPS.
```

Browsers block insecure iframes embedded in secure pages. MockGatehub was served over HTTP on port 38080.

## Solution Overview

Configure MockGatehub to be accessible via HTTPS at `https://mockgatehub.interledger.test` using:
1. Traefik reverse proxy for HTTPS termination
2. Self-signed certificate (already exists and supports wildcard `*.interledger.test`)
3. Hosts file entry for local DNS resolution
4. Docker internal networking for backend service communication

## Changes Made

### 1. Updated `/etc/hosts` (via Makefile)

Added:
```
127.0.0.1 mockgatehub.interledger.test
```

**File**: `Makefile` → `hosts` target

Run after making changes:
```bash
make hosts
```

### 2. Updated Certificate SAN Configuration

Added explicit entry for mockgatehub to certificate Subject Alternative Names:
```
DNS.9 = mockgatehub.interledger.test
```

**File**: `config/san.cnf`

Regenerate certificates:
```bash
make certs
```

> **Note**: The wildcard `*.interledger.test` already covers this, but explicit entry ensures compatibility.

### 3. Configured Traefik Routing for MockGatehub

Changed mockgatehub service configuration:
- Removed port mapping (`ports: 38080:8080`)
- Added `expose: ["8080"]` for Docker internal networking
- Added Traefik labels:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.mockgatehub.rule=Host(`mockgatehub.interledger.test`)"
  - "traefik.http.routers.mockgatehub.entrypoints=websecure"
  - "traefik.http.services.mockgatehub.loadbalancer.server.port=8080"
```

**File**: `mockgatehub.yaml`

This tells Traefik to:
- Route all requests to `mockgatehub.interledger.test` to the mockgatehub container
- Use the HTTPS entrypoint (websecure)
- Forward to port 8080 internally

## How It Works

### Browser Flow (Frontend)
1. Frontend runs at `https://interledger.test` (HTTPS via Traefik)
2. Backend returns iframe URL: `https://mockgatehub.interledger.test/?bearer=token`
3. Browser requests HTTPS iframe → Traefik intercepts request
4. Traefik terminates TLS using `config/certs/local.crt`
5. Traefik forwards HTTP request to mockgatehub container on internal network
6. Response returns to browser with valid HTTPS
7. **No mixed content error!**

### Backend Flow (Service Communication)
1. Backend needs to call Gatehub API endpoints
2. Uses `GATEHUB_API_BASE_URL=http://mockgatehub:8080` (already configured in wallet.yaml)
3. Docker DNS resolves `mockgatehub` to the container's internal IP
4. Connects via HTTP on port 8080 (no need for HTTPS on internal network)
5. Works seamlessly - no certificate validation needed

## Step-by-Step Setup

### First Time Setup

```bash
cd local

# 1. Regenerate certificates with new SAN entries
make certs

# 2. Update /etc/hosts (requires sudo)
make hosts

# 3. Trust the certificate on macOS (if needed)
make trust

# 4. Bring down any running services
make down

# 5. Start infrastructure first (includes Traefik)
make infrastructure

# 6. In another terminal, start application (includes mockgatehub)
make application
```

### Verification

```bash
# Check that mockgatehub is running
docker ps | grep mockgatehub

# Test HTTPS endpoint with curl (may show cert warning, that's normal for self-signed)
curl -k https://mockgatehub.interledger.test/health

# Test in browser - navigate to https://interledger.test and trigger KYC activation
# Iframe should load without mixed content errors
```

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Browser / Frontend                       │
│                   (https://interledger.test)                     │
└────────────────────────────────┬────────────────────────────────┘
                                 │
                                 │ HTTPS Request to
                                 │ https://mockgatehub.interledger.test
                                 │
                    ┌────────────▼────────────┐
                    │    Traefik (Docker)     │
                    │  Port 443 (HTTPS)       │
                    │                         │
                    │ • Terminates TLS        │
                    │ • Routes by hostname    │
                    │ • Uses config/certs/    │
                    │   local.crt/.key        │
                    └────────────┬────────────┘
                                 │
                                 │ HTTP to internal port 8080
                                 │ (Docker network DNS)
                    ┌────────────▼────────────┐
                    │   MockGatehub           │
                    │   :8080                 │
                    │                         │
                    │ • Internal port only    │
                    │ • No HTTPS needed       │
                    │ • Fast inter-container  │
                    │   communication         │
                    └─────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│ Backend Service (separate from browser)                         │
│                                                                 │
│ Calls: http://mockgatehub:8080/id/v1/users (Docker DNS)        │
└─────────────────────────────────────────────────────────────────┘
```

## Environment Variables Summary

### Backend Service (wallet.yaml)
```yaml
GATEHUB_API_BASE_URL: http://mockgatehub:8080
# ↑ Internal Docker networking - no HTTPS needed
# This is what backend uses to make API calls
```

### Frontend Users (Gatehub Widget)
```
https://mockgatehub.interledger.test/?bearer=iframe-token-xxx
# ↑ HTTPS through Traefik - browser loads this URL
# This is what browser iframe loads
```

## Troubleshooting

### Certificate Not Trusted (macOS)

Error: `ERR_CERT_AUTHORITY_INVALID`

Solution:
```bash
# Trust the self-signed certificate
make trust

# Restart browser
# Clear browser cache (Cmd+Shift+Delete)
```

### Still Seeing Mixed Content Error

Check:
1. Frontend is actually HTTPS: `https://interledger.test` (not `http://`)
2. Traefik is running: `docker ps | grep traefik`
3. Certificates were regenerated: `ls -la config/certs/`
4. Hosts file was updated: `grep mockgatehub /etc/hosts`
5. Docker containers can see each other: `docker compose exec mockgatehub ping localhost`

### MockGatehub Responding but Traefik Not Routing

Check Traefik dashboard:
```
http://traefik.test:28080
# Verify mockgatehub router is showing and healthy
```

View Traefik logs:
```bash
docker compose logs traefik | grep mockgatehub
```

### Connection Refused

Issue: Can't connect to `https://mockgatehub.interledger.test`

Check:
```bash
# Verify mockgatehub container is running
docker ps | grep mockgatehub

# Check container can respond internally
docker compose exec mockgatehub curl http://localhost:8080/health

# Check Traefik can reach container
docker compose logs traefik | tail -20
```

## Important Notes

1. **Internal vs. External URLs**: 
   - Backend uses `http://mockgatehub:8080` (internal, faster)
   - Browser uses `https://mockgatehub.interledger.test` (external, secure)

2. **Certificate Regeneration**: After updating `config/san.cnf`, always run `make certs` and restart services

3. **DNS Resolution**: `/etc/hosts` is only for local machine. Docker containers use service names via Docker DNS. No need to update container /etc/hosts.

4. **Traefik Discovery**: Services register with Traefik via Docker labels. Changes to labels require service restart.

5. **Self-Signed Certificates**: Browser warnings are normal and expected in development. Use `make trust` on macOS or browser settings to suppress warnings.

## Additional Resources

- [Traefik Documentation](https://doc.traefik.io/)
- [Self-Signed Certificates Guide](../README.md#certs)
- [MockGatehub API Documentation](../../mockgatehub/README.md)
- [Gatehub Account Activation Flow](./gatehub-account-activation.md)
