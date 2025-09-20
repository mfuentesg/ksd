# Krew Plugin Support

This directory contains files related to Krew (kubectl plugin manager) support.

## How to submit to Krew

1. **Create a release** - Push a tag to trigger GoReleaser which will generate the Krew manifest
2. **Fork krew-index** - Fork the [krew-index repository](https://github.com/kubernetes-sigs/krew-index)
3. **Add plugin manifest** - Copy the generated manifest from `dist/krew/ksd.yaml` to `plugins/ksd.yaml` in your fork
4. **Submit PR** - Create a pull request to the krew-index repository

## Generated files

When you run `goreleaser release`, it will generate:
- `dist/krew/ksd.yaml` - The Krew plugin manifest

## Testing locally

You can test the plugin locally before submitting:

```bash
# Install from local manifest
kubectl krew install --manifest=dist/krew/ksd.yaml

# Test the plugin
kubectl ksd --version

# Uninstall
kubectl krew uninstall ksd
```

## Plugin naming

The plugin will be available as `kubectl ksd` after installation via Krew.