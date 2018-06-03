# Kubernetes secret decoder a.k.a `ksd`

`ksd` it is a useful tool to decode kubernetes secrets, thought to work with pipes.
`ksd` supports `yaml` and `json` files.

## Usage

```
$ kubectl get secret <secret name> -o <yaml|json> | ksd
$ ksd < kubectl get secret <secret name> <secret file>.<yaml|json>
```

## Example

```json
kube_secret.json

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

```json
output

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
