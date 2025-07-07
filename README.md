# Kubernetes secret decoder a.k.a `ksd`


[![codecov](https://codecov.io/gh/mfuentesg/ksd/branch/main/graph/badge.svg)](https://codecov.io/gh/mfuentesg/ksd)

<a href="https://www.buymeacoffee.com/mfuentesg" target="_blank">
   <img height="41" src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" />
</a>

<br />
<br />

`ksd` is a tool, whose aim is help you to visualize in text plain your kubernetes secrets, either `yaml` or `json` outputs.

## Installation

### Go
```bash
$ go get github.com/mfuentesg/ksd
```

### Brew

```
brew install mfuentesg/tap/ksd
```

## Usage

```
$ kubectl get secret <secret name> -o <yaml|json> | ksd
$ ksd < kubectl get secret <secret name> <secret file>.<yaml|json>
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
