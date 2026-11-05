---
title: Command line reference
description: Commands supported by the interactive Copia CLI.
weight: 20
---

Start the program with `go run ./cmd/copia`. Add `-file path` to persist operations to a WAL file.

| Command | Behavior |
| --- | --- |
| `set <key> <value>` | Store a value. The value is the remainder of the input line. |
| `get <key>` | Print the value, or `not found`. |
| `del <key>` | Delete a key, or print `not found`. |
| `rename <source> <target>` | Move a value to another key. |
| `copy <source> <target>` | Copy a value to another key. |
| `exists <key>` | Print `true` or `false`. |
| `keys [prefix]` | Print keys with the prefix, in bytewise sorted order. |
| `count` | Print the number of keys. |
| `clear` | Remove all keys. |
| `help` | Print the command list. |
| `quit` / `exit` | End the session. |

Keys and values in the CLI are whitespace-delimited text. The Go API accepts arbitrary byte slices.
