# storage plan (home-server first, replication later)

this document tracks the current storage architecture for plinth core.

## goals

- home-server first
- local disk as primary blob storage
- stream-first APIs for very large files
- resumable uploads/downloads
- stable interfaces so replication can be added later
- official metadata backend: sqlite

## core architecture

`internal/core/storage/storage.go` is the coordinator.
it composes focused building blocks:

1. `path_resolver.go`
- maps logical keys to safe filesystem paths
- blocks traversal and root escape
- resolves upload temp paths

2. `atomic_writer.go`
- stream write primitives for object payload bytes
- atomic single-shot write: temp -> copy -> sync -> rename
- resumable primitives: append-at-offset, commit temp, abort temp

3. `metadata_store.go`
- `ObjectStore` interface for committed object records
- `UploadStore` interface for resumable upload session records
- both are backend-agnostic contracts

4. `metadata/sqlite/*`
- official backend implementation
- one `Store` with one shared `*sql.DB`
- implements both `ObjectStore` and `UploadStore`
- sqlite runtime config centralized in `metadata/sqlite/store.go`

## high-level storage API (v1 foundation)

object operations:
- `Put`
- `Open` (range-aware)
- `Stat`
- `List`
- `Delete`
- `Move`
- `Copy`

resumable upload operations:
- `CreateUpload`
- `AppendUpload`
- `StatUpload`
- `CommitUpload`
- `AbortUpload`

## resumable model

- upload session metadata is stored in `UploadStore`
- upload bytes are written to temp files on local disk
- `AppendUpload` enforces expected offset (server is source of truth)
- `CommitUpload` promotes temp file atomically and finalizes object record
- `AbortUpload` marks session and removes temp data

## read model

- `OpenOptions` supports `Offset` and `Length`
- use `io.NewSectionReader` for bounded ranged reads
- supports resume-friendly large downloads

## metadata model

`ObjectInfo` fields (current contract):

- key (namespace + path)
- size
- etag
- content_type
- filename
- version
- deleted (for future tombstones)
- created_at / modified_at
- attributes map

## sqlite (official backend)

recommended startup configuration:

- `PRAGMA journal_mode = WAL`
- `PRAGMA synchronous = NORMAL`
- `PRAGMA foreign_keys = ON`
- `PRAGMA busy_timeout = 5000`

notes:
- use a constrained sqlite pool (`max open conns = 1`)
- metadata writes are small; payload IO remains filesystem streaming

## sql tooling

- `goose` for schema migrations
- `sqlc` for typed query generation from SQL files
- keep object queries and upload queries separate, backed by same db

## implementation order

phase 1 (single-node correctness):
1. path resolver safety
2. file writer atomic + resumable primitives
3. sqlite schema/migrations (`objects`, `uploads`)
4. sqlc query generation and store method wiring
5. engine method wiring (`Put/Open/...` + resumable APIs)

phase 2 (hardening):
1. context cancellation + cleanup guarantees
2. offset/idempotency validation
3. fsync policy and error mapping
4. crash-recovery behavior for active uploads

phase 3 (replication-ready):
1. add change-log/event table
2. emit events after successful commits/deletes
3. add anti-entropy worker later

## practical rule

correctness first:
- never buffer full files in memory
- do not trust client offsets without verification
- ensure atomic commit semantics for final object writes
