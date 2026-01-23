# Release Notes v0.6.3

## ⚡ Major Changes

### ZenoVM Separation
ZenoVM has been decoupled from the core ZenoEngine repository.
- **ZenoEngine** (this repo) now focuses on the Developer Experience (DX) with the AST Interpreter.
- **ZenoVM** is now a standalone project for production deployments.
- `zeno compile` command has been disabled.

## 🧹 Improvements

- **Debug Logs Cleanup**: Removed verbose debug logging from `router.go`, `utils.go`, and `functions.go`.
- **Documentation**:
    - Added [MIGRATION_GUIDE.md](MIGRATION_GUIDE.md).
    - Updated [ZENOVM.md](ZENOVM.md) to reflect the separation.
    - Updated `README.md`.

## 🐛 Bug Fixes
- Added request timeouts in `router.go` to prevent hanging requests.
- Integrated `MultiTenantAuth` native middleware.
