# Plinth

A modular, headless file hosting platform. Core handles storage and plugin orchestration, then you extend with plugins.

## What is Plinth?

Plinth is a suite of client-server software for creating file hosting services. The idea is simple: keep the core minimal and let plugins do the heavy lifting.

Want a web UI? That's a plugin. Need encryption? Also a plugin. For compression, thumbnails, search you just add the plugin. You pick what you need, nothing more, nothing less.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      CORE (headless)                            │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────┐     │
│  │   Storage   │  │   Plugin    │  │    HTTP Reverse      │     │  
│  │   Engine    │  │   Registry  │  │    Proxy + gRPC API  │     │
│  │             │  │   + Loader  │  │                      │     │
│  └─────────────┘  └─────────────┘  └──────────────────────┘     │
│  ┌──────────────────────────────┐                               │
│  │   Pipeline Engine            │                               │
│  │   (in-process middleware)    │                               │
│  └──────────────────────────────┘                               │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ gRPC over Unix Sockets
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
   ┌──────────┐          ┌──────────┐          ┌──────────┐
   │   web    │          │ gallery  │          │  hello   │
   │(endpoint)│          │(endpoint)│          │(endpoint)│
   └──────────┘          └──────────┘          └──────────┘
```

Core is the brain. Endpoint plugins are the limbs. Middleware plugins live inside core as in-process transforms.

## Two-Tier Plugin System

Plinth uses two fundamentally different plugin types with different execution models:

### Endpoint Plugins (out-of-process)

Endpoint plugins are separate binaries that communicate with core via gRPC over Unix sockets using HashiCorp go-plugin. They serve HTTP routes, render UI, and handle user interaction.

This is where crash isolation matters. A buggy third-party endpoint plugin cannot take down core. If an endpoint plugin crashes, core detects it, logs the failure, and restarts it with exponential backoff. The rest of the system stays up.

Endpoint plugins can be written in any language that implements the gRPC service contract and the go-plugin handshake protocol.

### Middleware Plugins (in-process)

Middleware plugins are Go packages compiled into the core binary. They implement a simple `io.Reader`/`io.Writer` streaming interface and transform data as it flows through pipelines.

Middleware sits in the hot data path. Serializing megabytes of file data through gRPC for every compress or encrypt call would be wasteful. In-process execution gives zero-copy streaming with native error handling.

Third-party middleware authors publish Go packages implementing the `sdk.Middleware` interface. Users build a custom core binary that imports the packages they need. This follows the same pattern used by Caddy and Hugo.

**Official middleware** (compress, encrypt) ships with the default build.

**Third-party middleware** is included by specifying filesystem paths in a build configuration, then compiling a custom binary.

| Type | Execution | Crash Isolation | Language | Example |
|------|-----------|-----------------|----------|---------|
| `endpoint` | Out-of-process (gRPC) | Yes | Any (Go, Python, Rust) | web, gallery |
| `middleware` | In-process (Go interface) | No (recover on panic) | Go only | compress, encrypt |

## How It Works

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

### Pipelines

Pipelines chain middleware plugins together to process files on upload or download. They are defined in the YAML config:

```yaml
pipelines:
  upload:
    steps:
      - compress
      - encrypt
    on_error: fail

  download:
    steps:
      - decrypt
      - decompress
    on_error: fail
```

Each step receives an `io.Reader` from the previous step and produces an `io.Reader` for the next. Core pipes the output of each middleware directly into the input of the next with zero intermediate copies.

### Pipeline Data Contract

Data flows between pipeline steps as a byte stream with metadata:

```go
type FileContext struct {
    ContentType string            // e.g., "image/png", "video/mp4"
    Filename    string            // original filename
    Size        int64             // total size if known, -1 if streaming
    Extra       map[string]string // opaque metadata between steps
}
```

The `Extra` field carries information between steps without core needing to understand it:

- A compress plugin sets `extra["compression"] = "zstd"` so decompress knows the algorithm.
- An encrypt plugin sets `extra["encrypted"] = "true"` and `extra["key_id"] = "user-key-1"`.

Core does not parse `Extra`. It only reads `ContentType` for pipeline compatibility validation at startup.

### Error Handling

When a pipeline step fails or a middleware plugin panics, the pipeline operation fails and core reports the error to the caller. Steps are sequential dependencies — if compress fails, the file is not stored half-processed.

Core itself does not crash. It logs the error and continues serving other requests. The failure is scoped to that single file operation.

### Startup Health Check

When core starts, it connects to each configured endpoint plugin and reports the results:

```
[OK]   web         /usr/lib/plinth/plugins/web
[OK]   gallery     /usr/lib/plinth/plugins/gallery
[FAIL] hello       /usr/lib/plinth/plugins/hello — connection refused
```

If an endpoint plugin fails to start, core continues with what works and logs the failure. Requests to that plugin's routes return 503 until it recovers.

### Crash Recovery

If an endpoint plugin process dies during operation, core detects it (go-plugin monitors the subprocess), logs the failure, and restarts the plugin with exponential backoff (1s, 2s, 4s, 8s, max 60s). After 5 consecutive failures, core marks the plugin as permanently failed and stops restarting.

Core exposes plugin health status via its registry API so endpoint plugins can display it.

## SDK

The SDK (`github.com/pointlet/plinth/sdk`) is the shared contract between core and all plugins. Both core and plugins import it.

```
core (go.mod) ──imports──► sdk (go.mod) ◄──imports── plugins (go.mod)
```

### Middleware Interface

```go
import "github.com/pointlet/plinth/sdk"

