# Kubernetes secret decoder a.k.a `ksd`


![Travis](https://img.shields.io/travis/mfuentesg/ksd.svg)
[![codecov](https://codecov.io/gh/mfuentesg/ksd/branch/master/graph/badge.svg)](https://codecov.io/gh/mfuentesg/ksd)

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
