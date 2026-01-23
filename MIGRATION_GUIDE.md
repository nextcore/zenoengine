# Migration Guide: ZenoEngine & ZenoVM Separation

ZenoEngine has been refactored to focus on **Development Experience** (AST Interpretation), while **ZenoVM** has been separated into a standalone high-performance execution engine.

## ⚠️ Important Changes

### 1. ZenoVM Separation
The bytecode compiler and VM runtime have been moved to a separate repository (`github.com/user/ZenoVM`).
- **ZenoEngine (This Repo)**: Uses AST Interpreter. Best for hot-reloading, development, and debugging.
- **ZenoVM (Standalone)**: Uses Bytecode. Best for production performance and code protection.

### 2. Command Changes
- `zeno compile` is deprecated in this repository.
- Use `zeno-vm compile` (from the ZenoVM project) for bytecode compilation.

### 3. Usage Strategy
- **Development**: Use `zeno run src/main.zl` (AST Mode) for instant feedback.
- **Production**: Deploy with ZenoVM for maximum throughput.

## 🚀 Moving Forward

If you were relying on experimental VM flags (`--vm`):
1. Remove them from your startup scripts.
2. If you need raw performance, migrate your deployment to use the standalone ZenoVM binary.

## ❓ Need Help?

- Check [LANGUAGE_SPECIFICATION.md](LANGUAGE_SPECIFICATION.md) for valid syntax.
- See the ZenoVM repository for specific VM documentation.
