# Spec: ksd Rewrite (Modernization)

## Objective

`ksd` (Kubernetes Secret Decoder) is a small CLI/Unix filter: it reads a Kubernetes
`Secret` manifest (YAML or JSON) from stdin, base64-decodes the `data` field into a
human-readable `stringData` field, and writes the result to stdout. It's used as
`kubectl get secret <name> -o yaml | ksd`.

This is a **behavior-identical rewrite**, not a feature addition. Goals:

- Replace PR #32 (unmerged, unmergeable as-is — adds a fully-featured `internal/decoder`
  package that `main.go` never calls, requiring `kind: Secret` where today's tool is
  permissive, and commits a compiled binary into the repo) with a clean rewrite that
  actually wires the internal package into `main.go`.
- Modernize the toolchain: Go 1.27, current dependency versions, current GitHub Actions.
- Modernize the CI/CD pipeline: test/lint/build on every PR, a real lint config, pinned
  (non-`@master`) third-party actions, and a tag-triggered GoReleaser v2 release.
- Ship the same user-facing behavior the current `main`-branch tool has today — same
  flags, same stdin/stdout contract, same permissiveness (no new hard validation that
  would reject inputs the tool accepts today).

**Users:** developers/operators piping `kubectl get secret ... -o yaml|json` output
through `ksd` to read secret values without manually base64-decoding each field.

**Success looks like:** `go install github.com/mfuentesg/ksd@latest` (or
`brew install mfuentesg/tap/ksd`) gives a Go-1.27-built binary with identical CLI
behavior to today's `main`, backed by a passing, lint-clean, security-scanned CI
pipeline, with PR #32 closed in favor of this work.

## Tech Stack

- **Language:** Go 1.27 (`go.mod` sets `go 1.27`)
- **YAML:** `gopkg.in/yaml.v3` (dropping the older `yaml.v2`)
- **JSON:** stdlib `encoding/json`
- **CLI:** stdlib only — `os.Args`/`flag`, no cobra/urfave-cli. This is a single-purpose
  Unix filter; a CLI framework is unjustified dependency weight for it.
- **Testing:** `github.com/stretchr/testify` (already in use) — `assert`/`require`
- **Lint:** `golangci-lint` v2, config restored at `.golangci.yml` (deleted from the repo
  in an earlier commit and never replaced)
