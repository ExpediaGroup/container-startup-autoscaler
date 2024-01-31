# Changelog
- Based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
- This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
