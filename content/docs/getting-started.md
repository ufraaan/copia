---
title: Getting started
description: Run the Copia command line and save data between runs.
weight: 10
---

## Requirements

Install the Go version declared in [`go.mod`](https://github.com/ufraaan/copia/blob/main/go.mod).

## Run Copia

From the repository root, start an in-memory session:

```sh
go run ./cmd/copia
```

To append changes to a WAL file and replay them next time, pass `-file`:

```sh
go run ./cmd/copia -file copia.wal
```

The interactive prompt accepts one command per line. For example:

```text
set language Go
get language
Go
```

Use `quit` (or `exit`) to leave the session. An in-memory store starts empty on each run; a file-backed store loads its recorded operations when opened.
