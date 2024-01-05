# container-startup-autoscaler
A Helm chart for [container-startup-autoscaler](../../README.md) (CSA).

## Versioning
To promote immutability, a new version of the chart is produced for each new version of CSA. The CSA version that each
chart version ships is specified within the `appVersion` key of [Chart.yaml](Chart.yaml). If you're looking for a
specific version of CSA, `appVersion` is included within the [output](#repository) of `helm search repo` (or see
[CHANGELOG.md](CHANGELOG.md))

Chart versions are maintained separately to CSA versions; the chart version does not indicate CSA version.

## Repository
To add the chart repository, use:

`helm repo add container-startup-autoscaler https://ExpediaGroup.github.io/container-startup-autoscaler`

To locally update versions of the chart, use:

`helm repo update container-startup-autoscaler`

To view available versions of the chart, use:

`helm search repo -l container-startup-autoscaler`

## Configuration
See [values.yaml](values.yaml) for a commented list of all configuration items. 

## Tests
Chart tests are implemented using [helm-unittest](https://github.com/helm-unittest/helm-unittest) - see the
[tests](tests) directory. Unit tests can be run by executing `make test-run-helm` from the CSA root directory. 

## Requirements
- Helm 3.
- See [corresponding CSA version](CHANGELOG.md) for Kubernetes compatibility.
