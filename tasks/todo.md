# Task List: ksd Rewrite (Modernization)

Plan: [plan.md](plan.md) · Spec: [../SPEC.md](../SPEC.md)

---

## Phase 1: Foundation

### Task 1: Bump toolchain and dependencies ✅

**Description:** Update `go.mod` to `go 1.27`, bump `gopkg.in/yaml.v2` → `gopkg.in/yaml.v3`
and `stretchr/testify` to its current release, then `go mod tidy`.

**Acceptance criteria:**
- [x] `go.mod` declares `go 1.27`
- [x] `gopkg.in/yaml.v3` and current `testify` versions are the only direct deps
- [x] `go mod tidy` produces no further diff

**Verification:**
- [x] `go build ./...` succeeds
- [x] `go vet ./...` succeeds

**Note:** `yaml.v2` remained a direct dependency until Task 4 removed its last import
from `main.go`; `go mod tidy` was re-run then to drop it. testify landed at v1.12.1.

**Dependencies:** None

**Files likely touched:**
- `go.mod`
- `go.sum`

**Estimated scope:** XS (1-2 files)

---

## Phase 2: Core Rewrite

### Task 2: Implement `internal/decoder` package ✅

**Description:** Create the single source of decode logic per SPEC.md's code-style
section: one exported `Decode([]byte) ([]byte, error)` function, no `Config`/`Decoder`
struct, no `kind`/`apiVersion` validation. Auto-detects JSON vs YAML, decodes `data` into
`stringData`, passes through unchanged when there's no `data` field, converts non-string
values via `fmt.Sprintf`, leaves invalid-base64 values as-is.

**Acceptance criteria:**
- [x] Package exposes exactly one public entry point (`Decode`)
- [x] Behavior matches current `main`-branch semantics (see `main.go`/`main_test.go` on
      `main` for the reference behavior — not PR #32's stricter version)
- [x] No panics on malformed input; errors returned as-is (no unreachable wrapping)

**Verification:**
- [x] `go build ./internal/...`

**Found during implementation:** `yaml.v3` propagates the destination's *named* map
type onto nested mappings (unlike a plain `map[string]interface{}`), which silently
broke the `data`-field type assertion. Fixed by decoding into an unnamed
`map[string]interface{}` — documented inline in `decoder.go` since it's a non-obvious
library gotcha. Also had to explicitly set YAML output indent to 2 (yaml.v3 defaults to
4) to match yaml.v2's original output byte-for-byte.

**Dependencies:** Task 1

**Files likely touched:**
- `internal/decoder/decoder.go`

**Estimated scope:** S (1 file)

---

### Task 3: Test `internal/decoder` + add `testdata/` fixtures ✅

**Description:** Table-driven tests covering every branch: JSON input, YAML input,
missing `data` field, non-string values, invalid base64 (passthrough), invalid
JSON/YAML (error), empty input (error). Add `testdata/secret.json` and
`testdata/secret.yaml` as golden fixtures for the two happy-path cases (replacing the
root-level `mock.json`/`mock.yml`, removed in Task 7).

**Acceptance criteria:**
- [x] Every branch identified above has a test case
- [x] Happy-path JSON/YAML tests load from `testdata/`, not inline literals
- [x] Edge-case tests (invalid input, missing field) may use inline literals

**Verification:**
- [x] `go test -race ./internal/... -cover` — 95.6% (100% on all functions except
      `marshal`'s two YAML-encoder error branches, which are unreachable given the
      input domain — see note in plan.md risks; not worth a contrived test)

**Correction from original spec:** empty input does **not** error on `main` today (it
passes through silently, exit 0) — PR #32 introduced that error behavior, which we
deliberately did not carry over. Verified against the actual `main`-branch binary
before writing this test.

**Dependencies:** Task 2

**Files likely touched:**
- `internal/decoder/decoder_test.go`
- `testdata/secret.json`
- `testdata/secret.yaml`

**Estimated scope:** S (3 files)

---

### Task 4: Rewrite `main.go` ✅

**Description:** Reduce `main.go` to argument parsing, stdin/stdout plumbing, and exit
codes, calling `internal/decoder.Decode`. Remove the inline `cast`/`decode`/`parse`/
`marshal`/`unmarshal`/`isJSONString` functions — that logic now lives solely in
`internal/decoder`. Preserve current behavior exactly: `version` output, tty-detection
usage message on stderr, non-zero exit on decode failure.

**Acceptance criteria:**
- [x] `main.go` contains no parse/decode/marshal logic — only arg handling, stdin read,
      the call to `decoder.Decode`, and output/exit-code handling
- [x] `ksd version` output unchanged from current `main`
- [x] Non-pipe invocation still prints the current usage message to stderr and exits 1

**Verification:**
- [x] `go build -o ksd .`
- [x] `./ksd < testdata/secret.json` and `./ksd < testdata/secret.yaml` produce the
      expected decoded output, verified byte-for-byte against the original `main`
      binary
- [x] `./ksd version` prints the version string

**Deviation:** `io.ReadAll(os.Stdin)`'s error is now checked (original silently
discarded it). Harmless in practice, but avoids an `errcheck` lint failure ahead of
Task 6, so folding it in here avoided rework.

