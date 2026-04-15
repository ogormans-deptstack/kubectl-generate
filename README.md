# kubectl-example

Generate example Kubernetes YAML manifests from your cluster's OpenAPI v3 spec.

Instead of copy-pasting from documentation or memorizing resource schemas, `kubectl-example` reads the live OpenAPI spec from your cluster and generates valid, apply-ready YAML for any resource type -- including CRDs.

```
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
```

## Features

- **Dynamic schema-driven generation** -- reads OpenAPI v3 from the connected cluster, so generated YAML always matches the cluster's API version
- **CRD support** -- generates examples for any installed CRD (Gateway API, CronTab, Argo Workflows, etc.)
- **Smart field selection** -- required fields are always included; optional fields are included when they're important for a valid resource (strategy, ports, resources)
- **Sensible defaults** -- `nginx:latest` for images, `RollingUpdate` for strategies, proper label/selector wiring
- **Override flags** -- `--name`, `--image`, `--replicas`, and `--set key=value` for arbitrary fields
- **Pipe-friendly** -- output goes to stdout, ready for `kubectl create -f -`

## Installation

### From source

```bash
git clone https://github.com/ogormans-deptstack/kubectl-example.git
cd kubectl-example
make install
```

This builds the binary and copies it to `$GOPATH/bin/kubectl-example`. kubectl discovers it automatically via the `kubectl-` prefix.

### Verify

```bash
kubectl plugin list
# Should show: kubectl-example

kubectl example --version
```

## Usage

```bash
# Generate a Deployment
kubectl example Deployment

# Generate with overrides
kubectl example Deployment --name=web --image=myapp:v2 --replicas=3

# Generate and apply
kubectl example Service --name=web | kubectl create -f -

# Set arbitrary fields
kubectl example StatefulSet --set serviceName=my-svc

# List all available resource types
kubectl example --list

# Generate a CRD (must be installed in the cluster)
kubectl example CronTab
kubectl example HTTPRoute
```

### Flags

| Flag | Description | Example |
|------|-------------|---------|
| `--name` | Resource name and label value | `--name=web` |
| `--image` | Container image | `--image=nginx:1.25` |
| `--replicas` | Replica count | `--replicas=3` |
| `--set` | Arbitrary field override (repeatable) | `--set serviceName=my-svc` |
| `--kubeconfig` | Path to kubeconfig file | `--kubeconfig=~/.kube/config` |
| `--list` | List all supported resource types | |
| `--version` | Print version | |

The `--set` flag can be repeated and accepts `key=value` pairs. Values containing `=` are handled correctly (`--set annotation=foo=bar` sets `annotation` to `foo=bar`).

Override priority: `--set` overrides take effect after typed flags (`--name`, `--image`, `--replicas`), so `--set name=X` will override `--name=Y`.

## Supported Resources

All resource types in the cluster's OpenAPI v3 spec are supported, including CRDs. Core types tested end-to-end:

| Resource | API Group | Notes |
|----------|-----------|-------|
| Pod | v1 | |
| Deployment | apps/v1 | Includes strategy, selector, labels |
| Service | v1 | ClusterIP default, selector wired |
| ConfigMap | v1 | Example data keys |
| Secret | v1 | Type: Opaque |
| Job | batch/v1 | backoffLimit, restartPolicy |
| CronJob | batch/v1 | Schedule, jobTemplate nested |
| Ingress | networking.k8s.io/v1 | Rules with paths |
| NetworkPolicy | networking.k8s.io/v1 | Ingress/egress rules |
| StatefulSet | apps/v1 | VolumeClaimTemplates, serviceName |
| DaemonSet | apps/v1 | UpdateStrategy |
| PersistentVolumeClaim | v1 | AccessModes, storage request |
| HorizontalPodAutoscaler | autoscaling/v2 | CPU target, scaleTargetRef |

CRDs tested: CronTab (custom), Gateway API (HTTPRoute, Gateway, GatewayClass, GRPCRoute, TCPRoute, TLSRoute, UDPRoute, ReferenceGrant, BackendLBPolicy, BackendTLSPolicy).

## How It Works

1. **Fetch** -- Connects to the cluster via kubeconfig and downloads all OpenAPI v3 group-version schemas using the discovery API
2. **Resolve** -- Finds the schema for the requested resource type by matching Kind (case-insensitive) against the GVK index
3. **Walk** -- Recursively walks the schema tree, selecting fields based on:
   - Required fields (always included)
   - Important fields (ports, containers, resources, strategy, selector, etc.)
   - Depth limits (avoids deeply nested optional structures)
4. **Default** -- Fills values using a priority chain: override flags > field-specific defaults > schema defaults > schema patterns > enum first value > type defaults
5. **Post-process** -- Wires up label/selector consistency, injects restart policies, fixes strategy types, adds service selectors
6. **Emit** -- Marshals to YAML and writes to stdout

## Architecture

```
cmd/kubectl-example/
  main.go              Cobra CLI, flag parsing, override collection

pkg/
  openapi/
    fetcher.go         OpenAPI v3 schema fetcher (FetchAll via discovery API)
    openapi.go         Schema resolution, property extraction, allOf merging

  generator/
    generator.go       ResourceGenerator interface
    openapi_generator.go   Schema walker, field selection, override application
    yaml.go            JSON-to-YAML conversion

  defaults/
    defaults.go        Field and type defaults, important field registry

  flags/
    flags.go           Dynamic flag generation from schema introspection
```

## Schema Evolution

When Kubernetes adds, removes, or changes required fields between versions, the generated YAML adapts automatically because it reads the live schema. Some known considerations:

- **New required fields**: Automatically included in output since the walker always includes required fields
- **Fields becoming optional**: May still appear if they're in the important-fields list
- **Excluded fields**: Some fields are explicitly excluded to produce clean output (status, managedFields, initContainers, probes, etc.). New K8s fields may need to be added to the exclusion list if they cause validation issues.

The exclusion list is in `pkg/generator/openapi_generator.go` in the `isExcludedField` function. To add a new field exclusion:

```go
excluded := map[string]bool{
    // ... existing exclusions
    "newfieldname": true,  // K8s X.Y: description of why excluded
}
```

## Development

```bash
# Run unit tests
make test-unit

# Run e2e tests (requires kind cluster)
make test-e2e

# Run linter
make lint

# Build
make build
```

### Running e2e tests locally

```bash
# Create a kind cluster
kind create cluster --name demo

# Install CRDs for CRD tests
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml
kubectl apply -f test/fixtures/crontab-crd.yaml

# Run e2e
make test-e2e
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
