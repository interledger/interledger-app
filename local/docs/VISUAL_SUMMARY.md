# MockGatehub HTTPS Solution - Visual Summary

## The Problem (Visual)

```
🌐 Browser: https://interledger.test
     │
     ├─ Page loads secure (HTTPS) ✅
     │
     └─ Tries to load iframe from: http://mockgatehub:8080
        │
        └─ ❌ MIXED CONTENT! Browser blocks it
           
           Error: "The page was loaded over HTTPS, 
           but requested an insecure frame 'http://mockgatehub:8080'"
```

## The Solution (Visual)

```
🌐 Browser: https://interledger.test
     │
     ├─ Page loads secure (HTTPS) ✅
     │
     └─ Needs iframe from: https://mockgatehub.interledger.test
        │
        ├─ Both use HTTPS! ✅
        │
        ├─ /etc/hosts resolves to 127.0.0.1
        │
        ├─ Port 443 (Traefik) TLS termination
        │
        └─ Forward to mockgatehub:8080 (internal)
           
           ✅ No mixed content error!
           ✅ Iframe loads successfully!
```

## Files Changed

```
📁 /local/
├─ Makefile
│  └─ Added: mockgatehub.interledger.test to hosts target
│
├─ config/san.cnf
│  └─ Added: DNS.9 = mockgatehub.interledger.test
│
└─ mockgatehub.yaml
   ├─ Removed: ports: ["38080:8080"]
   ├─ Added: expose: ["8080"]
   └─ Added: traefik labels (4 lines)
```

## Communication Paths (Before vs After)

### BEFORE ❌

```
Browser → http://mockgatehub:8080 (mixed content error!)
Backend → http://mockgatehub:8080 (works, but iframe breaks)
```

### AFTER ✅

```
Browser → https://mockgatehub.interledger.test → Traefik:443 → mockgatehub:8080 (HTTPS wrapper)
Backend → http://mockgatehub:8080 (unchanged, still works)
```

## Setup Commands (Copy-Paste)

```bash
cd /home/stephan/interledger/interledger-app/local

# 1. Create certs with new hostname
make certs

# 2. Add to /etc/hosts (requires password)
make hosts

# 3. Optional: Trust cert on macOS
make trust

# 4. Restart services
make down
make infrastructure

# In another terminal:
make application

# 5. Test
curl -k https://mockgatehub.interledger.test/health
```

## URL Mapping Reference

| Request Source | URL | Protocol | Port | Via |
|---|---|---|---|---|
| Browser (public) | `https://mockgatehub.interledger.test` | HTTPS | 443 | Traefik |
| Backend (internal) | `http://mockgatehub:8080` | HTTP | 8080 | Docker DNS |

## Verification One-Liners

```bash
# Is certificate valid?
openssl x509 -in config/certs/local.crt -text -noout | grep mockgatehub

# Is /etc/hosts updated?
grep mockgatehub /etc/hosts

# Is Traefik running?
docker ps | grep traefik

# Is mockgatehub running?
docker ps | grep mockgatehub

# Can browser reach it?
curl -k https://mockgatehub.interledger.test/health

# Can backend reach it?
docker compose exec backend curl http://mockgatehub:8080/health

# Traefik routing healthy?
curl -k -H "Host: mockgatehub.interledger.test" https://localhost/health
```

## Decision Tree: Troubleshooting

```
Still seeing mixed content error?
│
├─ Is the iframe URL HTTPS? (check browser dev tools)
│  ├─ No → Backend not returning HTTPS URL
│  │       Check: GATEHUB_WIDGET_IFRAME_BASE_URL env var
│  │
│  └─ Yes → Continue below
│
├─ Can you reach mockgatehub.interledger.test?
│  ├─ No → Check /etc/hosts
│  │       grep mockgatehub /etc/hosts
│  │
│  └─ Yes → Continue below
│
├─ Does Traefik have mockgatehub route?
│  ├─ No → Restart mockgatehub: docker compose restart mockgatehub
│  │
│  └─ Yes → Check browser cache
│           Clear: Cmd+Shift+Delete
│           Hard reload: Cmd+Shift+R
│
└─ Still broken? → docker compose logs traefik | grep mockgatehub
```

## Key Insights

1. **Same Service, Two URLs**
   - Internal: `http://mockgatehub:8080` (fast, no TLS)
   - External: `https://mockgatehub.interledger.test` (secure, TLS)

2. **Traefik Magic**
   - Listens on 443 (HTTPS)
   - Routes by hostname
   - Terminates TLS
   - Forwards as HTTP internally

3. **No Backend Code Changes**
   - Backend still uses `http://mockgatehub:8080`
   - Backend doesn't need to know about HTTPS
   - Works via Docker internal DNS

4. **Certificate Covers It**
   - Wildcard: `*.interledger.test`
   - Includes: `mockgatehub.interledger.test`

## Architecture Benefits

✅ **Security**: Browser gets HTTPS  
✅ **Performance**: Backend uses HTTP (no TLS overhead)  
✅ **Simplicity**: No code changes needed  
✅ **Scalability**: Traefik handles load balancing  
✅ **Maintainability**: Centralized certificate management  

## Before & After Comparison

### BEFORE
- ❌ Port 38080 exposed directly
- ❌ HTTP only
- ❌ Mixed content errors
- ❌ Browser can't load iframe
- ✅ Backend works

### AFTER
- ✅ Port 443 via Traefik
- ✅ HTTPS everywhere (external)
- ✅ HTTP internal (fast)
- ✅ No mixed content errors
- ✅ Browser loads iframe
- ✅ Backend still works

## Files to Review After Setup

```
✅ config/certs/local.crt
   └─ Should include "mockgatehub.interledger.test"

✅ /etc/hosts
   └─ Should include "127.0.0.1 mockgatehub.interledger.test"

✅ docker-compose.yaml (includes mockgatehub.yaml)
   └─ Should show mockgatehub with expose (not ports)

✅ Traefik dashboard: http://traefik.test:28080
   └─ Should show mockgatehub router (green = healthy)
```

## Timeline to Fix

- **5 min**: Regenerate certs (`make certs`)
- **2 min**: Update /etc/hosts (`make hosts`)
- **3 min**: Restart services (`make down && make infrastructure && make application`)
- **2 min**: Test (`curl -k https://mockgatehub.interledger.test/health`)
- **Total**: ~12 minutes for full fix

## Remember

- 🔒 Self-signed certificates → browser warnings are normal
- 🖥️ macOS only: Run `make trust` to suppress warnings
- 🔄 Always restart after cert changes
- 📋 Backend keeps using `http://mockgatehub:8080`
- 🌐 Browser always uses `https://mockgatehub.interledger.test`

---

**Documentation**: See detailed guides in `/local/docs/`
- `IMPLEMENTATION_COMPLETE.md` ← Start here
- `mockgatehub-https-setup.md` ← Setup instructions
- `MOCKGATEHUB_HTTPS_INTEGRATION.md` ← Full details
- `MOCKGATEHUB_QUICKREF.md` ← Quick reference
