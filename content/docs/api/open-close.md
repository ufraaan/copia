---
title: Open and close
weight: 10
---

Use `Open()` for a fresh in-memory store. Use `OpenFile(path)` to replay existing WAL records and append new operations to that file.

```go
var db copia.KV
if err := db.OpenFile("copia.wal"); err != nil { return err }
defer db.Close()
```

`Close()` syncs and closes the WAL. Methods called before opening or after closing return `copia.ErrNotOpen`.
