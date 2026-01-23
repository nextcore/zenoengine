# ZenoVM (Separated)

**Behavior Change Notice**

ZenoVM has been separated into its own standalone project to decouple the development engine (ZenoEngine) from the production runtime (ZenoVM).

## ℹ️ What does this mean?

1.  **ZenoEngine (This Repository)**
    - Uses **AST Interpretation**.
    - Optimized for **Developer Experience** (Hot Reload, Debugging).
    - `zeno compile` is disabled.

2.  **ZenoVM (Standalone Repository)**
    - Uses **Bytecode Compilation & Execution**.
    - Optimized for **Production Performance**.
    - Supports `.zbc` files.

## 🔗 How to use ZenoVM

To run your code in VM mode, please use the ZenoVM binary:

```bash
# Example
zeno-vm run src/main.zl
```

For more details, please refer to the ZenoVM repository documentation.
