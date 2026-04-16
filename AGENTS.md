# kubectl-example -- Agent Context

## Project

kubectl plugin that generates example YAML from OpenAPI v3 specs. Go module at `github.com/ogormans-deptstack/kubectl-example`. Apache 2.0.

## Conventions

- Go 1.25+, strict linting via golangci-lint v2
- Table-driven tests, e2e against kind cluster with server-side dry-run
- Factory pattern: `ensureXXXCRD()` helpers for CRD test groups
- No `as any` / `@ts-ignore` equivalents -- no lint suppression
- Commit messages: no `Fixes #N` or auto-close keywords (Prow flags these)
- Use SSH for git push: `git@github.com:ogormans-deptstack/kubectl-example.git`

## Key Paths

| What | Where |
|------|-------|
| CLI entry | `cmd/kubectl-example/main.go` |
| Generator | `pkg/generator/openapi_generator.go` |
| YAML emitter | `pkg/generator/yaml.go` |
| Schema fetcher | `pkg/openapi/fetcher.go` |
| Defaults | `pkg/defaults/defaults.go` |
| Flags | `pkg/flags/flags.go` |
| E2e tests | `test/e2e/e2e_test.go` |
| CI | `.github/workflows/ci.yml` |
| CronTab CRD fixture | `test/fixtures/crontab-crd.yaml` |

## sig-cli Engagement

- **KEP**: kubernetes/enhancements#5576 (PR), tracking issue #5571
- **Meeting**: Presented at sig-cli bi-weekly, April 2026. Positive reception.
- **Contacts**: ardaguclu (sig-cli member, invited to meeting), soltysh (sig-cli, expressed interest), eddiezane (sig-cli lead, KEP approver)
- **Target**: v1.37 alpha (`kubectl alpha example`), earliest Enhancements Freeze ~May 2026

## Banked: Issue #5571 Response Draft

**Target posting date: week of April 21-25, 2026**

```markdown
/reopen
/remove-lifecycle rotten

Following up from the sig-cli bi-weekly discussion on April 16 -- thanks @ardaguclu for the invite and the feedback.

Working prototype is at https://github.com/ogormans-deptstack/kubectl-example (v0.1.0 released, available via krew). It reads the cluster's OpenAPI v3 spec and generates apply-ready YAML for any resource type, including CRDs.

**How it works:**

The plugin connects via the discovery API, fetches all OpenAPI v3 group-version schemas, and walks the schema tree to produce a minimal valid manifest. Required fields are always included; important optional fields (strategy, ports, resources, selectors) are pulled in via heuristics. Labels, selectors, and template metadata are wired up consistently.

**Demo output:**

$ kubectl example Deployment --name=web --image=myapp:v2 --replicas=5
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  labels:
    app.kubernetes.io/name: web
spec:
  replicas: 5
  selector:
    matchLabels:
      app.kubernetes.io/name: web
  strategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        app.kubernetes.io/name: web
    spec:
      containers:
        - name: web
          image: "myapp:v2"
          ports:
            - containerPort: 80
          resources:
            limits:
              cpu: "500m"
              memory: "256Mi"
            requests:
              cpu: "250m"
              memory: "128Mi"

**Current state (v0.1.0):**

- 30 core resource types pass server-side dry-run validation (Pod, Deployment, Service, ConfigMap, Secret, Job, CronJob, Ingress, NetworkPolicy, StatefulSet, DaemonSet, PVC, HPA, Role, ClusterRole, RoleBinding, ClusterRoleBinding, ServiceAccount, Namespace, ResourceQuota, LimitRange, PV, PDB, IngressClass, StorageClass, PriorityClass, RuntimeClass, ValidatingWebhookConfiguration, MutatingWebhookConfiguration, CRD)
- CRD support validated against: CronTab, 10 Gateway API types, 4 Argo Workflows types, 3 cert-manager types, 3 Crossplane types
- Override flags: `--name`, `--image`, `--replicas`, `--set key=value`
- ~266 unit tests, e2e suite against kind cluster
- CI with golangci-lint v2, Go 1.25/1.26 matrix, e2e on kind v1.33.0
- Distributed via krew (`kubectl krew install example`) and GitHub releases

**Next steps toward v1.37 alpha:**

1. Updating KEP #5576 to retarget v1.37 with the current prototype as evidence of feasibility
2. CLI polish: descriptive error messages for invalid types/flags, fuzzy matching suggestions (tracked in v0.1.1 milestone)
3. Happy to iterate on any design feedback from the meeting or this thread

cc @soltysh @eddiezane
```

**Before posting, verify:**
- [ ] Krew submission PR merged (check krew-index)
- [ ] CI green on latest main
- [ ] Re-read for AI tells -- must read as a human engineer wrote it

## Release Plan

- GoReleaser for cross-platform builds (darwin/linux/windows, amd64/arm64)
- Krew manifest in repo root
- GitHub Actions release workflow triggered on tag push (v*)
- First release: v0.1.0

## CRD Test Expansion

| CRD Group | Install Method | Key Kinds | Status |
|-----------|---------------|-----------|--------|
| CronTab | local fixture | CronTab | done |
| Gateway API | remote URL (v1.2.1) | HTTPRoute, Gateway, +8 | done |
| Argo Workflows | kustomize (minimal CRDs) | Workflow, CronWorkflow, WorkflowTemplate, ClusterWorkflowTemplate | done |
| Crossplane | remote URL | Composition, CompositeResourceDefinition, EnvironmentConfig | done |
| cert-manager | remote URL | Certificate, Issuer, ClusterIssuer | done |
