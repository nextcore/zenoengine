# WASM Plugin CLI Documentation

Complete guide for managing WASM plugins using the ZenoEngine CLI.

---

## Overview

The WASM Plugin CLI provides commands to list, inspect, and validate WASM plugins. These tools help developers manage plugins efficiently and ensure quality before deployment.

## Commands

### `zeno plugin help`

Display help information for plugin commands.

```bash
zeno plugin help
```

**Output:**
```
WASM Plugin Management Commands

Usage:
  zeno plugin <command> [arguments]

Available Commands:
  list              List all installed plugins
  info <name>       Show detailed information about a plugin
  validate <path>   Validate a plugin before installation
  help              Show this help message
```

---

### `zeno plugin list`

List all installed WASM plugins with basic information.

**Usage:**
```bash
zeno plugin list
```

**Example Output:**
```
Installed WASM Plugins:
============================================================

  📦 hello-rust (v1.0.0)
     Hello World plugin written in Rust
     Location: plugins\hello-rust
     Binary: hello.wasm

  📦 stripe-payment (v2.1.0)
     Stripe payment processing plugin
     Location: plugins\stripe-payment
     Binary: stripe.wasm
```

**What It Shows:**
- Plugin name and version
- Short description
- Installation location
- Binary filename

**Use Cases:**
- Quick overview of installed plugins
- Check plugin versions
- Verify plugin installation

---

### `zeno plugin info <name>`

Show detailed information about a specific plugin including permissions and metadata.

**Usage:**
```bash
zeno plugin info <plugin-name>
```

**Example:**
```bash
zeno plugin info hello-rust
```

**Example Output:**
```
============================================================
Plugin: hello-rust
============================================================
Version:     1.0.0
Author:      ZenoEngine Team
License:     MIT
Description: Hello World plugin written in Rust
Binary:      hello.wasm

Permissions:
  Scope:      read, write
  Network:    (none)
  Filesystem: (none)
  Database:   (none)
  Env:        (none)
```

**What It Shows:**
- Complete metadata (version, author, license, description)
- Binary filename
- Detailed permissions breakdown:
  - **Scope:** Read/write access to ZenoLang scope
  - **Network:** Allowed URLs/patterns
  - **Filesystem:** Allowed file paths
  - **Database:** Allowed database names
  - **Env:** Allowed environment variables

**Use Cases:**
- Review plugin permissions before use
- Check plugin compatibility
- Understand plugin capabilities
- Security audit

**Error Handling:**
```bash
# Plugin not found
$ zeno plugin info nonexistent
Error: Plugin 'nonexistent' not found or invalid
Details: open plugins/nonexistent/manifest.yaml: no such file or directory
```

---

### `zeno plugin reload [name]`

Reload a specific plugin or all plugins without restarting the server.

**Usage:**
```bash
# Reload specific plugin
zeno plugin reload <plugin-name>

# Reload all plugins
zeno plugin reload
```

**Example:**
```bash
zeno plugin reload hello-rust
```

**What It Does:**
- Unloads the existing plugin
- Reloads code from disk
- Re-registers all slots
- Updates configuration

**Use Cases:**
- Development workflow (hot reload)
- Updating plugins in production
- Resetting plugin state

---

### `zeno plugin validate <path>`

Validate a plugin directory before installation with comprehensive checks.

**Usage:**
```bash
zeno plugin validate <path-to-plugin>
```

**Example:**
```bash
zeno plugin validate ./examples/wasm-plugins/hello-rust
```

**Example Output (Success):**
```
Validating plugin at: ./plugins/hello-rust
============================================================
✅ Plugin directory exists
✅ Manifest found and valid
✅ WASM binary found: hello.wasm (130.5 KB)
✅ Permissions structure valid
============================================================

✅ Validation passed!
Plugin is valid and ready to install!
```

**Example Output (With Warnings):**
```
Validating plugin at: ./my-plugin
============================================================
✅ Plugin directory exists
✅ Manifest found and valid
⚠️  Optional field 'author' not set
⚠️  Optional field 'license' not set
✅ WASM binary found: plugin.wasm (45.2 KB)
✅ Permissions structure valid
============================================================

⚠️  Validation passed with 2 warning(s)
Plugin is valid but could be improved.
```

