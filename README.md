# scio

[![CI](https://github.com/zoobz-io/scio/actions/workflows/ci.yml/badge.svg)](https://github.com/zoobz-io/scio/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/zoobz-io/scio/branch/main/graph/badge.svg)](https://codecov.io/gh/zoobz-io/scio)
[![Go Report Card](https://goreportcard.com/badge/github.com/zoobz-io/scio)](https://goreportcard.com/report/github.com/zoobz-io/scio)
[![CodeQL](https://github.com/zoobz-io/scio/actions/workflows/codeql.yml/badge.svg)](https://github.com/zoobz-io/scio/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zoobz-io/scio.svg)](https://pkg.go.dev/github.com/zoobz-io/scio)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zoobz-io/scio)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/zoobz-io/scio)](https://github.com/zoobz-io/scio/releases)

URI-based data catalog with atomic operations for Go.

scio is the system's authoritative map of data sources — it knows where data lives and provides type-agnostic access via atoms.

## The Map

scio registers storage resources and routes operations through URIs:

```go
s := scio.New()

// Register resources (user code provides typed grub wrappers)
s.RegisterDatabase("db://users", usersDB.Atomic())
s.RegisterStore("kv://sessions", sessionsStore.Atomic())
s.RegisterBucket("bcs://documents", docsBucket.Atomic())

// Operations via URI — scio routes to the right provider
atom, _ := s.Get(ctx, "db://users/123")
s.Set(ctx, "kv://sessions/abc", sessionAtom)

// Query databases
results, _ := s.Query(ctx, "db://users", stmt, params)

// Introspect the topology
s.Sources()                    // all registered resources
s.FindBySpec(spec)             // resources sharing a type
s.Related("db://users")        // other resources with same spec
```

## Install

```bash
go get github.com/zoobz-io/scio
```

Requires Go 1.24 or higher.

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/zoobz-io/grub"
    "github.com/zoobz-io/scio"
)

func main() {
    // Create scio instance
    s := scio.New()

    // Register a database (assumes usersDB is a grub.Database[User])
    err := s.RegisterDatabase("db://users", usersDB.Atomic(),
        scio.WithDescription("User accounts"),
        scio.WithTag("owner", "auth-team"),
    )
    if err != nil {
        panic(err)
    }

    ctx := context.Background()

    // Get a record
    atom, err := s.Get(ctx, "db://users/123")
    if err != nil {
        panic(err)
    }

    // Access fields via typed maps
    fmt.Println(atom.Strings["email"])
    fmt.Println(atom.Ints["age"])

    // Find related resources
    for _, r := range s.Related("db://users") {
        fmt.Printf("Related: %s (%s)\n", r.URI, r.Variant)
    }
}
```

## URI Semantics

| Variant | Pattern | Example |
|---------|---------|---------|
| `db://` | `table/key` | `db://users/123` |
| `kv://` | `store/key` | `kv://sessions/abc` |
| `bcs://` | `bucket/path` | `bcs://docs/reports/q4.pdf` |

## Why scio?

- **URI-based routing** — logical addressing decoupled from physical storage
- **Type-agnostic** — operates on atoms, never needs to know your types
- **Topology awareness** — knows what resources exist and their relationships
- **Spec tracking** — auto-detects when multiple resources share the same type
- **Metadata support** — annotate resources for system use (LLM context, ownership, versioning)

## Documentation

### Learn
- [Quick Start](docs/2.learn/1.quickstart.md)
- [Core Concepts](docs/2.learn/2.concepts.md)
- [Architecture](docs/2.learn/3.architecture.md)

### Reference
- [API Reference](docs/5.reference/1.api.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
