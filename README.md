# copia

copia is an in-memory key/value store written in go. it includes a command-line
interface and a binary format for encoding entries. data is lost when the
command exits.

## run

```sh
go run ./cmd/copia
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

`Set` and `Get` copy byte slices. Operations before `Open` or after `Close`
return `copia.ErrNotOpen`.

## entry format

Entries contain two little-endian `uint32` lengths, followed by the key and
value bytes. `Entry.Decode` rejects records larger than 64 MiB.
