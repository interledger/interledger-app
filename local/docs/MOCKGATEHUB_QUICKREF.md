# Quick Reference: MockGatehub HTTPS Changes

## TL;DR - What Changed

| File | Change | Why |
|------|--------|-----|
| `Makefile` | Added `mockgatehub.interledger.test` to hosts target | Local DNS resolution for browser |
| `config/san.cnf` | Added `DNS.9 = mockgatehub.interledger.test` | Certificate includes mockgatehub domain |
| `mockgatehub.yaml` | Removed port mapping, added Traefik labels, removed external port | Traefik HTTPS routing instead of direct port exposure |

## To Apply Changes

```bash
cd local

# 1. Regenerate certs with new SAN
make certs

# 2. Update /etc/hosts 
make hosts

# 3. Restart everything
make down
make infrastructure
# in another terminal:
make application
```

## Result

### Before
- ❌ `http://mockgatehub:8080` → CORS issues, mixed content error in browser
- Frontend loads insecure iframe → blocked by browser

### After  
- ✅ `https://mockgatehub.interledger.test` → HTTPS, browser happy
- Backend still uses `http://mockgatehub:8080` → fast internal connection
- Frontend loads secure iframe → no errors!

## Verification

```bash
# Test HTTPS endpoint
curl -k https://mockgatehub.interledger.test/health

# Test backend can still access via HTTP
docker compose exec backend curl http://mockgatehub:8080/health

# Check Traefik routing
curl -k -H "Host: mockgatehub.interledger.test" https://localhost/health
```

## How It Works In One Picture

```
BROWSER REQUEST                    CONTAINER NETWORK
─────────────────                  ─────────────────

https://mockgatehub.interledger.test
           ↓
    127.0.0.1:443 (localhost)
           ↓
    Traefik: TLS termination
           ↓
    http://mockgatehub:8080 (Docker DNS)
           ↓
    MockGatehub response
           ↓
    Traefik: wraps in HTTPS
           ↓
    Browser gets HTTPS response ✓
```

Meanwhile, backend (in same Docker network):
```
Backend needs Gatehub API
           ↓
Uses: http://mockgatehub:8080
           ↓
Docker DNS resolves to internal IP
           ↓
Fast HTTP connection (same network)
           ↓
No certificate validation needed ✓
```

## Config Comparison

### Before
```yaml
mockgatehub:
  ports:
    - "38080:8080"  # Exposed on machine port
```
→ Accessible at `http://localhost:38080` but mixed content error!

### After
```yaml
mockgatehub:
  expose:
    - "8080"  # Internal only
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.mockgatehub.rule=Host(`mockgatehub.interledger.test`)"
    - "traefik.http.routers.mockgatehub.entrypoints=websecure"
    - "traefik.http.services.mockgatehub.loadbalancer.server.port=8080"
```
→ Accessible at `https://mockgatehub.interledger.test` (HTTPS) via Traefik!

## Important: Two Different URLs

**For Backend Services** (Go code):
```env
GATEHUB_API_BASE_URL=http://mockgatehub:8080
```
- No HTTPS needed
- Uses Docker internal DNS
- Fast

**For Browser Clients** (Frontend/iframe):
```
https://mockgatehub.interledger.test/?bearer=token
```
- Must be HTTPS
- Uses /etc/hosts DNS mapping
- Secure

## Troubleshooting One-Liners

```bash
# Still seeing mixed content error?
docker compose logs traefik | grep mockgatehub

# Certificate warnings?
make trust  # macOS only

# MockGatehub not responding?
docker ps | grep mockgatehub

# Traefik routing issues?
curl -v -k -H "Host: mockgatehub.interledger.test" https://localhost/health
```

---

**Full Details**: See [mockgatehub-https-setup.md](./mockgatehub-https-setup.md) and [MOCKGATEHUB_HTTPS_INTEGRATION.md](./MOCKGATEHUB_HTTPS_INTEGRATION.md)
