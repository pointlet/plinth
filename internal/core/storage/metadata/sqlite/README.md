# SQLite Metadata Backend (goose + sqlc)

This package uses:

- `goose` for schema migrations (`migrations/`)
- `sqlc` for typed query code generation (`queries/` -> `gen/`)

## Central flow (local + CI)

Use repository make targets so local and CI run the same commands:

```bash
make ci
```

This runs:

1. `sqlc generate`
2. drift check (`git status --porcelain`)
3. `go test ./...`

## Run migrations

```bash
make db-up DB=/path/to/plinth.db
```

## Generate query code

```bash
make sqlc-generate
```

`store.go` owns DB open/config.
`object_store.go` and `upload_store.go` should call generated `sqlcgen` methods.
