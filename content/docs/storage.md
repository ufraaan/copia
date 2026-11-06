---
title: Storage and WAL format
description: How Copia appends, verifies, and replays persistent operations.
weight: 40
---

`OpenFile(path)` opens or creates a WAL, replays its valid records into memory, then positions the file for appending. Each successful mutation is recorded and synced before the in-memory map is changed.

## Record framing

Each record has a 12-byte header followed by a payload:

| Header bytes | Meaning |
| --- | --- |
| 0–3 | `CPWL` magic bytes |
| 4–7 | Payload length as a little-endian `uint32` |
| 8–11 | IEEE CRC-32 checksum of the payload |

The payload begins with an operation byte. For `set`, `del`, `rename`, and `copy`, the remaining bytes encode an entry: two little-endian `uint32` lengths, then key bytes and value bytes. `clear` has only its operation byte. Entries are limited to 64 MiB.

## Replay and errors

Replay applies records in order. A final incomplete header or payload is treated as a truncated tail and removed before appending resumes. Invalid magic, an invalid size, a checksum mismatch, or an invalid complete operation returns an error instead of silently replaying past the corruption.
