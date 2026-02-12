#!/bin/bash

# ==========================================
# ZENOENGINE AUTH LOAD TESTS
# ==========================================
# Tests authentication performance under load

set -e

BASE_URL="http://localhost:3000"
RESULTS_FILE="results/load_test_results.txt"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "===========================================" | tee "$RESULTS_FILE"
echo "ZenoEngine Auth Load Tests" | tee -a "$RESULTS_FILE"
echo "Started: $(date)" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Check if Apache Bench is installed
if ! command -v ab &> /dev/null; then
    echo -e "${RED}Error: Apache Bench (ab) is not installed.${NC}"
    echo "Install it with: sudo apt-get install apache2-utils"
    exit 1
fi

echo -e "${YELLOW}Testing with Apache Bench (ab)${NC}" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Test 1: GET Login Page - Light Load
echo "=== Test 1: GET Login Page (100 requests, 10 concurrent) ===" | tee -a "$RESULTS_FILE"
ab -n 100 -c 10 -g results/login_page_100.tsv "$BASE_URL/tutorial/max/login" 2>&1 | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Test 2: GET Login Page - Medium Load
echo "=== Test 2: GET Login Page (500 requests, 50 concurrent) ===" | tee -a "$RESULTS_FILE"
ab -n 500 -c 50 -g results/login_page_500.tsv "$BASE_URL/tutorial/max/login" 2>&1 | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Test 3: GET Login Page - Heavy Load
echo "=== Test 3: GET Login Page (1000 requests, 100 concurrent) ===" | tee -a "$RESULTS_FILE"
ab -n 1000 -c 100 -g results/login_page_1000.tsv "$BASE_URL/tutorial/max/login" 2>&1 | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Test 4: POST Login - Medium Load
echo "=== Test 4: POST Login (200 requests, 20 concurrent) ===" | tee -a "$RESULTS_FILE"
# Create POST data file
echo "email=admin@demo.com&password=password123" > results/login_post_data.txt

ab -n 200 -c 20 \
   -p results/login_post_data.txt \
   -T "application/x-www-form-urlencoded" \
   -g results/login_post_200.tsv \
   "$BASE_URL/tutorial/max/auth/login" 2>&1 | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"

# Clean up
rm -f results/login_post_data.txt

# Summary
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "Load Test Summary" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"
echo "All load tests completed successfully!" | tee -a "$RESULTS_FILE"
echo "Completed: $(date)" | tee -a "$RESULTS_FILE"
echo "" | tee -a "$RESULTS_FILE"
echo "Key Metrics to Review:" | tee -a "$RESULTS_FILE"
echo "- Requests per second (higher is better)" | tee -a "$RESULTS_FILE"
echo "- Time per request (lower is better)" | tee -a "$RESULTS_FILE"
echo "- Failed requests (should be 0)" | tee -a "$RESULTS_FILE"
echo "- 95th percentile response time" | tee -a "$RESULTS_FILE"
echo "===========================================" | tee -a "$RESULTS_FILE"

echo -e "${GREEN}Load tests completed! Check results/load_test_results.txt for details.${NC}"
