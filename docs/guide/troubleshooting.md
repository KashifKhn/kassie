# Troubleshooting

This guide helps you diagnose and resolve common issues with Kassie.

## Connection Issues

### Cannot connect to host

**Error message**:
```
Error: failed to connect: dial tcp 127.0.0.1:9042: connect: connection refused
```

**Causes and solutions**:

1. **Cassandra/ScyllaDB not running**
   ```bash
   # Check if service is running
   systemctl status cassandra
   # or
   docker ps | grep scylla
   ```

2. **Wrong host or port**
   - Verify host in config: `~/.config/kassie/config.json`
   - Default CQL port is 9042
   - Check database logs for actual port

3. **Firewall blocking connection**
   ```bash
   # Test connectivity
   telnet 127.0.0.1 9042
   # or
   nc -zv 127.0.0.1 9042
   ```

4. **Database bound to different interface**
   ```bash
   # Check listening addresses
   netstat -tlnp | grep 9042
   ```
   
   Update database config to bind to 0.0.0.0 or specific IP.

### Authentication failed

**Error message**:
```
Error: authentication failed: Bad credentials
```

**Solutions**:

1. **Verify credentials**
   ```json
   {
     "auth": {
       "username": "cassandra",
       "password": "cassandra"
     }
   }
   ```

2. **Check environment variables**
   ```bash
   # If using ${VAR_NAME} syntax
   echo $CASSANDRA_PASSWORD
   ```

3. **Test credentials with cqlsh**
   ```bash
   cqlsh -u cassandra -p cassandra 127.0.0.1 9042
   ```

4. **Reset Cassandra password**
   ```bash
   # Default superuser
   cqlsh -u cassandra -p cassandra
   ALTER USER cassandra WITH PASSWORD 'new_password';
   ```

### SSL/TLS connection failed

**Error message**:
```
Error: TLS handshake failed: certificate verify failed
```

**Solutions**:

1. **Verify SSL is enabled on database**
   - Check Cassandra/ScyllaDB config
   - Ensure `client_encryption_options.enabled: true`

2. **Check certificate paths**
   ```json
   {
     "ssl": {
       "enabled": true,
       "cert_path": "/path/to/client.crt",
       "key_path": "/path/to/client.key",
       "ca_path": "/path/to/ca.crt"
     }
   }
   ```

3. **Verify certificate files exist and are readable**
   ```bash
   ls -l /path/to/*.crt /path/to/*.key
   chmod 600 /path/to/client.key
   ```

4. **Test SSL connection**
   ```bash
   openssl s_client -connect 127.0.0.1:9042 -CAfile ca.crt
   ```

### Connection timeout

**Error message**:
```
Error: query timeout after 10000ms
```

**Solutions**:

1. **Increase timeout in config**
   ```json
   {
     "timeout_ms": 30000
   }
   ```

2. **Check network latency**
   ```bash
   ping database-host
   ```

3. **Verify database performance**
   - Check database CPU/memory usage
   - Look for slow queries in database logs
   - Check for compaction or repair operations

4. **Reduce page size**
   ```json
   {
     "defaults": {
       "page_size": 50
     }
   }
   ```

## Configuration Issues

### Config file not found

**Error message**:
```
Warning: config file not found, using defaults
```

**Expected behavior**: Kassie will use built-in defaults and try to connect to `127.0.0.1:9042`.

**To fix**:

1. **Create config file**
   ```bash
   mkdir -p ~/.config/kassie
   nano ~/.config/kassie/config.json
   ```

2. **Use custom location**
   ```bash
   kassie tui --config /path/to/config.json
   ```

3. **Verify file exists**
   ```bash
   ls -la ~/.config/kassie/config.json
   ```

### Invalid JSON syntax

**Error message**:
```
Error: invalid config: unexpected token at line 5
```

**Solutions**:

1. **Validate JSON**
   ```bash
   # Use jq to validate
   jq . ~/.config/kassie/config.json
   ```

2. **Common JSON errors**
   - Missing comma between fields
   - Trailing comma at end of array/object
   - Unescaped quotes in strings
   - Missing closing brace/bracket

3. **Example valid JSON**
   ```json
   {
     "version": "1.0",
     "profiles": [
       {
         "name": "local",
         "hosts": ["127.0.0.1"],
         "port": 9042
       }
     ]
   }
   ```

### Profile not found

**Error message**:
```
Error: profile 'production' not found in config
```

**Solutions**:

1. **Check profile name**
   ```bash
   # List all profiles
   jq '.profiles[].name' ~/.config/kassie/config.json
   ```

2. **Use correct profile name**
   ```bash
   kassie tui --profile local
   ```

