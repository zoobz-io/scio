# Integration Tests

Integration tests verify scio's behavior with real storage providers.

## Running Integration Tests

```bash
make test-integration
```

## Test Structure

Integration tests are organized by provider:

```
integration/
├── database/    # SQL database tests
├── store/       # Key-value store tests
└── bucket/      # Blob storage tests
```

## Requirements

Integration tests may require running services. See individual test files for setup instructions.
