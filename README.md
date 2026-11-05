<p align="center">
  <img src="assets/logo.webp" width="96" alt="copia logo">
</p>

<h2 align="center">copia</h2>

copia is a key/value store written in go. it can keep data in memory or record
changes in a write-ahead log.

## run

```sh
go run ./cmd/copia
```

use `-file` to keep data between runs:

```sh
go run ./cmd/copia -file copia.wal
```

commands:

- `set <key> <value>`
- `get <key>`
- `del <key>`
- `rename <source> <target>`
- `copy <source> <target>`
- `exists <key>`
- `keys [prefix]`
- `count`
- `clear`
- `help`
- `quit`

## package layout

- `cmd/copia` contains the command-line interface.
- `internal/copia` contains the key/value store and entry encoding.

## documentation site

The project documentation is a Hugo site in `content/`, configured in `hugo.toml`, and uses the Hugo Book theme.

## store usage

```go
var db copia.KV
if err := db.Open(); err != nil {
    return err
}
defer db.Close()

_, err := db.Set([]byte("key"), []byte("value"))
value, found, err := db.Get([]byte("key"))
```

`Set` and `Get` copy byte slices. Use `Open` for an in-memory store or
`OpenFile(path)` to replay and append to a write-ahead log. Operations before
opening or after closing return `copia.ErrNotOpen`.

## entry format

Entries contain two little-endian `uint32` lengths, followed by the key and
value bytes. `Entry.Decode` rejects records larger than 64 MiB. The wal stores
each operation with a checksum and syncs it before changing the in-memory data.
