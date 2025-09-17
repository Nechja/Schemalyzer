#!/bin/bash
# Functional tests for schemalyzer fingerprint feature

set -e

echo "=== Schemalyzer Fingerprint Functional Tests ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# PostgreSQL connection
PG_CONN="postgres://testuser:testpass@localhost:5433/testdb?sslmode=disable"

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    local expected_exit="$3"
    
    echo -n "Testing: $test_name... "
    
    if eval "$test_cmd" > /dev/null 2>&1; then
        actual_exit=0
    else
        actual_exit=$?
    fi
    
    if [ "$actual_exit" -eq "$expected_exit" ]; then
        echo -e "${GREEN}PASSED${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}FAILED${NC} (expected exit $expected_exit, got $actual_exit)"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test fingerprint generation
echo "1. Testing fingerprint generation..."
HASH1=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema)
if [ ${#HASH1} -eq 64 ]; then
    echo -e "   Generate PostgreSQL fingerprint: ${GREEN}PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "   Generate PostgreSQL fingerprint: ${RED}FAILED${NC} (invalid hash length)"
    ((TESTS_FAILED++))
fi

# Test deterministic hashing (same schema = same hash)
echo
echo "2. Testing deterministic hashing..."
HASH2=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema)
if [ "$HASH1" = "$HASH2" ]; then
    echo -e "   Same schema produces same hash: ${GREEN}PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "   Same schema produces same hash: ${RED}FAILED${NC}"
    ((TESTS_FAILED++))
fi

# Test schema comparison - matching
echo
echo "3. Testing schema comparison..."
run_test "Matching schemas" \
    "./schemalyzer compare-fingerprints \
        --source-type postgresql --source-conn '$PG_CONN' --source-schema test_schema \
        --target-type postgresql --target-conn '$PG_CONN' --target-schema test_schema" \
    0

# Test schema comparison - different
run_test "Different schemas" \
    "./schemalyzer compare-fingerprints \
        --source-type postgresql --source-conn '$PG_CONN' --source-schema test_schema \
        --target-type postgresql --target-conn '$PG_CONN' --target-schema test_schema2" \
    2

# Test with pre-computed hashes
echo
echo "4. Testing pre-computed hash comparison..."
HASH_A=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema)
HASH_B=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema2)

run_test "Pre-computed matching hashes" \
    "./schemalyzer compare-fingerprints --source-hash '$HASH_A' --target-hash '$HASH_A'" \
    0

run_test "Pre-computed different hashes" \
    "./schemalyzer compare-fingerprints --source-hash '$HASH_A' --target-hash '$HASH_B'" \
    2

# Test JSON output
echo
echo "5. Testing JSON output..."
JSON_OUTPUT=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema --json)
if echo "$JSON_OUTPUT" | grep -q '"fingerprint"' && echo "$JSON_OUTPUT" | grep -q '"algorithm":"SHA256"'; then
    echo -e "   JSON output format: ${GREEN}PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "   JSON output format: ${RED}FAILED${NC}"
    ((TESTS_FAILED++))
fi

# Test verbose output
echo
echo "6. Testing verbose output..."
VERBOSE_OUTPUT=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema --verbose)
if echo "$VERBOSE_OUTPUT" | grep -q "Tables:" && echo "$VERBOSE_OUTPUT" | grep -q "SHA256"; then
    echo -e "   Verbose output format: ${GREEN}PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "   Verbose output format: ${RED}FAILED${NC}"
    ((TESTS_FAILED++))
fi

# Test tables-only flag
echo
echo "7. Testing tables-only flag..."
FULL_HASH=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema)
TABLES_HASH=$(./schemalyzer fingerprint --type postgresql --conn "$PG_CONN" --schema test_schema --tables-only)
if [ "$FULL_HASH" != "$TABLES_HASH" ]; then
    echo -e "   Tables-only produces different hash: ${GREEN}PASSED${NC}"
    ((TESTS_PASSED++))
else
    echo -e "   Tables-only produces different hash: ${RED}FAILED${NC}"
    ((TESTS_FAILED++))
fi

# Summary
echo
echo "======================================="
echo "Test Summary:"
echo -e "  Passed: ${GREEN}$TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "  Failed: ${RED}$TESTS_FAILED${NC}"
else
    echo -e "  Failed: $TESTS_FAILED"
fi
echo "======================================="

if [ $TESTS_FAILED -gt 0 ]; then
    exit 1
fi

echo
echo -e "${GREEN}All tests passed!${NC}"