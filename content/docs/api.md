---
title: Go API
description: Open and use Copia from Go code.
weight: 30
---

The store implementation is in the `internal/copia` package. Go's `internal` import rule means only code within this module tree can import it; it is not currently a public library API for other modules.

```go
package main

import (
    "log"

    "github.com/ufraaan/copia/internal/copia"
)

func main() {
    var db copia.KV
    if err := db.OpenFile("copia.wal"); err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if _, err := db.Set([]byte("language"), []byte("Go")); err != nil {
        log.Fatal(err)
    }
    value, found, err := db.Get([]byte("language"))
    if err != nil {
        log.Fatal(err)
    }
    if found {
        log.Println(string(value))
    }
}
```

Call `Open()` for a fresh in-memory store or `OpenFile(path)` to replay and append to a WAL. Always close an opened store; `Close()` syncs and closes its WAL.

## Operations

- `Set(key, value)` writes or replaces a value.
- `Get(key)` returns a copied value and whether the key exists.
- `Del(key)`, `Rename(oldKey, newKey)`, and `Copy(sourceKey, targetKey)` report whether a key was changed or found.
- `Exists(key)` checks for a key; `Keys(prefix)` returns matching keys sorted by bytes.
- `Len()` returns the entry count; `Clear()` removes all entries.

All methods return `ErrNotOpen` if called before opening the store or after closing it. Methods that return byte slices provide copies, so callers can change their slices without mutating stored data.
