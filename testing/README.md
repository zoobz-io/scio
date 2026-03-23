# Testing

This directory contains test utilities and infrastructure for scio.

## Structure

```
testing/
├── helpers.go          # Test utilities and mock implementations
├── helpers_test.go     # Tests for helpers
├── integration/        # Integration tests
├── benchmarks/         # Performance benchmarks
└── README.md
```

## Test Utilities

### Assertion Helpers

```go
import st "github.com/zoobz-io/scio/testing"

func TestSomething(t *testing.T) {
    ctx := st.WithTimeout(t, 5*time.Second)

    st.AssertNoError(t, err)
    st.AssertEqual(t, got, want)
    st.AssertTrue(t, condition)
}
```

### Mock Providers

```go
spec := st.TestSpec("User", "ID", "Email", "Name")

db := st.NewMockDatabase("users", spec)
store := st.NewMockStore(spec)
bucket := st.NewMockBucket(spec)
```

## Running Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# Benchmarks
make test-bench
```