**Dependencies:** Task 3

**Files likely touched:**
- `main.go`

**Estimated scope:** S (1 file)

---

### Task 5: Rewrite `main_test.go` ✅

**Description:** Test only CLI-level concerns now that decode logic has its own test
suite: `version` output, tty/usage-message path, exit codes, end-to-end piping through
the `testdata/` fixtures. Remove tests that duplicate `internal/decoder`'s coverage.

**Acceptance criteria:**
- [x] No test in `main_test.go` re-tests decode/parse/marshal internals
- [x] CLI-level behaviors (version, usage message, exit codes, pipe-through) are covered

**Verification:**
- [x] `go test -race ./...`

**Implementation note:** since `main.go` intentionally has no exported/testable
functions (per SPEC.md's "only main.go touches os.Args/os.Stdin/os.Exit" rule),
`main_test.go` builds the real binary once in `TestMain` and drives it as a subprocess
— this exercises the actual stdin/stdout/exit-code contract rather than internals, at
the cost of `go tool cover` reporting 0% for package `main` (expected and acceptable;
all real logic coverage lives in `internal/decoder`).

**Dependencies:** Task 4

**Files likely touched:**
- `main_test.go`

**Estimated scope:** S (1 file)

---

## Checkpoint: Core Rewrite Complete ✅

- [x] `go test -race ./...` passes
- [x] Manual smoke test output matches current `main` branch for the same inputs —
      verified: JSON decode identical, YAML decode identical (after fixing indent),
      `version` identical, empty-input identical (exit 0, no output), invalid-input
      identical (same error text, exit 1), no-data-field passthrough identical,
      no-pipe usage message identical modulo `os.Args[0]` (the binary path, expected
      to differ)
- [ ] **Pause here for human review before starting Phase 3** — awaiting go-ahead

---

## Phase 3: Tooling

### Task 6: Restore `.golangci.yml` ✅

**Description:** Add a lint config with a reasonable, non-maximal ruleset (govet,
staticcheck, errcheck, unused, ineffassign, gofmt/gofumpt) matching the codebase's small
size — not the 43-line config that was previously deleted verbatim, but something fit
for the rewritten code.

**Acceptance criteria:**
- [ ] `.golangci.yml` exists and is valid
- [ ] `golangci-lint run` passes against the Phase-2 code with zero suppressions added
      just to make it pass (fix real issues instead)

**Verification:**
- [x] `golangci-lint run` — 0 issues

**Dependencies:** Task 5

**Files likely touched:**
- `.golangci.yml`

**Estimated scope:** XS (1 file)

**Found during implementation:** the locally installed `golangci-lint` (built with
go1.24) refused to run against a `go 1.27` module — had to reinstall it via
`go install .../golangci-lint/v2@latest` using the local Go 1.27.1 toolchain. CI runners
pull a fresh binary each run so this is a local-machine-only issue, but worth knowing if
it recurs. Lint also caught two real bugs in `main_test.go`'s `TestMain` (a `defer`
after `os.Exit` that would never run, and an unchecked `os.RemoveAll` error) — fixed
both.

---

### Task 7: Remove root-level mocks, clean up `.gitignore` ✅

**Description:** Delete `mock.json`/`mock.yml` (superseded by `testdata/`). Review
`.gitignore` for duplicate/stale entries (currently has both `dist/` implicitly and a
bare `release` entry) and ensure build/coverage artifacts (`ksd`, `coverage.out`,
`coverage.html`, `dist/`) are covered without duplication.

