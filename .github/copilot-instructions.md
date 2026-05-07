# volumeleaders-go review instructions

Review this repository as a Go client library for authenticated, browser-backed VolumeLeaders endpoints. The public API should be predictable for client applications: typed request parameters, typed response structs, context propagation, explicit session handling, and explicit errors instead of process-level behavior.

Focus on bugs, security, data loss, broken API contracts, and project conventions. Do not nitpick formatting or style that `gofmt` or golangci-lint already handles.

## Project invariants

- Module path is `github.com/major/volumeleaders-go`.
- Public packages live under `volumeleaders` and `volumeleaders/browserauth`.
- The root `volumeleaders` package must stay free of browser-store, SQLite, desktop keyring, and `kooky` dependencies.
- Library code must not call `os.Exit`, write user-facing output, read hidden config files, or inspect environment variables unless a public option documents that behavior.
- Public methods that perform requests take `context.Context` as the first argument.
- Request paths, DataTables form fields, headers, JSON tags, and response structs must match captured VolumeLeaders browser traffic.
- Keep VolumeLeaders form keys, DataTables field names, request headers, and endpoint paths inside Go code. Higher-level porcelain and prompts should call typed APIs such as `TradeQuery` instead of building `url.Values` directly.
- Preserve typed errors and predicates from `volumeleaders/errors.go` so callers can classify authentication, status, body-limit, invalid-query, and unexpected-content failures.
- Keep exported identifiers documented with useful Go comments.

## Security and session safety

- Flag cookie, ASP.NET session ID, XSRF token, browser profile path, credential, or secret exposure in logs, errors, tests, docs, or generated output.
- Prefer explicit sessions with `NewSession` or `SessionFromCookies` for services and containers. Use `browserauth.New` only for desktop automation where the user is already logged in through a supported browser.
- Browser-backed auth must stay isolated to `volumeleaders/browserauth`; do not let browser-store dependencies leak into the root API package.
- Avoid silent fallback behavior around HTTP status handling, login-page detection, redirects, body decoding, or token extraction. Return clear typed errors instead.
- Capped pagination helpers must not allow unbounded LLM-triggered requests.

## API boundary expectations

- `ListTrades` and `ListTradesLimit` are the high-level trades surface for common filters. Keep raw DataTables mechanics behind typed Go APIs.
- Low-level `Get*` endpoint methods should preserve captured request shapes while exposing typed request and response models.
- Add typed Go APIs before exposing new endpoint families through prompts or porcelain.
- Do not silently mutate caller-provided cookies, headers, slices, maps, `url.Values`, or multipart form data.

## Testing expectations

- Use `testify/require` for assertions that must stop a test and `testify/assert` for non-critical checks.
- Use `httptest.NewServer()` for HTTP API mocks with inline request validation.
- Mark reusable test helpers with `t.Helper()`.
- Prefer table-driven subtests with `t.Run()`.
- Keep generated response data inline unless fixtures clearly improve readability.
- Verify request methods, paths, DataTables form fields, headers, cookies, redirects, and decoded response fields.
- Maintain the `Makefile` coverage floor. `make test` must fail if total coverage drops below 90%.

## Build and lint expectations

- CI runs `make test`, which uses gotestsum to emit `junit.xml` while preserving `-race`, `-coverprofile=coverage.out`, and the 90% coverage floor; CI also runs `go build ./...`, govulncheck, CodeQL, and golangci-lint v2.
- GoReleaser is source-only because this repository is a library with no `main` package.
- Nolint directives require a specific linter name and an explanation.
- US English spelling is enforced.
