---
title: List keys
weight: 80
---

`Keys(prefix)` returns keys that begin with the prefix, sorted in bytewise order.

```go
keys, err := db.Keys([]byte("user:"))
```
