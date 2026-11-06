---
title: Clear
weight: 100
---

`Clear()` removes every key. In file-backed mode, it appends and syncs a clear record before resetting the in-memory map.

```go
if err := db.Clear(); err != nil { return err }
```
