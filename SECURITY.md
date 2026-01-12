# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow these steps:

1. **DO NOT** create a public GitHub issue
2. Email security details to the maintainers
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

## Security Best Practices

When using Scio:

1. **URI Validation**: Scio parses URIs to route operations. Ensure URIs are properly validated before passing to Scio to prevent injection attacks.

2. **Spec Exposure**: Be aware that type metadata (field names, types) is accessible via the catalog. Review your struct definitions for information disclosure.

3. **Provider Security**: Scio delegates storage operations to grub providers. Ensure your underlying storage (databases, caches, blob stores) are properly secured.

4. **Access Control**: Scio does not implement authentication or authorization. Implement appropriate access controls in your application layer.

## Security Features

Scio is designed with security in mind:

- No direct network operations (delegated to providers)
- No file system operations (delegated to providers)
- Type-safe operations through atoms
- URI scheme validation prevents variant injection

## Acknowledgments

We appreciate responsible disclosure of security vulnerabilities.
