# Contributing

## Getting Started

```bash
git clone https://github.com/major/volumeleaders-go.git
cd volumeleaders-go
make build
make test
make lint
```

## Development

**Requirements:** Go 1.26 or later.

Linting uses [golangci-lint](https://golangci-lint.run/) with the config in `.golangci.yml`. Run `make lint` before submitting.

A few conventions to follow when writing code:

- Keep the root `volumeleaders` package free of browser-store, desktop keyring, and SQLite dependencies. Browser cookie discovery belongs in `volumeleaders/browserauth`.
- Prefer typed request structs and predicates over raw form keys or string-matched errors.
- Preserve compatibility for low-level APIs such as `GetTrades` when adding higher-level wrappers.
- Add tests for auth classification, request encoding, response decoding, and external-package imports when changing exported API.

## Pull Requests

Fork the repo, create a branch, and open a PR against `main`.

- Keep PRs focused on a single change. Unrelated fixes belong in separate PRs.
- All CI checks (tests and linting) must pass before merge.
- Add tests for any new functionality.

## Code Style

Follow the patterns already in the codebase. `make lint` catches most issues. When in doubt, match the existing client, session, error, and trades APIs.

## License

By contributing, you agree that your changes will be licensed under the [Apache-2.0 License](LICENSE).
