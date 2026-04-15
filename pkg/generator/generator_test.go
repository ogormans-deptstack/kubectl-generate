package generator

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogormans-deptstack/kubectl-example/pkg/openapi"
)

func TestResourceGeneratorInterface(t *testing.T) {
	t.Run("SupportedTypes returns at least 13 core types", func(t *testing.T) {
		gen := newTestGenerator(t)
		types := gen.SupportedTypes()
		if len(types) < 13 {
			t.Errorf("expected at least 13 supported types, got %d: %v", len(types), types)
		}

		coreTypes := []string{
			"Pod", "Deployment", "Service", "ConfigMap", "Secret",
			"Job", "CronJob", "Ingress", "NetworkPolicy",
			"StatefulSet", "DaemonSet", "PersistentVolumeClaim",
			"HorizontalPodAutoscaler",
		}
		supported := make(map[string]bool)
		for _, st := range types {
			supported[st] = true
		}
		for _, ct := range coreTypes {
			if !supported[ct] {
				t.Errorf("core type %s not in SupportedTypes()", ct)
			}
		}
	})
}

func TestGenerateYAML(t *testing.T) {
	coreTypes := []struct {
		name             string
		requiredFields   []string
		overrides        map[string]string
		expectedContains []string
	}{
		{
			name:           "Pod",
			requiredFields: []string{"apiVersion: v1", "kind: Pod", "metadata:", "spec:", "containers:"},
		},
		{
			name:             "Deployment",
			requiredFields:   []string{"apiVersion: apps/v1", "kind: Deployment", "spec:", "replicas:", "selector:", "template:"},
			overrides:        map[string]string{"replicas": "5", "name": "web"},
			expectedContains: []string{"replicas: 5", "name: web"},
		},
		{
			name:           "Service",
			requiredFields: []string{"apiVersion: v1", "kind: Service", "spec:", "ports:"},
		},
		{
			name:           "ConfigMap",
			requiredFields: []string{"apiVersion: v1", "kind: ConfigMap", "metadata:"},
		},
		{
			name:           "Secret",
			requiredFields: []string{"apiVersion: v1", "kind: Secret", "metadata:"},
		},
		{
			name:           "Job",
			requiredFields: []string{"apiVersion: batch/v1", "kind: Job", "spec:", "template:"},
		},
		{
			name:           "CronJob",
			requiredFields: []string{"apiVersion: batch/v1", "kind: CronJob", "spec:", "schedule:"},
		},
		{
			name:           "Ingress",
			requiredFields: []string{"apiVersion: networking.k8s.io/v1", "kind: Ingress", "spec:", "rules:"},
		},
		{
			name:           "NetworkPolicy",
			requiredFields: []string{"apiVersion: networking.k8s.io/v1", "kind: NetworkPolicy", "spec:", "podSelector:"},
		},
		{
			name:           "StatefulSet",
			requiredFields: []string{"apiVersion: apps/v1", "kind: StatefulSet", "spec:", "serviceName:"},
		},
		{
			name:           "DaemonSet",
			requiredFields: []string{"apiVersion: apps/v1", "kind: DaemonSet", "spec:", "selector:", "template:"},
		},
		{
			name:           "PersistentVolumeClaim",
			requiredFields: []string{"apiVersion: v1", "kind: PersistentVolumeClaim", "spec:", "accessModes:", "resources:"},
		},
		{
			name:           "HorizontalPodAutoscaler",
			requiredFields: []string{"apiVersion: autoscaling/v2", "kind: HorizontalPodAutoscaler", "spec:", "scaleTargetRef:"},
		},
	}

	gen := newTestGenerator(t)

	for _, tc := range coreTypes {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("generates valid YAML with required fields", func(t *testing.T) {
				var buf bytes.Buffer
				overrides := tc.overrides
				if overrides == nil {
					overrides = map[string]string{}
				}
				err := gen.Generate(tc.name, overrides, &buf)
				if err != nil {
					t.Fatalf("Generate(%s) failed: %v", tc.name, err)
				}
				yaml := buf.String()
				for _, field := range tc.requiredFields {
					if !strings.Contains(yaml, field) {
						t.Errorf("YAML for %s missing required field %q\ngot:\n%s", tc.name, field, yaml)
					}
				}
			})

			if tc.overrides != nil {
				t.Run("respects overrides", func(t *testing.T) {
					var buf bytes.Buffer
					err := gen.Generate(tc.name, tc.overrides, &buf)
					if err != nil {
						t.Fatalf("Generate(%s) with overrides failed: %v", tc.name, err)
					}
					yaml := buf.String()
					for _, expected := range tc.expectedContains {
						if !strings.Contains(yaml, expected) {
							t.Errorf("YAML for %s with overrides missing %q\ngot:\n%s", tc.name, expected, yaml)
						}
					}
				})
			}
		})
	}
}

func TestAliasResolution(t *testing.T) {
	gen := newTestGenerator(t)

	aliases := map[string]string{
		"po":     "Pod",
		"deploy": "Deployment",
		"svc":    "Service",
		"cm":     "ConfigMap",
		"cj":     "CronJob",
		"ing":    "Ingress",
		"netpol": "NetworkPolicy",
		"sts":    "StatefulSet",
		"ds":     "DaemonSet",
		"pvc":    "PersistentVolumeClaim",
		"hpa":    "HorizontalPodAutoscaler",
	}

	for alias, expectedKind := range aliases {
		t.Run(alias, func(t *testing.T) {
			var buf bytes.Buffer
			err := gen.Generate(alias, map[string]string{}, &buf)
			if err != nil {
				t.Fatalf("Generate(%s) failed: %v", alias, err)
			}
			if !strings.Contains(buf.String(), "kind: "+expectedKind) {
				t.Errorf("alias %s should generate %s, got:\n%s", alias, expectedKind, buf.String())
			}
		})
	}
}

func newTestGenerator(t *testing.T) ResourceGenerator {
	t.Helper()
	doc := loadMergedFixture(t)
	return NewOpenAPIGenerator(doc)
}

func loadMergedFixture(t *testing.T) *openapi.Document {
	t.Helper()
	fixtureDir := filepath.Join("..", "..", "test", "fixtures", "openapi")
	files, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Skipf("OpenAPI fixtures not found at %s: %v (run tests with a kind cluster first)", fixtureDir, err)
		return nil
	}

	merged := map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{},
		},
	}
	mergedSchemas := merged["components"].(map[string]any)["schemas"].(map[string]any)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(fixtureDir, f.Name()))
		if err != nil {
			t.Fatalf("read fixture %s: %v", f.Name(), err)
		}
		var doc map[string]any
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("parse fixture %s: %v", f.Name(), err)
		}
		components, _ := doc["components"].(map[string]any)
		if components == nil {
			continue
		}
		schemas, _ := components["schemas"].(map[string]any)
		for k, v := range schemas {
			mergedSchemas[k] = v
		}
	}

	data, err := json.Marshal(merged)
	if err != nil {
		t.Fatalf("marshal merged doc: %v", err)
	}
	doc, err := openapi.ParseDocument(data)
	if err != nil {
		t.Fatalf("parse merged doc: %v", err)
	}
	return doc
}
