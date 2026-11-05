---
title: Copia
description: A small key/value store written in Go.
---

Copia keeps byte keys and values in memory, with an optional write-ahead log for persistence.

> **Early project:** Copia is a small learning project. Review its guarantees and limitations before using it for important data.

## Start here

- [Install and run]({{< relref "docs/getting-started" >}}) the interactive command line.
- Learn the [Go API]({{< relref "docs/api" >}}) for using the store in a Go program.
- Read how [persistence works]({{< relref "docs/storage" >}}) and what the WAL records contain.

## What Copia provides

- In-memory storage or append-only persistence through a write-ahead log.
- Byte-slice keys and values, copied on write and read.
- Basic operations: set, get, delete, rename, copy, prefix key listing, count, and clear.
- A checksum on each WAL record and replay when opening a file.
