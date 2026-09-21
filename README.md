# Kubernetes secret decoder a.k.a `ksd`


[![codecov](https://codecov.io/gh/mfuentesg/ksd/branch/main/graph/badge.svg)](https://codecov.io/gh/mfuentesg/ksd)

<a href="https://www.buymeacoffee.com/mfuentesg" target="_blank">
   <img height="41" src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" />
</a>

<br />
<br />

`ksd` is a tool, whose aim is help you to visualize in text plain your kubernetes secrets, either `yaml` or `json` outputs.

## Installation

### Krew (kubectl plugin manager)
```bash
kubectl krew install ksd
```

### Homebrew (macOS/Linux)
```bash
brew install --cask mfuentesg/tap/ksd
```

### Go Install
```bash
go install github.com/mfuentesg/ksd@latest
```

### Download Binary
Download the latest binary for your platform from the [releases page](https://github.com/mfuentesg/ksd/releases).

## Usage

```
$ kubectl get secret <secret name> -o <yaml|json> | ksd
$ ksd < secret.<yaml|json>
```

Installed via Krew? Use `kubectl ksd` in place of `ksd`:

```
$ kubectl get secret <secret name> -o <yaml|json> | kubectl ksd
```

## Example

> secret.json
```json
{
    "apiVersion": "v1",
    "data": {
        "password": "c2VjcmV0",
        "app": "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg=="
    },
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "type": "Opaque"
}
```

```
$ ksd < secret.json
```

> output
```json
{
    "apiVersion": "v1",
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "stringData": {
        "app": "kubernetes secret decoder",
        "password": "secret"
    },
    "type": "Opaque"
}
```

## See also

- [kubectl-view-secret](https://github.com/elsesiy/kubectl-view-secret) — another kubectl plugin for viewing/decoding Kubernetes secrets, also available via Krew.
