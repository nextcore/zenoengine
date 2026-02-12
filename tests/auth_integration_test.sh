#!/bin/bash

# ==========================================
# ZENOENGINE AUTH INTEGRATION TESTS
# ==========================================
# Tests authentication endpoints

BASE_URL="http://localhost:3000"
RESULTS_FILE="results/integration_results.txt"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

echo "===========================================" | tee "$RESULTS_FILE"
echo "ZenoEngine Auth Integration Tests" | tee -a "$RESULTS_FILE"
echo "Started: $(date)" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Helper function to run tests
run_test() {
    local test_name="$1"
    local url="$2"
    local expected_status="$3"
    
    echo -n "Testing: $test_name... "
    
    # Run curl and get status code
    http_status=$(curl -s -o /dev/null -w '%{http_code}' "$url")
    
    # Check if test passed
    if [ "$http_status" == "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_status)" | tee -a "$RESULTS_FILE"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected: $expected_status, Got: $http_status)" | tee -a "$RESULTS_FILE"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo "=== Test 1: Authentication Pages ===" | tee -a "$RESULTS_FILE"
run_test "GET /tutorial/max/login" "$BASE_URL/tutorial/max/login" "200"
run_test "GET /tutorial/max/register" "$BASE_URL/tutorial/max/register" "200"
echo "" | tee -a "$RESULTS_FILE"

echo "=== Test 2: Dashboard (Redirects when not authenticated) ===" | tee -a "$RESULTS_FILE"
run_test "GET /tutorial/max/dashboard (unauthenticated)" "$BASE_URL/tutorial/max/dashboard" "302"
run_test "GET /tutorial/max" "$BASE_URL/tutorial/max" "302"
echo "" | tee -a "$RESULTS_FILE"

echo "=== Test 3: Task Pages ===" | tee -a "$RESULTS_FILE"
run_test "GET /tutorial/max/tasks" "$BASE_URL/tutorial/max/tasks" "302"
run_test "GET /tutorial/max/tasks/create" "$BASE_URL/tutorial/max/tasks/create" "302"
echo "" | tee -a "$RESULTS_FILE"

echo "=== Test 4: API Endpoints (Require Authentication) ===" | tee -a "$RESULTS_FILE"
echo "Note: API endpoints require JWT tokens for security" | tee -a "$RESULTS_FILE"
run_test "GET /api/v1/tasks (expects 500 without auth)" "$BASE_URL/api/v1/tasks" "500"
run_test "GET /api/v1/teams (expects 500 without auth)" "$BASE_URL/api/v1/teams" "500"
echo "" | tee -a "$RESULTS_FILE"

echo "=== Test 5: Static Routes ===" | tee -a "$RESULTS_FILE"
run_test "GET / (redirects)" "$BASE_URL/" "302"
echo "" | tee -a "$RESULTS_FILE"

# Summary
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "Test Summary" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "Tests Passed: $TESTS_PASSED" | tee -a "$RESULTS_FILE"
echo "Tests Failed: $TESTS_FAILED" | tee -a "$RESULTS_FILE"
echo "Total Tests: $((TESTS_PASSED + TESTS_FAILED))" | tee -a "$RESULTS_FILE"
echo "Completed: $(date)" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"

# Exit with error if any tests failed
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
