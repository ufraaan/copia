---
title: Go API
description: Open and use Copia from Go code.
weight: 30
bookCollapseSection: true
---

The store implementation is in the `internal/copia` package. Go's `internal` import rule means only code within this module tree can import it; it is not currently a public library API for other modules.

Call `Open()` for a fresh in-memory store or `OpenFile(path)` to replay and append to a WAL. Always close an opened store; `Close()` syncs and closes its WAL.

Choose an operation for details:

- [Open and close](open-close/)
- [Set](set/)
- [Set many values](set-many/)
- [Get](get/)
- [Delete](delete/)
- [Rename](rename/)
- [Copy](copy/)
- [Exists](exists/)
- [List keys](keys/)
- [Count entries](count/)
- [Clear](clear/)

All methods return `ErrNotOpen` if called before opening the store or after closing it. Methods that return byte slices provide copies, so callers can change their slices without mutating stored data.