type Middleware interface {
    Name() string
    Process(ctx context.Context, fc *FileContext, r io.Reader) (io.Reader, *FileContext, error)
}
```

### Endpoint Interface

```go
type Endpoint interface {
    Name() string
    BasePath() string
    DisplayName() string
}
```

### Writing a Middleware Plugin (Go)

Implement the interface and register via `init()`:

```go
package zstd

import (
    "context"
    "io"

    "github.com/pointlet/plinth/sdk"
)

type compressor struct{}

func (c *compressor) Name() string { return "compress" }

func (c *compressor) Process(ctx context.Context, fc *sdk.FileContext, r io.Reader) (io.Reader, *sdk.FileContext, error) {
    // compress the stream, update metadata
    fc.Extra["compression"] = "zstd"
    return newZstdReader(r), fc, nil
}

func init() {
    sdk.RegisterMiddleware(&compressor{})
}
```

**Creating third-party middleware:**

1. Create a directory anywhere on your filesystem
2. Initialize a Go module: `go mod init github.com/you/plinth-watermark`
3. Add the SDK dependency: `go get github.com/pointlet/plinth/sdk`
4. Implement the `sdk.Middleware` interface and register via `init()`
5. Users add the path to their `build.yaml` and run `plinth build`

The middleware is compiled into their custom binary and available for use in pipelines.

### Writing an Endpoint Plugin (Go)

Import the SDK and serve with go-plugin:

```go
package main

import "github.com/pointlet/plinth/sdk"

type gallery struct{}

func (g *gallery) Name() string        { return "gallery" }
func (g *gallery) BasePath() string     { return "/gallery" }
func (g *gallery) DisplayName() string  { return "Gallery" }

func main() {
    sdk.Serve(&gallery{})
}
```

### Writing an Endpoint Plugin (Other Languages)

Non-Go plugins must implement the gRPC service contract and the go-plugin handshake protocol.

**1. Check the magic cookie:**

The plugin binary must verify `PLINTH_PLUGIN=1` is set in the environment. If missing, print an error and exit.

**2. Start a gRPC server:**

Generate stubs from the `.proto` files in `proto/` using `protoc`. Implement the `EndpointService` (Health, HandleHTTP, Routes).

**3. Write the handshake line to stdout:**

```
1|1|unix|<socket_path>|grpc|
```

This tells go-plugin where to connect. See `proto/README.md` for the full protocol specification.

## Custom Builds

Official middleware (compress, encrypt) is included in the default build. To add third-party middleware, specify local filesystem paths in a build configuration file:

```yaml
# build.yaml
middleware:
  # Third-party middleware (local filesystem paths)
  watermark: /home/user/my-plugins/watermark
  transcode: /opt/custom-plugins/transcode
```

Each middleware directory must be a valid Go module with a `go.mod` file:

```
/home/user/my-plugins/watermark/
├── go.mod          # module github.com/user/watermark
└── watermark.go    # implements sdk.Middleware, registers via init()
```

Build a custom binary that includes official middleware plus your additions:

```
plinth build --config build.yaml -o plinth-custom
```

### How It Works

Go imports require module paths, not filesystem paths. The build tool:

1. Reads the build config and finds each middleware path
2. Extracts the module path from each middleware's `go.mod`
3. Adds `replace` directives to map module paths to local paths
4. Generates import statements for each middleware
5. Compiles everything into a single binary

The resulting binary contains all official middleware plus your custom middleware, all running in-process with zero serialization overhead.

This is the same pattern used by [xcaddy](https://github.com/caddyserver/xcaddy) for Caddy plugins.

## Configuration

Runtime configuration is defined in a single YAML file. Pass the path with `--config`:

```
plinth --config /path/to/plinth.yaml
```

### Example Config

```yaml
storage:
  path: ~/.local/share/plinth/files

plugins:
  # Endpoint plugins (out-of-process binaries)
  web:
    binary: /usr/lib/plinth/plugins/web
  gallery:
    binary: /usr/lib/plinth/plugins/gallery

pipelines:
  upload:
    steps:
      - compress
      - encrypt
    on_error: fail

  download:
    steps:
      - decrypt
      - decompress
    on_error: fail

  # Pipeline using third-party middleware
  gallery-upload:
    steps:
      - compress
      - watermark
      - encrypt
    on_error: fail
