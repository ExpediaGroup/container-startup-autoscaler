# Changelog
- Based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
- This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.8.0
2025-04-29

### Changed
- Enhance informer cache synchronization with configurable conditions.
- Move informer cache synchronization from a poll-based model to an event-based model.

### Helm Chart
[1.7.0](charts/container-startup-autoscaler/CHANGELOG.md#170)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.33        |     ✔️      |                   Beta                   |
|        1.32        |      ❌      |                  Alpha                   |
|        1.31        |      ❌      |                  Alpha                   |
|        1.30        |      ❌      |                  Alpha                   |
|        1.29        |      ❌      |                  Alpha                   |
|        1.28        |      ❌      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.7.0
2025-04-25

### Added
- Support for Kubernetes 1.33.
  - Resize status now examined through pod conditions.
- Additional integration tests.

### Changed
- Pod CSA status to reflect all validation errors.
- Pod CSA status to only update upon a more focused set of events.
- Verbosity level of some log messages.
- Upgrades Go to 1.24.2.
- Upgrades all dependencies.

### Removed
- Ability to perform memory-based scaling since `In-place Update of Pod Resources` now currently [forbids memory downsizing](https://github.com/kubernetes/enhancements/blob/master/keps/sig-node/1287-in-place-update-pod-resources/README.md#memory-limit-decreases).
  Support will be re-introduced when possible.
- State information from pod CSA status (still available in logs).
- `Validation` Kubernetes events.

### Fixed
- Unnecessary status patches when status hasn't changed.
- Duplicate `Scaling` Kubernetes events under certain conditions.

### Helm Chart
[1.6.0](charts/container-startup-autoscaler/CHANGELOG.md#160)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.33        |     ✔️      |                   Beta                   |
|        1.32        |      ❌      |                  Alpha                   |
|        1.31        |      ❌      |                  Alpha                   |
|        1.30        |      ❌      |                  Alpha                   |
|        1.29        |      ❌      |                  Alpha                   |
|        1.28        |      ❌      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.6.0
2025-03-07

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

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.32        |     ✔️      |                  Alpha                   |
|        1.31        |      ❌      |                  Alpha                   |
|        1.30        |      ❌      |                  Alpha                   |
|        1.29        |      ❌      |                  Alpha                   |
|        1.28        |      ❌      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.5.0
2024-12-12

### Added
- Support for Kubernetes 1.32.
  - Container resizes now performed through `resize` subresource.
- Ability to register an additional CA certificate (or chain) when building the kind node image for integration tests
  and sandbox scripts.

### Changed
- Upgrades Go to 1.23.3.
- Upgrades all dependencies.
- Renames controller-runtime controller name to shorten.

### Removed
- Examination of `AllocatedResources` within container status.
  - Not required and now behind a feature gate in Kubernetes 1.32.
- Controller name label from CSA metrics.

### Fixed
- Inconsistent status updates through informer cache race.
- CSA metrics not being published.

### Helm Chart
[1.4.0](charts/container-startup-autoscaler/CHANGELOG.md#140)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.32        |     ✔️      |                  Alpha                   |
|        1.31        |      ❌      |                  Alpha                   |
|        1.30        |      ❌      |                  Alpha                   |
|        1.29        |      ❌      |                  Alpha                   |
|        1.28        |      ❌      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.4.0
2024-11-29

### Changed
- Builds `kind` nodes locally for integration tests and sandbox scripts, instead of using pre-built images.
- Upgrades Go to 1.22.9.

### Helm Chart
[1.3.0](charts/container-startup-autoscaler/CHANGELOG.md#130)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.31        |     ✔️      |                  Alpha                   |
|        1.30        |     ✔️      |                  Alpha                   |
|        1.29        |     ✔️      |                  Alpha                   |
|        1.28        |     ✔️      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.3.0
2024-02-01

### Changed
- Some aspects of logging for simplification and consistency purposes. 

### Helm Chart
[1.2.0](charts/container-startup-autoscaler/CHANGELOG.md#120)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.31        |     ✔️      |                  Alpha                   |
|        1.30        |     ✔️      |                  Alpha                   |
|        1.29        |     ✔️      |                  Alpha                   |
|        1.28        |     ✔️      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.2.0
2024-02-01

### Removed
- https://github.com/pkg/errors in favor of the `errors` package from the Go standard library.

### Helm Chart
[1.1.0](charts/container-startup-autoscaler/CHANGELOG.md#110)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.29        |     ✔️      |                  Alpha                   |
|        1.28        |     ✔️      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |

## 0.1.0
2024-01-05

### Added
- Initial version.

### Helm Chart
[1.0.0](charts/container-startup-autoscaler/CHANGELOG.md#100)

### Kubernetes Compatibility
| Kubernetes Version | Compatible? | `In-place Update of Pod Resources` State |
|:------------------:|:-----------:|:----------------------------------------:|
|        1.29        |     ✔️      |                  Alpha                   |
|        1.28        |     ✔️      |                  Alpha                   |
|        1.27        |      ❌      |                  Alpha                   |
