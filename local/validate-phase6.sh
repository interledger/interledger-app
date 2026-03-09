#!/bin/bash
# Phase 6 Validation Script
# Tests the local environment setup for Xago integration

set -e

echo "==================================================================="
echo "Phase 6 Validation: Xago Local Environment Setup"
echo "==================================================================="
echo ""

cd /home/stephan/interledger/interledger-app/local

echo "✓ Step 1: Validating Docker Compose configuration..."
docker compose config --quiet
echo "  Configuration is valid"
echo ""

echo "✓ Step 2: Verifying mockxago service is included..."
if docker compose --profile application config | grep -q "mockxago:"; then
    echo "  mockxago service found in application profile"
else
    echo "  ERROR: mockxago service not found"
    exit 1
fi
echo ""

echo "✓ Step 3: Checking files created..."
for file in mockxago.yaml; do
    if [ -f "$file" ]; then
        echo "  ✓ $file exists"
    else
        echo "  ✗ $file missing"
        exit 1
    fi
done
echo ""

echo "✓ Step 4: Verifying XAGO environment variables in wallet.yaml..."
if grep -q "XAGO_API_BASE_URL" wallet.yaml; then
    echo "  ✓ XAGO_API_BASE_URL found"
else
    echo "  ✗ XAGO_API_BASE_URL missing"
    exit 1
fi

if grep -q "XAGO_WEBHOOK_SECRET" wallet.yaml; then
    echo "  ✓ XAGO_WEBHOOK_SECRET found"
else
    echo "  ✗ XAGO_WEBHOOK_SECRET missing"
    exit 1
fi

if grep -q "MOCKXAGO_ENDPOINT" wallet.yaml; then
    echo "  ✓ MOCKXAGO_ENDPOINT found"
else
    echo "  ✗ MOCKXAGO_ENDPOINT missing"
    exit 1
fi

if grep -A 5 "depends_on:" wallet.yaml | grep -q "mockxago"; then
    echo "  ✓ mockxago in backend depends_on"
else
    echo "  ✗ mockxago not in backend depends_on"
    exit 1
fi
echo ""

echo "==================================================================="
echo "Phase 6 Configuration Validation: PASSED ✓"
echo "==================================================================="
echo ""
echo "Next Steps to Complete Phase 6:"
echo "--------------------------------"
echo ""
echo "1. Start the full environment:"
echo "   cd local"
echo "   make all"
echo "   # or: docker compose --profile infrastructure --profile services --profile application up -d"
echo ""
echo "2. Wait for services to start (~30 seconds)"
echo ""
echo "3. Verify health checks:"
echo "   docker compose exec mockxago curl -f http://localhost:8080/health"
echo "   docker compose exec backend curl -f http://localhost:8080/health"
echo ""
echo "4. Test inter-service communication:"
echo "   docker compose exec backend curl -f http://mockxago:8080/health"
echo "   docker compose exec mockxago curl -f http://backend:8080/health"
echo ""
echo "5. Verify Redis connectivity:"
echo "   docker compose exec redis redis-cli -n 4 PING"
echo ""
echo "6. Check service logs:"
echo "   docker compose logs mockxago | tail -n 50"
echo "   docker compose logs backend | tail -n 50"
echo ""
echo "==================================================================="
