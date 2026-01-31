# WASM Plugin Configuration Guide

## Configuration (.env)

The easiest way to configure WASM plugins is via your `.env` file. This is especially recommended for Windows users.

### 1. Enable Plugins

Add this to your `.env`:
```bash
ZENO_PLUGINS_ENABLED=true
```

### 2. Set Plugin Directory (Optional)

Default is `./plugins`. You can change it if needed:
```bash
ZENO_PLUGIN_DIR=./plugins
```

---

## Alternative: Environment Variables (Linux/macOS)

You can also use `export` if you prefer not to use a `.env` file:

```bash
export ZENO_PLUGINS_ENABLED=true
export ZENO_PLUGIN_DIR=./plugins
```

## Configuration File

Create `config/plugins.yaml` (optional):

```yaml
wasm:
  enabled: true
  plugin_dir: ./plugins
  
  # Global resource limits
  limits:
    max_memory: 100MB
    max_execution_time: 30s
    max_cpu: 0.5
  
  # Security settings
  security:
    strict_mode: true
    allow_network: true
    allow_filesystem: false
  
  # Plugin whitelist (optional)
  whitelist:
    - hello
    - stripe
    - aws
```

## Plugin Directory Structure

```
plugins/
├── hello/
│   ├── manifest.yaml
│   └── hello.wasm
├── stripe/
│   ├── manifest.yaml
│   └── stripe.wasm
└── aws/
    ├── manifest.yaml
    └── aws.wasm
```

## Usage in ZenoLang Scripts

Once plugins are loaded, their slots are available like any other slot:

```zenolang
# Use hello plugin
hello.greet {
    name: "World"
}

# Result stored in scope
log.info: $message  # "Hello, World! 👋"

# Use stripe plugin (example)
stripe.charge {
    amount: 1000
    token: $payment_token
    currency: "usd"
}

# Check result
if: "$charge_id != nil" {
    then {
        log.info: "Payment successful!"
    }
}
```

## Host Functions Available to Plugins

Plugins can call these host functions:

### Logging
```go
host_log("info", "Message from plugin")
```

### Database Queries
```go
host_db_query(connection, sql, params)
```

### HTTP Requests
```go
host_http_request(method, url, headers, body)
```

### Scope Access
```go
host_scope_get(key)
host_scope_set(key, value)
```

### File Operations (if permitted)
```go
host_file_read(path)
host_file_write(path, content)
```

### Environment Variables (if permitted)
```go
host_env_get(key)
```

## Security

### Permission System

Plugins must declare permissions in `manifest.yaml`:

```yaml
permissions:
  network:
    - https://api.stripe.com/*
  env:
    - STRIPE_API_KEY
  scope:
    - read
    - write
  filesystem:
    - /tmp/plugin-data
  database:
    - default
```

### Sandboxing

- Plugins run in isolated WASM sandbox
- No direct memory access
- All operations go through host functions
- Permission checks enforced at runtime

## Troubleshooting

### Plugins Not Loading

1. Check `ZENO_PLUGINS_ENABLED=true`
2. Verify plugin directory exists
3. Check manifest.yaml syntax
4. Ensure .wasm file is present
5. Check logs for errors

### Plugin Execution Errors

1. Check plugin permissions
2. Verify slot parameters
3. Check host function availability
4. Review plugin logs

### Performance Issues

1. Check plugin resource limits
2. Monitor execution time
3. Profile WASM execution
4. Consider plugin optimization

## Examples

See `examples/wasm-plugins/` for working examples:

- `hello-go/` - Simple Go plugin
- More examples coming soon!

## Development

### Creating a Plugin

1. Write plugin code (Go/Rust/etc.)
2. Create manifest.yaml
3. Build to WASM
4. Test locally
5. Deploy to plugins directory

### Testing

```bash
# Enable plugins
export ZENO_PLUGINS_ENABLED=true

# Run ZenoEngine
zeno run test.zl

# Check logs
tail -f logs/app.log | grep "WASM Plugin"
```

## Best Practices

1. **Keep plugins focused** - One plugin, one purpose
2. **Minimize dependencies** - Smaller WASM = faster loading
3. **Handle errors gracefully** - Always return proper responses
4. **Document thoroughly** - Clear manifest and README
5. **Test extensively** - Unit tests + integration tests
6. **Version carefully** - Use semantic versioning
7. **Monitor performance** - Track execution time

## Limitations

- Plugins cannot directly access Go runtime
- All operations must go through host functions
- Resource limits enforced
- Permission system required
- WASM overhead (~10-20% vs native)

## Future Enhancements

- [ ] Plugin hot-reload
- [ ] Plugin marketplace
- [ ] Plugin versioning & updates
- [ ] Plugin dependency management
- [ ] Enhanced debugging tools
- [ ] Performance profiling
- [ ] Plugin analytics
