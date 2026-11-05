---
title: Delete
weight: 40
---

`Del(key)` deletes the key. It returns whether a key was deleted and an error. A missing key returns `false` without an error.

```go
deleted, err := db.Del([]byte("language"))
```