- **Release:** GoReleaser v2 (`.goreleaser.yml` `version: 2`), Homebrew tap
  (`mfuentesg/homebrew-tap`) as the sole binary-distribution channel — no Krew plugin
  (out of scope; PR #32's Krew workflow/manifest/issue-template are dropped)
- **Security scanning:** `govulncheck` and `gosec`, both pinned to specific released
  versions (never `@master`/`@main`)

## Commands

```bash
# Build
go build -o ksd .

# Run tests (race detector + coverage)
go test -race -coverprofile=coverage.out ./...

# View coverage
go tool cover -func=coverage.out

# Lint
golangci-lint run

# Vet
go vet ./...

# Format
gofmt -l .

# Vulnerability scan
govulncheck ./...

# Local release dry-run
goreleaser release --snapshot --clean

# Manual smoke test
kubectl get secret <name> -o yaml | ./ksd
./ksd < testdata/secret.yaml
./ksd version
```

All of the above (except the release dry-run) run in CI on every pull request.

## Project Structure

```
.
├── main.go                    # Thin entrypoint: arg parsing, stdin/stdout wiring, exit codes
├── main_test.go                # CLI-level tests (arg handling, stdin plumbing, exit codes)
├── internal/
│   └── decoder/
│       ├── decoder.go          # Decode logic: format detection, cast, base64 decode, marshal
│       └── decoder_test.go     # Table-driven unit tests for decoder logic
├── testdata/                   # Renamed from mock.json/mock.yml — golden fixture files
│   ├── secret.json
│   └── secret.yaml
├── .golangci.yml                # Restored lint config
├── .goreleaser.yml              # v2 config: build, archive, checksum, changelog, brews
├── .github/
│   └── workflows/
│       ├── ci.yml               # test + lint + build, on push/PR to main
│       ├── release.yml          # GoReleaser release, on `v*` tag push
│       └── maintenance.yml      # Scheduled govulncheck + dependency-update PR
├── go.mod / go.sum
├── LICENSE
└── README.md
```

`main.go` is the only file allowed to call `os.Exit`, read `os.Stdin`, or touch
`os.Args`. All decode/parse/marshal logic lives in `internal/decoder` so it's testable
without a process boundary — and, unlike PR #32, `main.go` actually imports and calls it.

## Code Style

Follows standard Go conventions (`gofmt`, `go vet`, `golangci-lint` clean). Example of
the target shape for the decoder package:

```go
// internal/decoder/decoder.go
package decoder

// Decode converts a Kubernetes Secret manifest's base64-encoded `data` field
// into a plaintext `stringData` field, preserving every other field untouched.
// It auto-detects JSON vs YAML input and returns output in the same format.
func Decode(in []byte) ([]byte, error) {
    format := detectFormat(in)

    doc, err := unmarshal(in, format)
    if err != nil {
        return nil, fmt.Errorf("parsing input: %w", err)
    }

    data, ok := stringMap(doc["data"])
    if !ok || len(data) == 0 {
        return in, nil // nothing to decode; return input unchanged
    }

    doc["stringData"] = decodeValues(data)
    delete(doc, "data")

    return marshal(doc, format)
}
```

Conventions:
- Exported package API is a small set of functions (`Decode`, not a `Decoder` struct
  with a `Config` object) — no speculative configuration knobs (`ShowBinary`,
  `Validate`, `OutputFormat` overrides) that nothing in the CLI ever sets, unlike PR #32.
- Errors wrapped with `%w` and enough context to be useful in a one-line CLI error
  message (`could not decode secret: %v`).
- No panics outside of truly unrecoverable programmer errors.
- Comments only where behavior is non-obvious (e.g., why invalid-base64 values pass
  through unchanged instead of erroring).

## Testing Strategy

- **Framework:** `go test` + `testify` (`assert`/`require`), table-driven tests.
- **Location:** unit tests live beside the code they test (`internal/decoder/decoder_test.go`,
  `main_test.go`), matching current convention.
- **Levels:**
  - Unit tests on `internal/decoder` cover: JSON input, YAML input, missing `data`
    field, non-string values in `data`, invalid base64 (passthrough), invalid
    JSON/YAML (error), empty input (error).
  - `main_test.go` covers CLI-level concerns: `version` flag output, stdin-is-a-tty
    usage message, non-zero exit codes on decode failure.
- **Fixtures:** golden files under `testdata/` (JSON and YAML variants of the same
  secret), replacing the ad hoc root-level `mock.json`/`mock.yml`.
- **Coverage:** no hard percentage gate given the codebase's size, but every branch in
  `internal/decoder` must have a corresponding test case — CI runs
  `go test -race -coverprofile=coverage.out ./...` and coverage is visible via
  Codecov, informational only (not a merge blocker).
- **Regression check:** every current passing test in `main_test.go` (as of `main`,
  not PR #32) must have an equivalent passing after the rewrite — no silent behavior
  changes.

## Boundaries

**Always do:**
- Run `go test ./...`, `go vet ./...`, and `golangci-lint run` before considering any
  task complete.
- Preserve the current CLI's external behavior exactly (stdin/stdout contract, `version`
  flag, error messages' meaning if not their exact text, exit codes).
- Pin third-party GitHub Actions to a specific released version or SHA — never
  `@master`/`@main`/`@latest` branch refs.
- Keep decode logic permissive by default (no new hard-fail validation on `kind`/
  `apiVersion` unless explicitly requested) — this preserves today's real-world
  tolerance for partial/hand-edited manifests.

**Ask first:**
- Adding any new runtime dependency beyond `yaml.v3` and `testify`.
- Any change to the public CLI surface (new flags, new subcommands, changed output
  format).
- Closing PR #32.
- Changing the release/distribution channel (e.g., re-adding Krew, adding npm/apt, etc).
- Modifying `.github/workflows/*` secrets or permissions beyond what's specified here.

**Never do:**
- Commit a compiled binary into the repository (PR #32 did this — `ksd` binary
  committed at repo root).
- Introduce a breaking change to the `data` → `stringData` conversion behavior without
  it being called out explicitly as a version bump.
- Skip or weaken CI checks (lint, test, vet) to make a PR pass.

## Success Criteria

- [ ] `go build ./...` and `go vet ./...` succeed on Go 1.27.
- [ ] `golangci-lint run` passes with a real (non-empty) `.golangci.yml`.
- [ ] All existing behavior from current `main` (`version`, stdin pipe decode, tty
      usage message, JSON/YAML auto-detect) is preserved and covered by tests.
- [ ] `internal/decoder` is the single source of decode logic; `main.go` contains no
      duplicate parse/decode/marshal code.
- [ ] No compiled binaries are tracked in git.
- [ ] CI runs test + lint + build on every PR to `main`; release runs on `v*` tags via
      GoReleaser v2 with actions pinned to specific versions.
- [ ] PR #32 is closed once this work supersedes it (with your go-ahead).
- [ ] README reflects the actual install/usage flow (no Krew references, since Krew
      support is out of scope).

## Open Questions

- None currently — all prior ambiguities were resolved via clarifying questions
  (scope: behavior-identical; CLI: stdlib-only; YAML lib: yaml.v3; CI ambition: full
  automation suite).