```

Middleware referenced in pipelines must be compiled into the binary (via `build.yaml`). If a middleware name is not found, core fails at startup with an error.

Endpoint plugins are separate binaries specified by path. They run out-of-process and communicate via gRPC.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go (pure, no CGo) |
| Endpoint plugin system | HashiCorp go-plugin |
| Endpoint IPC | gRPC over Unix sockets |
| Middleware plugin system | In-process Go interfaces |

## Project Structure

The repository is a multi-module Go workspace. Core, SDK, and each endpoint plugin are separate Go modules, tied together by `go.work` for local development.

```
plinth/
├── go.mod               # core module (imports sdk, never imported externally)
├── go.work              # workspace for local development
├── cmd/
│   └── plinth/          # core binary entry point
├── internal/
│   └── core/
│       ├── config/      # YAML config parsing
│       ├── storage/     # local filesystem storage
│       ├── registry/    # plugin registration and discovery
│       └── pipeline/    # pipeline execution (middleware chaining)
├── proto/               # language-agnostic .proto files
├── sdk/
│   ├── go.mod           # SDK module (github.com/pointlet/plinth/sdk)
│   ├── sdk.go           # shared contract (interfaces, types)
│   └── serve.go         # go-plugin serve helpers for endpoint authors
└── plugin/
    ├── web/             # endpoint plugin (own go.mod, own binary)
    ├── gallery/         # endpoint plugin
    ├── compress/        # middleware plugin (Go package, imported by core)
    └── encrypt/         # middleware plugin
```

### Directory Responsibilities

| Directory | What it does |
|-----------|--------------|
| `cmd/plinth/` | Main binary entry point |
| `internal/` | Private packages that Go enforces, no external imports |
| `internal/core/config/` | YAML config parsing |
| `internal/core/storage/` | Local filesystem storage engine |
| `internal/core/registry/` | Plugin registration and health tracking |
| `internal/core/pipeline/` | Pipeline execution (chains middleware via io.Reader) |
| `proto/` | Protocol buffer definitions for endpoint plugins |
| `sdk/` | Go module, shared contract between core and plugins |
| `plugin/` | Official plugins (endpoints have own go.mod, middleware are Go packages) |

## Storage

The storage engine uses the local filesystem. Files are stored under the configured `storage.path`. The storage API exposes read, write, list, and delete operations to endpoint plugins via gRPC.

## Authentication

Authentication is deferred until the core functionality works. The eventual design:

- **Plugin-to-core auth:** API key in the config file, presented by plugins when calling core's gRPC API.
- **User-facing auth:** Owned entirely by the web plugin (sessions, passwords, OAuth).

## Design Principles

**Core stays dumb.** It stores files, runs pipelines, and proxies HTTP. That's it.

**Plugins own everything else.** UI, features, processing logic.

**Data flows through core.** Plugins never talk directly to each other.

**Crash isolation for endpoints.** One endpoint plugin dying doesn't kill the system.

**Streaming for middleware.** In-process io.Reader chains, zero serialization overhead.

**Language agnostic for endpoints.** gRPC means any language can serve routes.

**Compose your deployment.** Only load what you need.

## Roadmap

### Phase 1: Core + Middleware Pipeline

Proves the pipeline engine works.

- Core binary that reads `plinth.yaml` and initializes storage
- Local filesystem storage (read, write, list, delete)
- Pipeline execution engine (chains middleware via `io.Reader`)
- SDK with `Middleware` and `FileContext` types
- One real middleware plugin: compress (zstd or gzip)
- CLI or simple HTTP endpoint that triggers a pipeline (upload a file, get it compressed and stored)

### Phase 2: Endpoint Plugins + Crash Isolation

Proves the crash isolation model works. This is the core value proposition.

- go-plugin integration for launching and managing endpoint subprocesses
- gRPC service definition for endpoints (Health, HandleHTTP, Routes)
- Core HTTP reverse proxy (matches paths to plugins, forwards requests)
- Health check on startup with status reporting
- Crash detection and automatic restart with exponential backoff
- One real endpoint plugin: hello (registers a route, serves a page)
- Demo: kill the endpoint process, core stays up, restarts it, endpoint comes back online

### Phase 3: Web UI + Navigation

Proves the headless model and plugin discovery.

- Web endpoint plugin with Go templates + HTMX
- Navigation registration convention (endpoint plugins declare routes at startup)
- Web plugin reads registered routes from core's registry and renders navigation
- Second endpoint plugin (gallery) to prove multi-plugin navigation

### Phase 4: Polish + Ecosystem

- Encrypt middleware plugin
- Hot-reload (watch plugin binary, restart on change without restarting core)
- Non-Go plugin examples (Python endpoint)
- Authentication (API key for plugin-to-core, user auth in web plugin)
- XDG Base Directory Specification support for config file locations
- Additional middleware and endpoint plugins as needed

## License

Plinth is licensed under the [PolyForm Noncommercial License 1.0.0](LICENSE). Free to use, modify, and share for noncommercial purposes. Commercial use requires a separate agreement.
