---
title: Get
weight: 30
---

`Get(key)` returns a copy of the value, whether the key exists, and an error.

```go
value, found, err := db.Get([]byte("language"))
```

The returned bytes do not alias the store's internal value.
