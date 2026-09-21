# Implementation Plan: ksd Rewrite (Modernization)

Spec: [SPEC.md](../SPEC.md)

## Overview

Rewrite `ksd` to actually deliver what PR #32 attempted but failed to wire together:
a small `internal/decoder` package with the decode logic, called from a thin `main.go`,
on Go 1.27, with a modernized, security-conscious CI/CD pipeline. External CLI behavior
stays identical to what's on `main` today. PR #32 is superseded, not built upon.

## Architecture Decisions

- **`internal/decoder` as single source of decode logic, actually imported by `main.go`.**
  PR #32's core defect was adding this package and never calling it. Task 4 removes the
  duplicate inline logic from `main.go` so there is exactly one implementation.
- **No `Config`/`Decoder` struct.** PR #32 added `OutputFormat`, `PrettyPrint`,
  `ShowBinary`, `Validate` knobs that nothing in the CLI ever sets. A single `Decode`
  function matches the tool's actual (single) use case.
- **No `kind`/`apiVersion` validation.** PR #32's unused validator would have rejected
  inputs the current tool accepts. Staying permissive preserves real-world behavior
  (hand-edited or partial manifests still decode).
- **Stdlib-only CLI, `gopkg.in/yaml.v3`, Homebrew-only distribution.** Per spec answers —
  no cobra, no Krew.
- **Consolidate CI into one `ci.yml`** (test+lint+build) instead of the three
  overlapping/stale workflows currently on `main` (`ci.yml`, `code-quality.yml`,
  `golangci-lint.yml`).
- **All third-party GitHub Actions pinned to specific released versions**, never
  `@master`/`@main` — directly reverses PR #32's `securecodewarrior/github-action-gosec@master`
  and `sonatypecommunity/nancy-github-action@main`.

## Dependency Graph

```
go.mod (Go 1.27, yaml.v3, testify)
    │
    ├── internal/decoder (core logic)
    │       │
    │       ├── internal/decoder tests + testdata fixtures
    │       │
    │       └── main.go (thin entrypoint, calls decoder.Decode)
    │               │
    │               └── main_test.go (CLI-level tests)
    │
    ├── .golangci.yml (lints the rewritten code)
    │
    └── .goreleaser.yml (builds/releases the rewritten binary)
            │
            ├── .github/workflows/ci.yml (test + lint + build)
            │       │
            │       ├── remove old workflows (code-quality.yml, golangci-lint.yml)
            │       │
            │       ├── .github/workflows/release.yml (tag-triggered)
            │       │
            │       └── .github/workflows/maintenance.yml (scheduled)
            │
            └── README.md (reflects final install/usage)
```

Implementation order follows this graph bottom-up: toolchain → decoder → main.go →
tests/fixtures → lint config → CI → release config → docs → cleanup.

## Task List

### Phase 1: Foundation

- [x] Task 1: Bump toolchain and dependencies

### Checkpoint: Foundation
- [x] `go build ./...` and `go vet ./...` succeed on the bumped toolchain

### Phase 2: Core Rewrite (vertical slice — decode works end-to-end)

- [x] Task 2: Implement `internal/decoder` package
- [x] Task 3: Test `internal/decoder` + add `testdata/` fixtures
- [x] Task 4: Rewrite `main.go` to call `internal/decoder`, remove duplicate logic
- [x] Task 5: Rewrite `main_test.go` for CLI-level concerns only

### Checkpoint: Core Rewrite
- [x] `go test -race ./...` passes
- [x] Manual piping test (`./ksd < testdata/secret.json`, `.yaml`, `./ksd version`)
      matches current `main`-branch output (byte-for-byte, after fixing a YAML
      indentation regression from the yaml.v2→v3 switch)
- [x] Reviewed with human — approved to proceed

### Phase 3: Tooling

- [x] Task 6: Restore `.golangci.yml`
- [x] Task 7: Remove root-level `mock.json`/`mock.yml`, clean up `.gitignore`

### Phase 4: CI/CD Pipeline

- [x] Task 8: Rewrite `.github/workflows/ci.yml` (test + lint + build, pinned actions)
- [x] Task 9: Remove redundant `code-quality.yml` / `golangci-lint.yml`
- [x] Task 10: Write `.github/workflows/release.yml` (tag-triggered, GoReleaser v2)
- [x] Task 11: Rewrite `.goreleaser.yml` for v2 (Homebrew only, no Krew)
- [x] Task 12: Write `.github/workflows/maintenance.yml` (govulncheck + dependency PR)

### Checkpoint: Pipeline
- [x] All workflow YAML reviewed — no floating (`@master`/`@main`) third-party action refs
- [x] `goreleaser check` and `goreleaser release --snapshot --clean` succeed locally

### Phase 5: Docs & Cleanup

- [x] Task 13: Update `README.md`
- [x] Task 14: Final full verification pass (build/vet/test/lint/goreleaser + output parity)
- [ ] Task 15 (ask first): Close PR #32, pointing to the replacement work

### Checkpoint: Complete
- [x] All SPEC.md success criteria met
- [ ] Ready for human review / merge — pending your review; Task 15 (closing PR #32) still needs your explicit go-ahead

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Go map iteration order is non-deterministic, so multi-key JSON/YAML output order isn't stable | Med | Assert on parsed structures (`assert.Equal` on maps), not raw multi-key output strings; only single-key cases get raw-string assertions |
| `golangci-lint` default ruleset surfaces latent issues (e.g. `errcheck` on currently-ignored errors) | Low | Fix as found during Task 6; keep the config reasonably scoped rather than maximal, to avoid unrelated churn |
| GoReleaser v1→v2 config migration has breaking syntax changes (e.g. `replacements` removed, `name_template` required) | Med | Validate with `goreleaser check` before considering Task 11 done; PR #32's (broken but directionally useful) v2 config is a reference, not a copy source |
| Homebrew tap publish needs a cross-repo token not available in this sandbox | Low | Only verify via `goreleaser release --snapshot --clean` (no publish) locally; real tap publish happens on an actual tag push, outside this session's scope |
| Closing PR #32 is a visible, hard-to-reverse GitHub action | Low | Explicit ask-first task (Task 15); not done automatically even after everything else is verified |

## Open Questions

- None — deferred items (Krew, feature additions, CLI framework choice) were already
  resolved as out-of-scope in SPEC.md.
