# ksd — Kubernetes Secret Decoder

[![codecov](https://codecov.io/gh/mfuentesg/ksd/branch/main/graph/badge.svg)](https://codecov.io/gh/mfuentesg/ksd)
[![Go Reference](https://pkg.go.dev/badge/github.com/mfuentesg/ksd.svg)](https://pkg.go.dev/github.com/mfuentesg/ksd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<a href="https://www.buymeacoffee.com/mfuentesg" target="_blank">
   <img height="41" src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" />
</a>

`ksd` is a small command-line filter that decodes the base64-encoded `data`
field of a Kubernetes Secret manifest into a human-readable `stringData`
field. It reads a `Secret` (YAML or JSON, auto-detected) from stdin and
writes the decoded manifest back out — every other field is left untouched.

```
$ kubectl get secret my-secret -o yaml | ksd
```

## Installation

### Krew (kubectl plugin manager)

```bash
kubectl krew install ksd
```

### Homebrew (macOS/Linux)

```bash
brew install --cask mfuentesg/tap/ksd
```

### Go install

```bash
go install github.com/mfuentesg/ksd@latest
```

### Download a prebuilt binary

Grab the archive for your platform (linux/darwin/windows, amd64/arm64) from
the [releases page](https://github.com/mfuentesg/ksd/releases), extract it,
and put the `ksd` binary on your `PATH`.

## Usage

```
ksd < secret.<yaml|json>
kubectl get secret <secret-name> -o <yaml|json> | ksd
```

Installed via Krew? Use `kubectl ksd` in place of `ksd` — same behavior,
same flags:

```
kubectl get secret <secret-name> -o <yaml|json> | kubectl ksd
```

```
ksd version   # print the ksd version
```

Running `ksd` without piping anything into it prints a usage message to
stderr and exits non-zero — there's no separate `--help` flag.

## Examples

### JSON in, JSON out

> secret.json
```json
{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes-secret-decoder",
        "namespace": "ksd"
    },
    "type": "Opaque",
    "data": {
        "password": "c2VjcmV0",
        "app": "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="
    }
}
```

```
$ ksd < secret.json
```

```json
{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes-secret-decoder",
        "namespace": "ksd"
    },
    "stringData": {
        "app": "kubernetes secret decoder",
        "password": "secret"
    },
    "type": "Opaque"
}
```

### YAML in, YAML out

> secret.yaml
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubernetes-secret-decoder
  namespace: ksd
type: Opaque
data:
  password: c2VjcmV0
  app: a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==
```

```
$ ksd < secret.yaml
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubernetes-secret-decoder
  namespace: ksd
stringData:
  app: kubernetes secret decoder
  password: secret
type: Opaque
```

### Straight from the cluster

```bash
kubectl get secret my-secret -o yaml | ksd
```

### Round-tripping through kubectl apply

Because `ksd` preserves the full manifest shape, its output can be piped
straight back into `kubectl apply` — Kubernetes accepts `stringData` on
write, base64-encoding it for you:

```bash
kubectl get secret my-secret -o yaml | ksd | kubectl apply -f -
```

### Save a plaintext copy for review or diffing

```bash
kubectl get secret my-secret -o yaml | ksd > my-secret.decoded.yaml
```

Decoded values are printed in plaintext — be mindful about where that
output ends up (shared terminals, CI logs, committed files, etc).

## Development

Requires Go 1.27+.

```bash
git clone https://github.com/mfuentesg/ksd.git
cd ksd

go build -o ksd .              # build
go test -race -cover ./...     # test
go vet ./...                   # vet
golangci-lint run              # lint (golangci-lint v2)
```

Project layout:

```
main.go             # CLI entrypoint: arg parsing, stdin/stdout, exit codes
internal/decoder/   # decode logic (the only place format detection/parsing lives)
testdata/           # fixtures used by both the decoder and CLI tests
.goreleaser.yml      # release config: binaries, Homebrew cask, Krew manifest
```

Releases are automated with [GoReleaser](https://goreleaser.com): pushing a
`v*` tag builds cross-platform binaries, publishes a GitHub release, updates
the [Homebrew tap](https://github.com/mfuentesg/homebrew-tap), and generates
the Krew plugin manifest.

## See also

- [kubectl-view-secret](https://github.com/elsesiy/kubectl-view-secret) — another kubectl plugin for viewing/decoding Kubernetes secrets, also available via Krew.

## License

[MIT](LICENSE)
