---
title: Set
weight: 20
---

`Set(key, value)` writes or replaces a value. It returns whether it updated the store and an error.

```go
updated, err := db.Set([]byte("language"), []byte("Go"))
```

The store copies both byte slices. In file-backed mode, the WAL record is synced before the in-memory value changes.
