# Mooring Simulator

Browser mooring simulator for sailboats and catamarans (top-down schematic dock, boats, lines, wind, engine, rudder).

## Current stack (active)

- TypeScript + Node
- Vite front end (`src/`, `vite.config.ts`, `package.json`)

```bash
npm install
npm run dev
```

## Legacy

An earlier Go → WebAssembly path remains in the repo (`main.go`, `internal/`, `Makefile` `build-wasm`). It is not the maintained implementation. Prefer the TypeScript app above.

## Domain

- Place boats near the dock
- Mooring lines / springs (boat side at cleats only)
- Wind, throttle, rudder, prop walk, center of mass
- Contact with dock / other boats

## Status

Personal project. Active development is on the TypeScript side.

## License

Add a license before publishing.