**Example Output (Failure):**
```
Validating plugin at: ./broken-plugin
============================================================
✅ Plugin directory exists
❌ Manifest invalid: yaml: unmarshal errors
❌ WASM binary not found: plugin.wasm
============================================================

❌ Validation failed with 2 error(s) and 0 warning(s)
Plugin is NOT ready to install.
```

**Validation Checks:**

| Check | Type | Description |
|-------|------|-------------|
| Directory exists | Error | Plugin directory must exist |
| Manifest valid | Error | `manifest.yaml` must be valid YAML |
| Required: name | Error | Plugin must have a name |
| Required: version | Error | Plugin must have a version |
| Required: binary | Error | Plugin must specify WASM binary |
| Optional: author | Warning | Author field recommended |
| Optional: description | Warning | Description field recommended |
| Optional: license | Warning | License field recommended |
| Binary exists | Error | WASM file must exist |
| Binary extension | Warning | Should have `.wasm` extension |
| Permissions valid | Error | Permissions structure must be valid |

**Use Cases:**
- Pre-installation validation
- Development testing
- CI/CD pipeline checks
- Quality assurance

**Exit Codes:**
- `0` - Validation passed (with or without warnings)
- `1` - Validation failed (errors found)

---

## Configuration

### Plugin Directory

The plugin directory is configured via environment variable:

```bash
# In .env file
ZENO_PLUGIN_DIR=./plugins
```

**Default:** `./plugins`

All CLI commands respect this configuration.

---

## Workflow Examples

### 1. Installing a New Plugin

```bash
# Step 1: Validate the plugin
zeno plugin validate ./downloads/my-plugin

# Step 2: Copy to plugins directory (if valid)
cp -r ./downloads/my-plugin ./plugins/

# Step 3: Verify installation
zeno plugin list

# Step 4: Check details
zeno plugin info my-plugin
```

### 2. Security Audit

```bash
# List all plugins
zeno plugin list

# Check each plugin's permissions
zeno plugin info plugin-name

# Look for:
# - Excessive network permissions
# - Broad filesystem access
# - Unnecessary env var access
```

### 3. Development Workflow

```bash
# During development
cd my-plugin/

# Validate frequently
zeno plugin validate .

# Fix issues until validation passes
# Then copy to plugins directory
```

### 4. CI/CD Integration

```bash
#!/bin/bash
# validate-plugins.sh

for plugin in ./plugins/*; do
    echo "Validating $plugin..."
    if ! zeno plugin validate "$plugin"; then
        echo "❌ Validation failed for $plugin"
        exit 1
    fi
done

echo "✅ All plugins valid"
```

---

## Troubleshooting

### Plugin Not Found

**Problem:**
```
Error: Plugin 'my-plugin' not found or invalid
```

**Solutions:**
1. Check plugin name: `zeno plugin list`
2. Verify directory exists: `ls plugins/my-plugin`
3. Check `ZENO_PLUGIN_DIR` environment variable

### Validation Fails

**Problem:**
```
❌ Manifest invalid: yaml: unmarshal errors
```

**Solutions:**
1. Check YAML syntax in `manifest.yaml`
2. Ensure all required fields present
3. Validate YAML online: https://www.yamllint.com/

### Binary Not Found

**Problem:**
```
❌ WASM binary not found: hello.wasm
```

**Solutions:**
1. Check binary filename in manifest matches actual file
2. Ensure binary is in plugin directory
3. Rebuild WASM if necessary

---

## Best Practices

### 1. Always Validate Before Installing
```bash
zeno plugin validate ./new-plugin
```

### 2. Review Permissions Carefully
```bash
zeno plugin info plugin-name
# Check Network, Filesystem, Env permissions
```

### 3. Use Descriptive Metadata
```yaml
# manifest.yaml
name: my-plugin
version: 1.0.0
author: Your Name
description: Clear description of what plugin does
license: MIT
```

### 4. Keep Plugins Updated
```bash
# Regular audit
zeno plugin list
# Check for outdated versions
```

### 5. Document Custom Plugins
- Add README.md to plugin directory
- Document required permissions
- Provide usage examples

---

## Related Documentation

- [WASM Plugin Development Guide](WASM_PLUGIN_QUICKREF.md)
- [Plugin Manifest Reference](WASM_PLUGIN_CONFIG.md)
- [Security Best Practices](WASM_PLUGIN_PROGRESS.md)

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/yourusername/zenoengine/issues
- Documentation: https://zenoengine.dev/docs/plugins
