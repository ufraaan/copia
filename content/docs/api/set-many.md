---
title: Set many values
weight: 25
---

`SetMany(entries)` writes several key/value pairs together and returns an error if the store is closed or the batch cannot be written.

```go
entries := []copia.Entry{
    {Key: []byte("user:1"), Value: []byte("ada")},
    {Key: []byte("user:2"), Value: []byte("grace")},
}
if err := db.SetMany(entries); err != nil {
    return err
}
```

An empty batch does nothing. If a key appears more than once, its last value is stored. Copia copies the keys and values, so changing the input slices afterward does not change the store.

For a file-backed store, the batch is written as one checksummed WAL record and synced once. When reopening the file, Copia applies the batch only if the whole record is present and its checksum is valid; an incomplete final record is discarded. A batch record is limited to the WAL's 64 MiB maximum record size.