3. **Add missing profile**
   ```json
   {
     "profiles": [
       {
         "name": "production",
         "hosts": ["prod.example.com"],
         "port": 9042
       }
     ]
   }
   ```

### Environment variable not expanded

**Issue**: Password shows as literal `${VAR_NAME}` instead of value.

**Solutions**:

1. **Verify environment variable is set**
   ```bash
   echo $CASSANDRA_PASSWORD
   ```

2. **Export variable before running Kassie**
   ```bash
   export CASSANDRA_PASSWORD="secret123"
   kassie tui
   ```

3. **Add to shell profile**
   ```bash
   # ~/.bashrc or ~/.zshrc
   export CASSANDRA_PASSWORD="secret123"
   source ~/.bashrc
   ```

## Query Issues

### Filter syntax error

**Error message**:
```
Error: invalid filter syntax: unexpected token 'SELECT'
```

**Solutions**:

1. **Use WHERE clause only** (no SELECT, FROM, etc.)
   ```cql
   # ✓ Correct
   id = '550e8400-e29b-41d4-a716-446655440000'
   
   # ✗ Wrong
   SELECT * FROM users WHERE id = '...'
   ```

2. **Check operator support**
   - Supported: `=`, `>`, `<`, `>=`, `<=`, `IN`, `CONTAINS`
   - Quote strings: `status = 'active'`
   - Use correct syntax for IN: `status IN ('active', 'pending')`

3. **Test query in cqlsh first**
   ```bash
   cqlsh> SELECT * FROM keyspace.table WHERE your_filter;
   ```

### Query returns no results

**Issue**: Filter seems correct but returns empty result set.

**Debugging steps**:

1. **Remove filter and check raw data**
   - Press `Esc` to clear filter
   - Verify data exists in table

2. **Check data types**
   ```cql
   # Wrong: id is UUID, not string
   id = 'abc123'
   
   # Correct: proper UUID format
   id = '550e8400-e29b-41d4-a716-446655440000'
   ```

3. **Verify partition key**
   - Filters must include partition key for best performance
   - Check table schema to identify partition key

4. **Check consistency level**
   - Try `consistency: "ALL"` in config for debugging

### Query timeout

**Error message**:
```
Error: query timeout after 10000ms
```

**Solutions**:

1. **Use more specific filter**
   ```cql
   # Include partition key
   user_id = 123 AND created_at > '2024-01-01'
   ```

2. **Increase timeout**
   ```json
   {
     "timeout_ms": 30000
   }
   ```

3. **Reduce page size**
   ```json
   {
     "defaults": {
       "page_size": 50
     }
   }
   ```

4. **Check database performance**
   - High CPU/memory usage
   - Compaction running
   - Large partition warnings

## Permission Errors

### Insufficient permissions

**Error message**:
```
Error: Unauthorized: User does not have permission
```

**Solutions**:

1. **Grant required permissions**
   ```cql
   # As superuser
   GRANT SELECT ON KEYSPACE app_data TO kassie_user;
   ```

2. **Use correct user in config**
   ```json
   {
     "auth": {
       "username": "read_only_user",
       "password": "password"
     }
   }
   ```

3. **Verify user permissions**
   ```cql
   LIST ALL PERMISSIONS OF kassie_user;
   ```

4. **Use superuser for full access** (development only)
   ```json
   {
     "auth": {
       "username": "cassandra",
       "password": "cassandra"
     }
   }
   ```

## Performance Issues

### Slow data loading

**Symptoms**: Data takes >5 seconds to load.

**Solutions**:

1. **Reduce page size**
   ```json
   {
     "defaults": {
       "page_size": 25
     }
   }
   ```

2. **Apply filters to reduce dataset**
   ```cql
   partition_key = '...' AND clustering_key > '...'
   ```

3. **Check network latency**
   ```bash
   ping database-host
   traceroute database-host
   ```

4. **Monitor database performance**
   - Check nodetool metrics
   - Review database logs
   - Check for large partitions

### High memory usage

**Symptoms**: Kassie using >500MB RAM.

**Solutions**:

1. **Reduce page size** (less data in memory)
   ```json
   {
     "defaults": {
       "page_size": 50
     }
   }
   ```

2. **Close inspector when not needed** (TUI)
   - Press `Esc` to close inspector

3. **Restart Kassie periodically**
   - Long-running sessions may accumulate state

### TUI is laggy

**Solutions**:

1. **Check terminal emulator performance**
   - Try different terminal (e.g., Alacritty, kitty)
   - Disable terminal transparency
   - Reduce font size

2. **Reduce data displayed**
   - Smaller page size
   - Apply filters
   - Close inspector

