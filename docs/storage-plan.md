# storage plan (home-server first, replication later)

this is the plan we agreed on.

## what we care about

- home-server first
- you own your data
- local disk is the default and primary path
- no forced third-party storage
- storage api should be stable so we can add replication later
- plugins should keep using the same storage methods over time

## high-level direction

start with single-node local storage.
then add multi-node replication later.

important note:

- single-node mode can be close to stateless at higher layers, but storage itself is local state.
- replicated multi-node with local copies is a stateful replica model (eventual consistency), not pure stateless shared-storage mode.

that is fine and matches your goal: every machine keeps its own copy.

## core storage shape

`internal/core/storage/storage.go` should be the coordinator layer.
it should call small internal building blocks instead of doing everything inline.

### building blocks

1. `path_resolver.go`
- maps logical path to safe on-disk path
- blocks traversal (`..`), absolute escapes, and unsafe symlink behavior

2. `atomic_writer.go`
- handles safe streaming writes
- temp file -> stream copy -> optional fsync -> atomic rename
- cleanup temp files on error/cancel

3. `metadata_store.go`
- reads/writes file metadata
- first version can be sidecar json or sqlite-backed metadata
- keeps metadata logic separate from blob write logic

4. `storage.go`
- high-level methods: `put`, `get`, `stat`, `list`, `delete`, `move`, `copy`
- composes the three helpers above

## method behavior (v1)

- `put`
  - resolve path
  - write bytes through atomic writer
  - persist metadata (size, type, etag, times, extra)

- `get`
  - resolve path
  - stream with `os.Open` / `io.ReadCloser`
  - range support can be added after base path is stable

- `stat`
  - resolve path
  - merge filesystem stat + metadata record

- `delete`
  - delete blob + metadata record

- `move`
  - resolve src/dst
  - move blob safely
  - move/update metadata record

- `copy`
  - resolve src/dst
  - copy blob safely
  - copy/update metadata record

- `list`
  - list files by prefix
  - return metadata for each entry

## metadata and file type handling

core should treat file bytes as opaque.
core should not parse file internals.

store metadata like:

- namespace
- path
- size
- filename
- content_type
- etag
- created_at
- modified_at
- extra map

type handling:

- if caller sends content-type, store it
- if missing/generic, detect from initial bytes (`http.DetectContentType`) during stream path
- downstream plugin decides how to parse bytes

## concurrency model

correctness-critical concurrency goes inside storage layer.

inside storage:

- per-path lock for conflicting ops on same path
- optional global read/write semaphore limits
- atomic writer semantics always enforced

outside storage (higher layer):

- rate limiting
- request throttling
- background worker pools (later replication)

## naming and file layout

common go file naming is lowercase with underscores for multiword names.

recommended files:

- `internal/core/storage/storage.go`
- `internal/core/storage/path_resolver.go`
- `internal/core/storage/atomic_writer.go`
- `internal/core/storage/metadata_store.go`

possible future files:

- `internal/core/storage/versioning.go`
- `internal/core/storage/events.go`
- `internal/core/storage/replication_queue.go`

## roadmap

### phase 1: single-node local storage

- local disk blob storage
- metadata store (simple first)
- implement `put/get/stat/list/delete/move/copy`
- strong path safety + atomic writes

### phase 2: replication-ready foundation

- add version/event fields in metadata or change log
- emit change events after successful write ops
- keep api stable

### phase 3: multi-node replication

- nodes sync changes and missing blobs
- each node stores local copy
- define conflict policy (start with single-writer per namespace)
- add anti-entropy/reconciliation jobs

## consistency expectations

in replication mode:

- expect eventual consistency unless stronger coordination is added
- define conflict policy explicitly from day one
- do not hide conflicts; surface them clearly in metadata/events

## practical rule

build and harden single-node storage first.
make sure write semantics are correct and atomic.
then add replication on top of the same storage methods.

this keeps complexity under control and avoids rewrites.
