---
title: Rename
weight: 50
---

`Rename(oldKey, newKey)` moves a value to another key. It returns whether the source key existed and an error.

```go
renamed, err := db.Rename([]byte("old"), []byte("new"))
```