3. **Disable features**
   ```json
   {
     "clients": {
       "tui": {
         "vim_mode": false
       }
     }
   }
   ```

## TUI-Specific Issues

### Characters look broken

**Symptoms**: Boxes, question marks, or garbled text.

**Solutions**:

1. **Set proper locale**
   ```bash
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   kassie tui
   ```

2. **Use UTF-8 terminal**
   - Check terminal settings for UTF-8 support
   - Most modern terminals support UTF-8 by default

3. **Install Unicode fonts**
   - Recommended: Nerd Fonts, Fira Code, JetBrains Mono

### Colors are wrong

**Solutions**:

1. **Set TERM variable**
   ```bash
   export TERM=xterm-256color
   kassie tui
   ```

2. **Try different theme**
   ```json
   {
     "clients": {
       "tui": {
         "theme": "default"
       }
     }
   }
   ```

3. **Check terminal color support**
   ```bash
   # Test 256 colors
   for i in {0..255}; do printf "\x1b[38;5;${i}m%03d " "$i"; done; echo
   ```

### Keyboard shortcuts not working

**Solutions**:

1. **Check terminal key bindings**
   - Some terminals intercept certain keys
   - Try alternative keys (e.g., `Ctrl+N` instead of `n`)

2. **Disable conflicting shell bindings**
   ```bash
   # Temporarily disable fish vi mode
   set -U fish_key_bindings fish_default_key_bindings
   ```

3. **Use mouse as fallback** (if supported)

## Web UI Issues

### Page won't load

**Solutions**:

1. **Check server is running**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Try different browser**
   - Clear cache and cookies
   - Try incognito/private mode

3. **Check browser console**
   - Press `F12` → Console tab
   - Look for JavaScript errors

4. **Check port availability**
   ```bash
   lsof -i :8080
   # or
   netstat -tlnp | grep 8080
   ```

### WebSocket connection failed

**Error**: gRPC-Web connection error in browser console.

**Solutions**:

1. **Ensure server supports HTTP/2**
   - Kassie server supports HTTP/2 by default
   - Check for proxies that might downgrade

2. **Check for CORS issues**
   - Look for CORS errors in browser console
   - Kassie enables CORS by default in development

3. **Try different port**
   ```bash
   kassie web --port 3000
   ```

### Slow performance

**Solutions**:

1. **Enable hardware acceleration** in browser settings

2. **Reduce data displayed**
   - Smaller page size
   - Apply filters

3. **Close unnecessary tabs**

4. **Update browser** to latest version

## Debug Mode

Enable debug logging to troubleshoot issues:

```bash
kassie tui --log-level debug
```

**Log locations**:
- TUI: stderr (visible in terminal)
- Web/Server: stdout
- Docker: `docker logs <container>`

**Debug output includes**:
- Connection attempts
- Query execution
- Response times
- Error stack traces

## Getting Help

If you can't resolve your issue:

1. **Check existing issues**
   - [GitHub Issues](https://github.com/KashifKhn/kassie/issues)
   - Search for your error message

2. **Open a new issue**
   - Include Kassie version: `kassie version`
   - Include error messages
   - Include config (redact sensitive data)
   - Include debug logs
   - Describe steps to reproduce

3. **Community support**
   - GitHub Discussions
   - Stack Overflow (tag: kassie)

## Common Error Codes

| Code | Description | Common Cause |
|------|-------------|--------------|
| `AUTH_REQUIRED` | No token provided | Not logged in |
| `AUTH_INVALID` | Invalid token | Token expired or corrupted |
| `PROFILE_NOT_FOUND` | Profile missing | Wrong profile name |
| `CONNECTION_FAILED` | Can't connect | Database down or wrong host |
| `QUERY_ERROR` | CQL error | Invalid filter syntax |
| `INVALID_FILTER` | Filter syntax error | Wrong WHERE clause |
| `CURSOR_EXPIRED` | Pagination error | Cursor timeout (refresh) |
| `INTERNAL` | Server error | Check server logs |

See [Error Codes Reference](/reference/error-codes) for complete list.

## Quick Fixes

| Problem | Quick Fix |
|---------|-----------|
| Can't connect | Check database is running |
| Auth failed | Verify username/password |
| Config not found | Create `~/.config/kassie/config.json` |
| Filter error | Use WHERE clause only |
| Slow loading | Reduce page size, add filters |
| Broken characters | `export LANG=en_US.UTF-8` |
| Wrong colors | `export TERM=xterm-256color` |
| Port in use | Use different port with `--port` |

## Next Steps

- [Configuration Guide](/guide/configuration) - Review config options
- [TUI Usage](/guide/tui-usage) - Learn TUI features
- [Development Setup](/development/setup) - Build from source for debugging
