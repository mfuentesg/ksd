---
name: Krew Plugin Submission
about: Track submission of ksd to the Krew plugin index
title: 'Submit ksd to Krew plugin index'
labels: ['krew', 'plugin', 'enhancement']
assignees: ''

---

## Krew Plugin Submission Checklist

This issue tracks the submission of `ksd` to the official Krew plugin index.

### Prerequisites
- [ ] Latest release has been created with GoReleaser
- [ ] Krew manifest has been generated in `dist/krew/ksd.yaml`
- [ ] Plugin has been tested locally with `make krew-test`

### Submission Steps
- [ ] Fork the [krew-index repository](https://github.com/kubernetes-sigs/krew-index)
- [ ] Copy `dist/krew/ksd.yaml` to `plugins/ksd.yaml` in the forked repository
- [ ] Create a pull request to kubernetes-sigs/krew-index
- [ ] Address any review feedback
- [ ] Wait for PR approval and merge

### Testing
- [ ] Test installation: `kubectl krew install --manifest=dist/krew/ksd.yaml`
- [ ] Test functionality: `kubectl ksd --version`
- [ ] Test with real secret: `kubectl get secret <name> -o yaml | kubectl ksd`
- [ ] Clean up: `kubectl krew uninstall ksd`

### Links
- [ ] Link to krew-index PR: <!-- Add PR link here -->
- [ ] Link to release: <!-- Add release link here -->

### Notes
<!-- Add any additional notes or considerations here -->