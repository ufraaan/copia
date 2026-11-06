---
title: Exists
weight: 70
---

`Exists(key)` reports whether the key exists and returns an error.

```go
found, err := db.Exists([]byte("language"))
```
