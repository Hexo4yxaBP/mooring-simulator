# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

Phase 4 (Implementation) complete. All packages compile. All tests pass. WASM build verified.

## Project

A browser-based 2D top-down mooring simulator for sailboats.

- **Language**: Go 1.24, compiled to WASM via `GOOS=js GOARCH=wasm`
- **Engine**: Ebiten v2.9.9
- **Physics**: Custom pure-Go (no CGO — required for WASM)

## Quick Start

```bash
make build-wasm   # produces main.wasm + copies wasm_exec.js
make serve        # http://localhost:8080
```

## Testing

```bash
go test ./internal/...       # all tests
go vet ./...                 # static analysis
```

See AGENTS.md for full build/test/lint commands and code conventions.
