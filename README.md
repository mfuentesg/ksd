# Kubernetes secret decoder a.k.a `ksd`


[![codecov](https://codecov.io/gh/mfuentesg/ksd/branch/main/graph/badge.svg)](https://codecov.io/gh/mfuentesg/ksd)

<a href="https://www.buymeacoffee.com/mfuentesg" target="_blank">
   <img height="41" src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" />
</a>

<br />
<br />

`ksd` is a tool, whose aim is help you to visualize in text plain your kubernetes secrets, either `yaml` or `json` outputs.

## Installation

### Krew (kubectl plugin)
```bash
kubectl krew install ksd
```

### Homebrew (macOS/Linux)
```bash
brew install mfuentesg/tap/ksd
```

### Go Install
```bash
go install github.com/mfuentesg/ksd@latest
```

### Download Binary
Download the latest binary from the [releases page](https://github.com/mfuentesg/ksd/releases).

## Usage

### As a kubectl plugin (after installing via Krew)
```bash
kubectl get secret <secret-name> -o yaml | kubectl ksd
kubectl get secret <secret-name> -o json | kubectl ksd
```

### As a standalone tool
```bash
kubectl get secret <secret-name> -o yaml | ksd
kubectl get secret <secret-name> -o json | ksd
ksd < secret.yaml
ksd < secret.json
```

## Example

> kube_secret.json
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
$ ksd < kube_secret.json
```

> output
```json
{
    "apiVersion": "v1",
    "data": {
        "password": "secret",
        "app": "kubernetes secret decoder"
    },
    "kind": "Secret",
    "metadata": {
        "name": "kubernetes secret decoder",
        "namespace": "ksd"
    },
    "type": "Opaque"
}
```
