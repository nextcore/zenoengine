# ZenoEngine Testing Suite

This directory contains comprehensive tests for the ZenoEngine authentication module.

## Test Types

### 1. Integration Tests
**File**: `auth_integration_test.sh`

Tests complete authentication flows including:
- Login page rendering
- User registration
- User login with valid/invalid credentials
- Authenticated access to protected routes
- Logout functionality
- Cookie/session management

**Usage**:
```bash
# Make sure ZenoEngine server is running first
./zeno

# In another terminal, run integration tests
cd tests
chmod +x auth_integration_test.sh
./auth_integration_test.sh
```

**Expected Results**:
- All tests should pass (green ✓)
- Results saved to `results/integration_results.txt`

### 2. Load Tests
**File**: `load_test_auth.sh`

Tests authentication performance under load using Apache Bench:
- Light load: 100 requests, 10 concurrent
- Medium load: 500 requests, 50 concurrent
- Heavy load: 1000 requests, 100 concurrent
- POST login: 200 requests, 20 concurrent

**Prerequisites**:
```bash
# Install Apache Bench
sudo apt-get install apache2-utils
```

**Usage**:
```bash
# Make sure ZenoEngine server is running first
./zeno

# In another terminal, run load tests
cd tests
chmod +x load_test_auth.sh
./load_test_auth.sh
```

**Expected Results**:
- 0 failed requests
- Average response time < 100ms
- Requests per second > 100
- Results saved to `results/load_test_results.txt`

## Test Results

All test results are saved in the `results/` directory:
- `integration_results.txt` - Integration test results
- `load_test_results.txt` - Load test summary
- `*.tsv` - Detailed timing data for load tests

## Success Criteria

### Integration Tests
- ✅ All tests pass (100% success rate)
- ✅ Login/logout flows work correctly
- ✅ Authentication cookies are set/cleared properly
- ✅ Invalid credentials are rejected

### Load Tests
- ✅ Handle 500+ concurrent requests without errors
- ✅ Average response time < 100ms
- ✅ No memory leaks or crashes
- ✅ Consistent performance across test runs

## Troubleshooting

### Integration Tests Failing
1. **Check server is running**: `curl http://localhost:3000/tutorial/max/login`
2. **Check database**: Ensure `tutorial_max.db` exists and has demo users
3. **Check logs**: Review ZenoEngine server logs for errors

### Load Tests Failing
1. **Increase system limits**: `ulimit -n 10000`
2. **Check server resources**: Monitor CPU/memory during tests
3. **Reduce concurrency**: Start with lower concurrent requests

## Running All Tests

```bash
# Run both integration and load tests
cd tests
./auth_integration_test.sh && ./load_test_auth.sh
```

## Continuous Integration

These tests can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Integration Tests
  run: |
    ./zeno &
    sleep 5
    cd tests && ./auth_integration_test.sh
    
- name: Run Load Tests
  run: |
    cd tests && ./load_test_auth.sh
```
