# Plinth

A modular, headless file hosting platform. Core handles storage and plugin orchestration, then you extend with plugins.

## What is Plinth?

Plinth is a suite of client-server software for creating file hosting services. The idea is simple: keep the core minimal and let plugins do the heavy lifting.

Want a web UI? That's a plugin. Need encryption? Also a plugin. For compression, thumbnails, search you just add the plugin. You pick what you need, nothing more, nothing less.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      CORE (headless)                            │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────┐    │
│  │   Storage   │  │   Plugin    │  │      gRPC API        │    │
│  │   Engine    │  │   Registry  │  │   Storage ops        │    │
│  │             │  │   + Loader  │  │   Plugin register    │    │
│  └─────────────┘  └─────────────┘  └──────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ Unix Sockets
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌─────────┐           ┌─────────┐           ┌─────────┐
   │   web   │           │ gallery │           │ encrypt │
   └─────────┘           └─────────┘           └─────────┘
```

Core is the brain. Plugins are the limbs.

## How It Works

### Communication

Plugins are separate binaries that talk to core via gRPC over Unix sockets.

This means plugins can be written in any language that speaks gRPC, whether that's Go, Python, Rust, or something else entirely. It also means a plugin crash won't take down core, and there are no shared memory headaches to deal with.

### Data Flow

Plugins don't talk to each other directly. All data goes through core's Storage API.

```
Plugin A                    Core                    Plugin B
    │                         │                         │
    ├── Store("/a/file") ────►│                         │
    │                         │◄── Get("/a/file") ──────┤
    │                         ├────── returns data ────►│
```

This keeps plugins decoupled. Plugin B doesn't need to know Plugin A exists.

### Plugin Configuration

Users define their own names for plugins and map them to binaries:

```yaml
plugins:
  compress: /path/to/gzip-plugin
  encrypt: /path/to/age-plugin
  store: /path/to/s3-plugin
```

This means you control the naming. Want short names? Use them. Prefer verbose? Go for it. If you ever want to swap gzip for zstd, just change the path. Pipelines stay the same.

### Pipelines

Pipelines reference plugins by the names you defined:

```yaml
pipelines:
  upload:
    - compress
    - encrypt
    - store

  download:
    - decrypt
    - decompress
```

You can also build pipelines in code using a string-based builder:

```
Pipeline().Use("compress").Use("encrypt").Use("store").Run(file)
```

Or reference predefined pipelines by name:

```
Process(file, "upload")
```

### Startup Health Check

When core starts, it runs a health check against all configured plugins. It tries to connect to each one and reports the results.

```
[OK]   compress    /path/to/gzip-plugin
[OK]   encrypt     /path/to/age-plugin
[FAIL] store       /path/to/s3-plugin — connection refused
```

If a plugin fails, core doesn't crash. It continues with what works and logs the failures. Pipelines that depend on a failed plugin will error when called, but the rest of the system stays up.

### Frontend

Each plugin owns its frontend while core aggregates navigation from registered plugins.

Templating is done via Go templates with HTMX, and PWA is supported. For styling, you can use the shared SDK or roll your own. The web UI itself is a plugin, so if you don't need a UI, just don't load it.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go (pure, no CGo) |
| Plugin system | HashiCorp go-plugin |
| IPC | gRPC over Unix sockets |
| Frontend | Go templates + HTMX |
| Remote access | VPN / ZTNA (e.g., Twingate) |

## Project Structure

```
plinth/
├── cmd/
│   └── plinth/
├── internal/
│   └── core/
│       ├── storage/
│       ├── registry/
│       └── orchestrator/
│   └── server/
├── proto/
├── sdk/
│   ├── go/
│   ├── proto/
│   └── web/
└── plugin/
    ├── web/
    ├── gallery/
    ├── compress/
    └── encrypt/
```

### Directory Responsibilities

| Directory | What it does |
|-----------|--------------|
| `cmd/plinth/` | Main binary entry point |
| `internal/` | Private packages that Go enforces, so external projects can't import them |
| `internal/core/storage/` | File storage engine |
| `internal/core/registry/` | Plugin registration and discovery |
| `internal/core/orchestrator/` | Pipeline execution |
| `internal/server/` | gRPC server wiring |
| `proto/` | Protocol buffer definitions and generated code |
| `sdk/` | Public SDK for plugin authors |
| `sdk/go/` | Go-specific helpers for plugin development |
| `sdk/proto/` | Proto files for generating stubs in any language |
| `sdk/web/` | Shared templates and CSS, totally optional |
| `plugin/` | Official plugins where each subdirectory builds to its own binary |

## Plugin Types

| Type | Purpose | Example |
|------|---------|---------|
| `endpoint` | Serves a feature with its own routes and UI | gallery, documents |
| `middleware` | Processes data in pipelines | compress, encrypt |

Endpoint plugins register a path and appear in navigation. Middleware plugins are called as part of other plugins' pipelines.

## Writing Plugins

In Go, you import the SDK, implement the interface, and serve via go-plugin.

In other languages, you grab the `.proto` files from `sdk/proto/`, generate stubs with `protoc`, and implement the gRPC service. Core doesn't care what language you use.

## Configuration

Plinth follows the XDG Base Directory Specification for config file locations.

| Type | Default Location |
|------|------------------|
| Config | `~/.config/plinth/` |
| Data | `~/.local/share/plinth/` |
| Cache | `~/.cache/plinth/` |

System-wide config can live in `/etc/plinth/`. You can also pass `--config` to override everything.

Precedence: flag → environment variable → user config → system config → defaults.

## Plugin Lifecycle

Core starts plugins and keeps them running. If a plugin binary is updated, core hot-reloads it without restarting itself. Core is the heart of the system and should stay as stable as possible.

## Error Handling

When a pipeline step fails, core skips it and continues with the next step. The health check reports failures with error messages when available. This keeps the system running even when individual plugins have issues.

## Storage Backends

The default storage engine uses the local filesystem. However, the storage layer is pluggable, so you could swap it with an S3-compatible backend if needed.

## Authentication

Authentication and authorization are required but the specific approach is not yet decided. The goal is a permission system that's easy to maintain.

## Design Principles

**Core stays dumb.** It stores files and runs pipelines. That's it.

**Plugins own everything else.** UI, features, processing logic.

**Data flows through core.** Plugins never talk directly to each other.

**Crash isolation.** One plugin dying doesn't kill the system.

**Language agnostic.** gRPC means any language can be used for a plugin.

**Compose your deployment.** Only load what you need.

## License

Plinth is licensed under the [PolyForm Noncommercial License 1.0.0](LICENSE). Free to use, modify, and share for noncommercial purposes. Commercial use requires a separate agreement.
