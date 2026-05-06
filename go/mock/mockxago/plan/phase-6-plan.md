# Phase 6 Implementation Plan: Local Environment Setup for Xago

## Overview

Phase 6 is a critical infrastructure phase that sets up the local development environment to support Xago integration. Without this phase, E2E tests cannot run because MockXago won't be accessible and the backend won't have the configuration needed to communicate with it.

**Phase Outcome**: Full local Docker Compose stack (Interledger App + MockXago) is operational and both services can communicate.

## Current Status

- ✓ Phase 5 complete: MockXago with 65%+ coverage and full feature tests
- ✗ Local environment setup incomplete for Xago
  - Missing: `local/mockxago.yaml` service definition
  - Missing: XAGO_* environment variables in `local/wallet.yaml`
  - Missing: MockXago background configuration in E2E features

## Scope

### What's Included
- Creating `local/mockxago.yaml` service definition
- Adding XAGO_* environment variables to backend configuration
- Configuring Docker Compose networking and service discovery
- Setting up Redis backend for MockXago
- Configuring webhook delivery from MockXago to backend
- Verifying service inter-communication

### What's NOT Included
- MockXago code development (Phase 5)
- E2E test implementations (Phases 7+)
- UI/Frontend changes

## Implementation Tasks

### Task 1: Create MockXago Service Definition (1 hour)
**Objective**: Create `local/mockxago.yaml` for Docker Compose

