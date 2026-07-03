# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability in MockSmith, please report it responsibly by emailing the maintainer directly. Do NOT open a public GitHub issue.

## Security Considerations

### Configuration Files
- Config files may contain sensitive mock data (fake API keys, tokens, etc.)
- MockSmith does NOT send any data externally — all processing is local
- Config files are not validated for sensitive content (by design — they're mock data)

### Network Exposure
- MockSmith binds to `0.0.0.0` by default — ensure firewall rules are appropriate
- Use `host: "127.0.0.1"` for local-only access
- HTTPS is available via TLS config for production-like testing

### Request Logging
- By default, request logs include the full request path and headers
- Use `--json-log` for structured logging that can be parsed by log aggregators
- Be cautious with sensitive data in mock response bodies

### File Watching
- Hot-reload watches the config file for changes (polling-based)
- No file system access beyond the specified config file
- Config validation runs on every reload

## Best Practices

1. **Never use real credentials** in mock response bodies
2. **Use environment-specific configs** for dev/staging/production
3. **Restrict network access** with firewall rules or bind to localhost
4. **Review config files** before committing to version control
5. **Use HTTPS** for sensitive API mocking
