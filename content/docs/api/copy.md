---
title: Copy
weight: 60
---

`Copy(sourceKey, targetKey)` copies a value to another key. It returns whether the source key existed and an error.

```go
copied, err := db.Copy([]byte("source"), []byte("backup"))
```