**Action**:
1. Create file: `local/mockxago.yaml` with service configuration
2. Use `bosbaber/mockxago:local/mockxago.yaml` as reference (it's a complete template)
3. Key elements to include:
   - **Build context**: `context: ../go/mock/mockxago` (relative to local/)
   - **Profiles**: Include in `application` profile
   - **Ports**: Expose 8080 internally
   - **Traefik labels**: Configure routing to `mockxago.interledger.test` over TLS
   - **Environment**:
     ```
     LOG_LEVEL: debug
     MOCKXAGO_REDIS_URL: redis:6379
     MOCKXAGO_REDIS_DB: '4'  # Separate DB from other services
     WEBHOOK_URL: http://backend:8080/webhooks/xago
     WEBHOOK_SECRET: (generate secure value)
     XAGO_API_PUBLIC_KEY: test-public-key
     XAGO_API_SECRET: test-secret
     XAGO_MOCK_TEST_MODE: "true"
     ```
   - **Healthcheck**: HTTP GET to `/health`
   - **Extra hosts**: May need `host.docker.internal` mapping

**Output file structure**:
```yaml
services:
  mockxago:
    build:
      context: ../go/mock/mockxago
    profiles:
      - application
    restart: always
    expose:
      - "8080"
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.mockxago.rule=Host(`mockxago.interledger.test`)"
      - "traefik.http.routers.mockxago.entrypoints=websecure"
      - "traefik.http.services.mockxago.loadbalancer.server.port=8080"
    environment:
      - LOG_LEVEL=debug
      - MOCKXAGO_REDIS_URL=redis:6379
      - MOCKXAGO_REDIS_DB=4
      - WEBHOOK_URL=http://backend:8080/webhooks/xago
      - WEBHOOK_SECRET=xago-webhook-secret
      - XAGO_API_PUBLIC_KEY=test-public-key
      - XAGO_API_SECRET=test-secret
      - XAGO_MOCK_TEST_MODE=true
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 5s
```

**Validation**:
- File exists at `local/mockxago.yaml`
- YAML is valid: `docker-compose config` runs without errors
- All required fields present

### Task 2: Add XAGO Environment Variables to Backend (1 hour)
**Objective**: Configure backend to communicate with MockXago

**Files to modify**: `local/wallet.yaml`

**Action**:
1. Add XAGO environment variables to the `backend` service environment section:
   ```yaml
   # MockXago Persona API endpoint
   - MOCKXAGO_ENDPOINT=http://mockxago:8080
   
   # Xago provider configuration (MockXago for local development)
   - XAGO_API_BASE_URL=http://mockxago:8080/v1
   - XAGO_IDENTITY_BASE_URL=http://mockxago:8080/v1
   - XAGO_API_PUBLIC_KEY=test-public-key
   - XAGO_API_SECRET=test-secret
   
   # Xago operations accounts (used for balance tracking)
   - XAGO_POLICY_ID=5e2585a474b0e90012ce8ff1
   - XAGO_USD_OPS_ACCOUNT=868196c3-f6b4-4920-bbfb-d1c7f6a98183
   - XAGO_ZAR_OPS_ACCOUNT=b0944908-16e6-4ef4-8677-192165e33c59
   
   # Ledger IDs for Xago operations
   - XAGO_LEDGER_ID_ZAR=9246927
   - XAGO_LEDGER_ID_USD=9246873
   ```

2. Ensure `mockxago` service is added to backend depends_on (if applicable)
   ```yaml
   depends_on:
     - mockxago
   ```

**Important notes**:
- `XAGO_API_BASE_URL` and `XAGO_IDENTITY_BASE_URL` must point to MockXago service
- Use `http://mockxago:8080` not `https://mockxago.interledger.test` (internal Docker network)
- Account IDs can be duplicated from bosbaber/mockxago
- Keep credentials as test values

**Validation**:
- `local/wallet.yaml` contains all XAGO_* variables
- YAML is valid: `docker-compose config` succeeds
- Variables match MockXago environment setup

### Task 3: Update Docker Compose Main Configuration (30 min)
**Objective**: Ensure MockXago is included in the main Docker Compose orchestration

**Files to modify**: `local/docker-compose.yaml` (if it exists) or verify service inclusion

**Action**:
1. Check `local/docker-compose.yaml` structure
2. Verify services are properly included:
   - `docker-compose config` shows all services (infrastructure, services, application)
   - MockXago is in the `application` profile

2. Test service connectivity:
   ```bash
   cd local
   docker-compose config | grep -A 5 "mockxago"  # Verify service defined
   ```

**Likely no changes needed** if docker-compose uses profile-based loading

**Validation**:
- `docker-compose config` executes without errors
- `docker-compose config | grep -i mockxago` shows service definition

### Task 4: Verify Redis Configuration for MockXago (30 min)
**Objective**: Ensure Redis is available for MockXago to use

**Action**:
1. Verify `local/redis.yaml` exists and is properly configured
2. Check MockXago uses correct Redis:
   - Connection: `redis:6379` (Docker internal network name)
   - Database: `4` (separate from other services)
   - No authentication required (test environment)

3. Test Redis connectivity:
   ```bash
   docker-compose exec redis redis-cli -n 4 PING
   # Should return: PONG
   ```

**Validation**:
- Redis service is running
- MockXago can connect to Redis DB 4
- `docker-compose logs redis` shows no connection errors

### Task 5: Verify Service Discovery & Networking (30 min)
**Objective**: Confirm services can reach each other

**Action**:
1. Start the full stack:
   ```bash
   cd local
   docker-compose up -d
   ```

2. Verify service availability:
   ```bash
   # Check MockXago health
   docker-compose exec mockxago wget -q -O- http://localhost:8080/health
   
   # Check backend can reach MockXago
   docker-compose exec backend wget -q -O- http://mockxago:8080/health
   
   # Check MockXago can reach backend webhook
   docker-compose exec mockxago wget -q -O- http://backend:8080/health
   ```

3. Check DNS resolution:
   ```bash
   docker-compose exec mockxago getent hosts mockxago
   docker-compose exec backend getent hosts mockxago
   ```

**Validation**:
- All health checks return HTTP 200
- DNS resolution succeeds
- Logs show no connection errors

### Task 6: Test Backend-MockXago Communication (1 hour)
**Objective**: Verify backend and MockXago can exchange messages

**Action**:
1. Create a test user in Xago (via MockXago API):
   ```bash
   curl -X POST http://localhost:8080/1/api/v1/auth/login \
     -H "x-gatehub-app-id: local-test-app" \
     -H "x-gatehub-timestamp: $(date +%s)" \
     -H "x-gatehub-signature: test-sig" \
     -d '{"email":"test@example.com","password":"test"}'
   ```

2. Verify backend receives webhook configuration from MockXago:
   ```bash
   docker logs backend | grep -i "webhook\|xago"
   ```

3. Test webhook delivery (MockXago sends event to backend):
   - Create a deposit in MockXago
   - Backend should receive webhook event
   - Check: `docker logs backend | grep "xago.*webhook\|deposit.*completed"`

**Validation**:
- MockXago API responds correctly
- Backend recognizes Xago webhooks
- Webhook URL is reachable from MockXago
- Event payload processed without errors

### Task 7: Final Environment Validation (30 min)
**Objective**: Complete system validation before proceeding to E2E tests

**Action**:
1. Restart services cleanly:
   ```bash
   cd local
   docker-compose down -v  # Clean state
   docker-compose up -d
   sleep 30  # Wait for startup
   ```

2. Run validation tests:
   ```bash
   # Check all services health
   docker-compose exec mockxago wget -q -O- http://localhost:8080/health
   docker-compose exec backend wget -q -O- http://localhost:8080/health
   
   # Verify Redis
   docker-compose exec redis redis-cli -n 4 DBSIZE
   ```

3. Check logs for errors:
   ```bash
   docker-compose logs --tail=50 | grep -i "error\|fatal"
   # Should return nothing
   ```

**Validation**:
- All services start without errors
- Health checks all pass
- No error logs in any service
- Redis is clean and accessible

## Testing Strategy

### Service Health Checks
```bash
# Individual service health
docker-compose exec mockxago wget -q -O- http://localhost:8080/health
docker-compose exec backend wget -q -O- http://localhost:8080/health
docker-compose exec redis redis-cli ping

# Network connectivity
docker-compose exec backend wget -q -O- http://mockxago:8080/health
docker-compose exec mockxago wget -q -O- http://backend:8080/health
```

### Before Proceeding to Phase 7

All of the following must pass:

```bash
cd local

# 1. Services start without error
docker-compose up -d
sleep 30

# 2. All health endpoints respond
docker-compose exec mockxago wget -q -O- http://localhost:8080/health | grep -q health || false
docker-compose exec backend wget -q -O- http://localhost:8080/health | grep -q health || false

# 3. No error logs
! docker-compose logs | grep -i "error\|fatal"

# 4. Services can communicate
docker-compose exec backend wget -q -O- http://mockxago:8080/health || false

# 5. Redis is operational
docker-compose exec redis redis-cli -n 4 ping
```

## Implementation Notes

### Why Phase 6 is Critical
- E2E tests cannot run without MockXago being accessible
- Backend cannot communicate with Xago without environment variables
- Service discovery (Docker networking) is essential
- Phase 7 depends 100% on Phase 6 being complete

### Configuration Decisions

**Redis Database Selection**:
- MockXago uses DB 4 to avoid conflicts with other services
- Review `local/redis.yaml` to see what DB numbers other services use

**Webhook Secret**:
- Must match between MockXago and backend
- Test environment: simple value like `xago-webhook-secret`
- Production: would need cryptographically secure value

**Traefik Routing**:
- MockXago must be accessible at `mockxago.interledger.test` for E2E tests
- TLS certificates must be valid (local self-signed for dev)
- Traefik is already configured in `local/docker-compose.yaml`

**Service Naming**:
- Docker Compose DNS: `mockxago`, `backend`, `redis` (service names)
- External URLs: `mockxago.interledger.test`, `interledger.test` (Traefik routing)
- Backend uses internal names, E2E tests use external URLs

## Common Issues & Debugging

| Issue | Diagnosis | Fix |
|-------|-----------|-----|
| MockXago fails to start | `docker logs mockxago` shows error | Check Dockerfile, go build errors |
| Backend can't reach MockXago | `docker logs backend \| grep mockxago` | Verify service name in URL, check DNS |
| Webhook delivery fails | `docker logs backend \| grep webhook` | Verify WEBHOOK_URL in mockxago, check firewall |
| Redis connection refuses | `redis-cli -n 4 ping` fails | Verify redis service running, check port 6379 |
| Traefik routing fails | `https://mockxago.interledger.test` unreachable | Check Traefik config, verify certificate |
| Port 8080 already in use | `docker-compose up` fails with port error | Kill process on port 8080 or change mapping |

## Success Criteria

Phase 6 is **COMPLETE** when:

1. ✅ `local/mockxago.yaml` exists and contains all required fields
2. ✅ `local/wallet.yaml` contains all XAGO_* environment variables
3. ✅ `docker-compose config` runs without errors
4. ✅ `docker-compose up -d` starts all services successfully
5. ✅ All health endpoints respond (200 OK)
6. ✅ Services can communicate:
   - Backend can reach MockXago API
   - MockXago can reach backend webhook
   - Redis is accessible to MockXago
7. ✅ No error logs from any service
8. ✅ Manual test: Create Xago user via MockXago API succeeds
9. ✅ Manual test: Webhook delivery from MockXago to backend works

**All tests must pass before Phase 7 can proceed.**

## Notes for Implementation Team

- **Reference**: Use `bosbaber/mockxago:local/mockxago.yaml` as your template
- **Don't guess**: If you're not sure about a config value, find it in bosbaber/mockxago
- **Test incrementally**: Validate each task before moving to the next
- **Clean slate**: `docker-compose down -v` before testing to avoid state issues
- **Log monitoring**: Keep `docker-compose logs -f` running in a separate terminal
- **Save time**: Copy all configs from bosbaber/mockxago rather than reinventing

---

**Phase**: 6 (Infrastructure Setup)
**Prerequisite**: Phase 5 (MockXago Foundation)
**Next Phase**: Phase 7 (Xago E2E Signup)
**Last Updated**: March 6, 2026 (REVISED)
