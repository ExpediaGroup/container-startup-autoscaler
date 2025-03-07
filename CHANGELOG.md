# Changelog
- Based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
- This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.6.0
2025-??-?? TODO(wt)

### Added
- Ability to supply _either_ CPU or memory scaling configuration through pod annotations (rather than always requiring
  both).
- New `enabledForResources` sub-item within the `scale` status item that indicates which resources are enabled for
  scaling.
- New `failure_configuration` counter metric that shows the number of reconciles where there was a configuration-related
  failure.
- Added additional sandbox scripts to scale upon either one or both of CPU and memory resources.

### Changed
- Multiple reconciler failure metrics collapsed into a single metric with `reason` label.
- Upgrades Go to 1.23.6.
- Upgrades all dependencies.

### Helm Chart
[1.5.0](charts/container-startup-autoscaler/CHANGELOG.md#150)

| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.32     |     ✔️      |                    Alpha                    |
|     1.31     |      ❌      |                    Alpha                    |
|     1.30     |      ❌      |                    Alpha                    |
|     1.29     |      ❌      |                    Alpha                    |
|     1.28     |      ❌      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |

## 0.5.0
2024-12-12

### Added
- Support for Kube 1.32.
  - Container resizes now performed through `resize` subresource.
- Ability to register an additional CA certificate (or chain) when building the kind node image for integration tests
  and sandbox scripts.

### Changed
- Upgrades Go to 1.23.3.
- Upgrades all dependencies.
- Renames controller-runtime controller name to shorten.

### Removed
- Examination of `AllocatedResources` within container status.
  - Not required and now behind feature gate in Kube 1.32.
- Controller name label from CSA metrics.

### Fixed
- Inconsistent status updates through informer cache race.
- CSA metrics not being published.

### Helm Chart
[1.4.0](charts/container-startup-autoscaler/CHANGELOG.md#140)

| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.32     |     ✔️      |                    Alpha                    |
|     1.31     |      ❌      |                    Alpha                    |
|     1.30     |      ❌      |                    Alpha                    |
|     1.29     |      ❌      |                    Alpha                    |
|     1.28     |      ❌      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |

## 0.4.0
2024-11-29

### Changed
- Builds `kind` nodes locally for integration tests and sandbox scripts, instead of using pre-built images.
- Upgrades Go to 1.22.9.

### Helm Chart
[1.3.0](charts/container-startup-autoscaler/CHANGELOG.md#130)

### Kubernetes Compatibility
| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.31     |     ✔️      |                    Alpha                    |
|     1.30     |     ✔️      |                    Alpha                    |
|     1.29     |     ✔️      |                    Alpha                    |
|     1.28     |     ✔️      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |

## 0.3.0
2024-02-01

### Changed
- Some aspects of logging for simplification and consistency purposes. 

### Helm Chart
[1.2.0](charts/container-startup-autoscaler/CHANGELOG.md#120)

### Kubernetes Compatibility
| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.31     |     ✔️      |                    Alpha                    |
|     1.30     |     ✔️      |                    Alpha                    |
|     1.29     |     ✔️      |                    Alpha                    |
|     1.28     |     ✔️      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |

## 0.2.0
2024-02-01

### Removed
- https://github.com/pkg/errors in favor of the `errors` package from the Go standard library.

### Helm Chart
[1.1.0](charts/container-startup-autoscaler/CHANGELOG.md#110)

### Kubernetes Compatibility
| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.29     |     ✔️      |                    Alpha                    |
|     1.28     |     ✔️      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |

## 0.1.0
2024-01-05

### Added
- Initial version.

### Helm Chart
[1.0.0](charts/container-startup-autoscaler/CHANGELOG.md#100)

### Kubernetes Compatibility
| Kube Version | Compatible? | `In-place Update of Pod Resources` Maturity |
|:------------:|:-----------:|:-------------------------------------------:|
|     1.29     |     ✔️      |                    Alpha                    |
|     1.28     |     ✔️      |                    Alpha                    |
|     1.27     |      ❌      |                    Alpha                    |
