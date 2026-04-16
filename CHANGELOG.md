# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.2.1 (Unreleased)

BUG FIXES:

- Fix `--set` overrides silently overwritten by post-processors ([#3](https://github.com/ogormans-deptstack/kubectl-generate/issues/3)). `applyOverrides()` now runs after all post-processors, and supports dot-path keys (e.g. `--set spec.template.spec.restartPolicy=OnFailure`) and array indexing (e.g. `--set containers[0].image=nginx`).

ENHANCEMENTS:

- Add fuzzy matching for resource type suggestions ([#2](https://github.com/ogormans-deptstack/kubectl-generate/issues/2)). Typos now suggest the closest match using Levenshtein distance (e.g. `Deploymnet` suggests `Deployment`).

## v0.2.0

Released: 2026-04-15

This release renames the project from `kubectl-example` to `kubectl-generate`.

ENHANCEMENTS:

- Rename `kubectl-example` to `kubectl-generate` across all binaries, module path, and documentation
- Expand native resource coverage from 13 to 30 types (RBAC, storage, scheduling, admission, CRDs)
- Add CRD support for Gateway API (10 types), Argo Workflows (4 types), cert-manager (3 types), Crossplane (3 types)
- Add GoReleaser config and krew manifest for distribution
- Add GitHub infrastructure: issue templates, PR template, CODEOWNERS, branch protection via OpenTofu
- Strip mutually exclusive issuer types from cert-manager Issuer/ClusterIssuer
- Strip mutually exclusive Argo template types, fix CronWorkflow schedule field
- Fix krew template indentation for addURIAndSha output

BUG FIXES:

- Fix Argo CRD install in CI (use full CRDs from upstream manifests)
- Fix Gateway API CI (use experimental-install.yaml for full type coverage)
- Fix CronWorkflow validation (use `schedules` plural field, not `schedule`)

## v0.1.0

Released: 2026-04-14

Initial release.

ENHANCEMENTS:

- OpenAPI v3 schema-driven YAML generation from live cluster
- 13 core resource types with server-side dry-run validation
- CRD support (CronTab custom resource)
- Smart field selection: required fields always included, optional fields via important-fields registry
- Sensible defaults: nginx:latest images, RollingUpdate strategy, label/selector wiring
- Override flags: `--name`, `--image`, `--replicas`, `--set key=value`
- Dynamic flag generation from schema introspection
- `--list` to enumerate all available resource types
- CI pipeline with unit tests, lint, and e2e against kind cluster