**Acceptance criteria:**
- [x] No references to `mock.json`/`mock.yml` remain anywhere in the repo (also fixed
      one stale reference in `SPEC.md`'s example command)
- [x] `.gitignore` has no duplicate entries and covers all build/coverage artifacts

**Verification:**
- [x] `grep -rn "mock\.\(json\|yml\)" --include=*.go .` returns nothing

**Dependencies:** Task 3

**Files likely touched:**
- `mock.json` (deleted)
- `mock.yml` (deleted)
- `.gitignore`

**Estimated scope:** XS (2-3 files)

---

## Phase 4: CI/CD Pipeline

### Task 8: Rewrite `.github/workflows/ci.yml` ✅

**Description:** Single consolidated workflow triggered on push/PR to `main`: a test job
(`go test -race -coverprofile=coverage.out ./...` on Go 1.27), a lint job
(`golangci-lint-action` pinned to a specific version), and a build job (`go build` +
smoke-test the binary against `testdata/` fixtures). All actions pinned to specific
released versions.

**Acceptance criteria:**
- [x] Workflow triggers on push/PR to `main`
- [x] test, lint, build, goreleaser-check, and security (govulncheck + gosec) jobs
- [x] No action reference uses `@master`, `@main`, or `latest`-style floating tags —
      every action pinned to a real current major-version tag, checked via `gh api
      repos/<org>/<repo>/releases/latest`

**Verification:**
- [x] `actionlint .github/workflows/*.yml` — clean

**Addressing the user's request for binary artifacts from CI:** the `build` job now
uploads the compiled linux/amd64 binary via `actions/upload-artifact` on every run
(previously it only built-and-discarded to smoke-test). A `goreleaser-check` job also
runs a full cross-platform `goreleaser release --snapshot --clean` on every PR, so
release-config regressions surface before a tag is ever pushed.

**Dependencies:** Task 6

**Files likely touched:**
- `.github/workflows/ci.yml`

**Estimated scope:** S (1 file)

---

### Task 9: Remove redundant workflows ✅

**Description:** Delete `.github/workflows/code-quality.yml` and
`.github/workflows/golangci-lint.yml` now that `ci.yml` covers test+lint.

**Acceptance criteria:**
- [x] Only one workflow runs tests and only one runs lint across `.github/workflows/`

**Verification:**
- [x] `ls .github/workflows/` shows no overlapping responsibility

**Dependencies:** Task 8

**Files likely touched:**
- `.github/workflows/code-quality.yml` (deleted)
- `.github/workflows/golangci-lint.yml` (deleted)

**Estimated scope:** XS (2 files)

---

### Task 10: Write `.github/workflows/release.yml` ✅

**Description:** New workflow triggered on `v*` tag push: run tests, then run
GoReleaser v2 (pinned action version) with `GITHUB_TOKEN`. No Krew job (out of scope).

**Acceptance criteria:**
- [x] Triggers only on `v*` tags
- [x] Runs tests before release
- [x] GoReleaser action pinned to a specific version, not `@master`

**Verification:**
- [x] `actionlint` — clean

**Note:** the release job also passes through a `HOMEBREW_TAP_GITHUB_TOKEN` secret,
needed because publishing to the separate `mfuentesg/homebrew-tap` repo requires a
token with write access there — the default `GITHUB_TOKEN` is scoped to this repo only.
**You'll need to create that secret in this repo's settings before the first real tag
release**, or GoReleaser will fail at the Homebrew-cask publish step.

**Dependencies:** Task 8

**Files likely touched:**
- `.github/workflows/release.yml`

**Estimated scope:** S (1 file)

---

### Task 11: Rewrite `.goreleaser.yml` for v2 ✅

**Description:** `version: 2` config: builds for linux/darwin/windows × amd64/arm64,
ldflags embedding version/commit/date, checksum, changelog grouped by conventional-commit
type, Homebrew tap publish (`mfuentesg/homebrew-tap`) only — no `krews` section.

**Acceptance criteria:**
- [x] `goreleaser check` passes
- [x] `goreleaser release --snapshot --clean` succeeds locally — verified all 6
      linux/darwin/windows × amd64/arm64 binaries build and archive correctly
- [x] No Krew-related config present

**Verification:**
- [x] `goreleaser check`
- [x] `goreleaser release --snapshot --clean`

**Deviation (flagged and confirmed with you):** GoReleaser v2 deprecated `brews` in
favor of `homebrew_casks`. You chose to migrate now rather than keep the deprecated
(but still functional) `brews` key. This means `brew install mfuentesg/tap/ksd` becomes
`brew install --cask mfuentesg/tap/ksd`, and the separate `mfuentesg/homebrew-tap` repo
needs a `Casks/` directory instead of (or alongside) `Formula/` — that repo isn't part
of this codebase, so you'll need to set that up there directly. Also added explicit
`format_overrides` so Windows archives ship as `.zip` (Linux/macOS stay `.tar.gz`) —
the original config had no format override, which would've given Windows users a
`.tar.gz` too.

**Dependencies:** Task 1

**Files likely touched:**
- `.goreleaser.yml`

**Estimated scope:** S (1 file)

---

### Task 12: Write `.github/workflows/maintenance.yml` ✅

**Description:** Scheduled (weekly) workflow: `govulncheck ./...` scan, plus a
`go get -u && go mod tidy` job that opens a PR when dependencies change. Actions pinned
to specific versions; permissions scoped minimally per job.

**Acceptance criteria:**
- [x] Runs on a weekly `schedule` trigger plus `workflow_dispatch`
- [x] `govulncheck` job present
- [x] Dependency-update job opens a PR only when `go.mod`/`go.sum` actually changed
- [x] No action reference uses `@master`/`@main`

**Verification:**
- [x] `actionlint` — clean

**Scope addition (to fully match the "full automation suite" you chose in SPEC.md,
which explicitly listed semantic PR titles, auto-labeling, and gosec — I'd only done
Codecov + govulncheck + the dependency bot at first pass):**
- Added `gosec` (pinned to `v2.29.0`, run via `go install` rather than a third-party
  action — one less floating dependency) to `ci.yml`'s security job, run on every PR.
- Added a new `.github/workflows/pr-validation.yml`: Conventional-Commits PR title
  check (`amannn/action-semantic-pull-request@v6`) and file-based auto-labeling
  (`dorny/paths-filter@v4` + `actions/github-script@v9`). Deliberately dropped PR #32's
  size-check and breaking-change-label jobs — not part of what was actually requested.

**Dependencies:** Task 8

**Files likely touched:**
- `.github/workflows/maintenance.yml`

**Estimated scope:** S (1 file)

---

## Checkpoint: Pipeline Complete ✅

- [x] No third-party action anywhere in `.github/workflows/` is pinned to `@master`/`@main`
      — every one checked against its actual latest release tag via `gh api`
- [x] `goreleaser check` and `goreleaser release --snapshot --clean` succeed locally
- [x] All 5 workflow files pass `actionlint` with zero findings

---

## Phase 5: Docs & Cleanup

### Task 13: Update `README.md` ✅

**Description:** Reflect the final install (Go install + Homebrew, no Krew) and usage
flow; verify examples match the `testdata/` fixtures and actual CLI output.

**Acceptance criteria:**
- [x] No Krew references
- [x] Install/usage examples are accurate and runnable — Homebrew install updated to
      `brew install --cask ...` per the Task 11 cask migration

**Verification:**
- [x] Ran the documented example against the real binary; found and fixed one
      discrepancy (README showed `password` before `app` in `stringData`, but
      `json.Marshal` sorts map keys alphabetically, so `app` comes first)

**Dependencies:** Task 5, Task 11

**Files likely touched:**
- `README.md`

**Estimated scope:** XS (1 file)

---

### Task 14: Final full verification pass ✅

**Description:** Run the complete check suite and confirm output parity with the
pre-rewrite `main` branch for identical inputs.

**Acceptance criteria:**
- [x] `go build ./...`, `go vet ./...` clean
- [x] `go test -race -cover ./...` passes (main: 0% — subprocess-tested by design;
      internal/decoder: 95.6%)
- [x] `golangci-lint run` passes — 0 issues
- [x] `gosec ./...` passes — 0 issues (added beyond the original acceptance list, per
      the "full automation suite" scope)
- [x] `goreleaser check` and `goreleaser release --snapshot --clean` pass
- [x] Piping `testdata/secret.json` and `testdata/secret.yaml` through the rewritten
      `ksd` produces output equivalent to the current `main` branch for the same
      inputs — confirmed at the Phase 2 checkpoint via direct binary diff

**Verification:** all of the above, run directly

**Dependencies:** Task 12, Task 13

**Files likely touched:** none (verification only)

**Estimated scope:** XS (0 files)

---

### Task 15: Close PR #32 (ask first)

**Description:** Once the rewrite is merged, close PR #32 with a comment linking to the
replacement work. **Do not do this automatically — confirm with the human first**, per
SPEC.md boundaries.

**Acceptance criteria:**
- [ ] Human has explicitly approved closing PR #32
- [ ] PR #32 closed with a comment pointing to the merged replacement

**Verification:** PR #32 shows as closed with an explanatory comment

**Dependencies:** Task 14

**Files likely touched:** none (GitHub action, not a file change)

**Estimated scope:** XS

---

## Checkpoint: Complete

- [ ] All SPEC.md success criteria met
- [ ] Ready for human review / merge
