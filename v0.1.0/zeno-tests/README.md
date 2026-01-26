# ZenoEngine Cross-Platform Test Suite

## Purpose

This test suite ensures **100% compatibility** across all ZenoEngine editions:
- Go Edition (reference implementation)
- .NET 10 Edition
- Rust Edition

## Directory Structure

```
zeno-tests/
├─ syntax/           # Syntax and language feature tests
├─ runtime/          # Runtime behavior tests (HTTP, DB, Files)
├─ integration/      # End-to-end integration tests
└─ expected/         # Expected outputs for each test
```

## Test Files

### Syntax Tests
- `001_basic_routing.zl` - Basic HTTP routing
- `002_variables.zl` - Variable assignment and type coercion
- `003_conditionals.zl` - Conditional logic (if-then-else)

### Runtime Tests
- `010_database_query.zl` - Database queries (raw SQL)
- `011_file_upload.zl` - File upload/download
- `012_image_processing.zl` - Image processing (resize, thumbnail)

## Running Tests

### On Go Edition (Baseline)
```bash
cd "d:\Z\HL\HL Go\ZenoEngine"
./zeno.exe run zeno-tests/syntax/001_basic_routing.zl
```

### On .NET Edition (Future)
```bash
cd zeno-dotnet
dotnet run -- test ../zeno-tests/syntax/001_basic_routing.zl
```

### On Rust Edition (Future)
```bash
cd zeno-rust
cargo run -- test ../zeno-tests/syntax/001_basic_routing.zl
```

## Test Runner

Use `test_compatibility.sh` to run all tests across all platforms:

```bash
./test_compatibility.sh

# Output:
Testing ZenoLang Compatibility...
  Go:     6/6 ✅
  .NET:   6/6 ✅
  Rust:   6/6 ✅

✅ 100% Compatibility Achieved
```

## Adding New Tests

1. Create `.zl` file in appropriate directory
2. Run test on Go edition
3. Save expected output to `expected/`
4. Verify on all platforms

## Success Criteria

A test passes if:
- ✅ Syntax parses identically
- ✅ Execution produces identical results
- ✅ Error messages match (if any)
- ✅ Performance within acceptable range

## Notes

- Tests MUST be platform-agnostic
- Use parameterized SQL queries (no hardcoded values)
- Avoid platform-specific features
- Keep tests simple and focused
