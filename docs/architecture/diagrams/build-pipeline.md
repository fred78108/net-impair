# Build Pipeline

The frontend is compiled at build time and embedded directly into the Go binary via `//go:embed`. The resulting binary is fully self-contained — no separate web server or runtime dependency is required.

Related: [ADR-001](../decisions/ADR-001-language-golang.md) · [ADR-003](../decisions/ADR-003-frontend-react-vite.md)

```mermaid
flowchart LR
    subgraph src["Source"]
        go_src["Go source\ncmd/ · internal/"]
        ts_src["React + TypeScript\nfrontend/src/"]
    end

    subgraph build["Build Steps"]
        npm["npm run build\n(Vite)"]
        dist["frontend/dist/\nHTML · JS · CSS · assets"]
        embed["//go:embed frontend/dist\n(compile-time inclusion)"]
        gobuild["go build"]
    end

    binary["net-impair binary\n(single static binary\nper target platform)"]

    ts_src --> npm
    npm --> dist
    dist --> embed
    go_src --> gobuild
    embed --> gobuild
    gobuild --> binary

    style binary fill:#e8ffd0,stroke:#5aa13a
```

## Cross-Platform Targets

```mermaid
flowchart LR
    gobuild["go build"]
    linux["net-impair\n(linux/amd64 · arm64)"]
    macos["net-impair\n(darwin/amd64 · arm64)"]
    windows["net-impair.exe\n+ wintun.dll\n(windows/amd64)"]

    gobuild -->|"GOOS=linux"| linux
    gobuild -->|"GOOS=darwin"| macos
    gobuild -->|"GOOS=windows"| windows
```
