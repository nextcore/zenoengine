# Rust PHP Native Bridge

This directory contains a **Rust implementation** of the Zeno PHP Bridge sidecar. It mimics the behavior of the Zig version but leverages Rust's ecosystem for those who prefer it.

## Features
- **JSON-RPC Protocol**: Fully compatible with Zeno sidecar protocol.
- **Embedded PHP**: Designed to statically link `libphp` for a single-binary distribution.
- **Cross-Platform**: Supports Windows, Linux, and macOS.

## Building (Static Linking)

To build a truly portable binary that contains the PHP runtime, you must link against the static version of the PHP library.

### Windows (MSVC)

1.  **Download PHP Development Pack**:
    - Go to [windows.php.net/download](https://windows.php.net/download/)
    - Download the **Thread Safe (TS)** or **Non Thread Safe (NTS)** version (matching your Zeno build).
    - Download the **Development Package (SDK)** for that version.

2.  **Extract**:
    - Extract the SDK, e.g., to `C:\php-sdk`.

3.  **Configure `build.rs`**:
    - Edit `build.rs` in this directory.
    - Update the path `C:/php/dev/lib` to point to your extracted SDK `lib` folder (containing `php8.lib`).

4.  **Build**:
    ```powershell
    cargo build --release
    ```

5.  **Runtime DLLs**:
    - Even with static linking, on Windows you might still need `php8.dll` in the same folder as the executable unless you compile PHP itself statically from source (complex).
    - Copy `php8.dll` from the PHP binary package to the `target/release` folder.

### Linux / macOS

1.  **Install PHP Dev**:
    - Ubuntu/Debian: `sudo apt install php-dev`
    - macOS: `brew install php`

2.  **Build**:
    - Rust's `cc` crate usually finds the library automatically if `php-config` is in your PATH.
    ```bash
    cargo build --release
    ```

## Usage

Update `manifest.yaml` in the parent directory to point to the compiled binary:

```yaml
# manifest.yaml
binary: ./rust-src/target/release/php-native-bridge.exe # Windows
# or
binary: ./rust-src/target/release/php-native-bridge     # Linux/Mac
```
