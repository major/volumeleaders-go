---
applyTo: "**/*.go"
---

# Go review instructions

- Prefer small, focused functions with explicit error handling and clear return values.
- Use `%w` when wrapping errors with `fmt.Errorf`.
- Preserve context propagation for API calls and cancellation.
- Keep exported identifiers documented with useful Go comments.
- Do not add process-level behavior such as `os.Exit`, stdout writes, or hidden environment/config reads to library packages.
- Request builders should keep paths and DataTables fields named exactly as captured VolumeLeaders traffic expects them.
- Response structs should expose useful fields to callers instead of forcing callers to parse raw JSON unless the response shape is genuinely endpoint-specific or not yet modeled.
- Keep browser-store, SQLite, desktop keyring, and `kooky` dependencies isolated to `volumeleaders/browserauth`.
- Do not suggest style-only changes that `gofmt` or golangci-lint already enforces.
