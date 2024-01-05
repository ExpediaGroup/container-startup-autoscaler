Thank you for contributing to CSA! Please provide all details requested below prior to submitting your pull request
(PR).

# Important Information
- We don't formalize coding standards for this project, but please try to use the existing code as a convention guide.
- CSA and the Helm chart adhere to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) - please be aware of this
  when incrementing versions.
- Releases of CSA and the Helm chart are managed independently of merging to `main` - a maintainer will prepare
  release(s) for you post-merge.

# Areas Affected by PR
Please indicate what areas are affected by your PR (check all that apply):

- [ ] CSA
- [ ] Helm chart

# Nature of PR
Please indicate the nature of your PR (check all that apply):

- [ ] New feature
- [ ] Bug fix
- [ ] Security
- [ ] Other (please provide details within the description)

# Issue Links
Please link to any issues related to your PR, or indicate if not applicable.

# Description
Please provide a description of the change(s) contained within your PR here.

# Checklist
Please ensure you work through any of **applicable** items below prior to submitting your PR:

## Commits
- [ ] Commits are of logical units (to help us work through your changes)

## Tests
- [ ] New unit/integration tests are implemented
- [ ] Existing unit/integration tests are updated
- [ ] All unit/integration tests are passing

## Sandbox Scripts
- [ ] Sandbox scripts are updated

## Docs
- [ ] `README.md` is updated
- [ ] `CHANGELOG.md` is updated

## Helm Chart
- [ ] New tests are implemented
- [ ] Existing tests are updated
- [ ] All tests are passing
- [ ] `version` within `Chart.yaml` is incremented (only if changes made to CSA or the chart itself)
- [ ] `appVersion` within `Chart.yaml` is incremented (only if changes made to CSA)
- [ ] `README.md` is updated
- [ ] `CHANGELOG.md` is updated
