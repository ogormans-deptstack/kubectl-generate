package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ogormans-deptstack/kubectl-example/pkg/defaults"
	"github.com/ogormans-deptstack/kubectl-example/pkg/openapi"
)

type OpenAPIGenerator struct {
	doc       *openapi.Document
	gvkIndex  map[string]openapi.GVK
	overrides map[string]string
	inVCT     bool
}

func NewOpenAPIGenerator(doc *openapi.Document) *OpenAPIGenerator {
	return &OpenAPIGenerator{
		doc:      doc,
		gvkIndex: buildGVKIndex(doc),
	}
}

func (g *OpenAPIGenerator) Generate(resourceType string, overrides map[string]string, w io.Writer) error {
	gvk, ok := g.resolveGVK(resourceType)
	if !ok {
		return fmt.Errorf("no example available for %q. Try --list", resourceType)
	}

	schema, err := g.doc.SchemaForGVK(gvk.Group, gvk.Version, gvk.Kind)
	if err != nil {
		return fmt.Errorf("schema lookup failed for %s: %w", gvk.Kind, err)
	}

	g.overrides = overrides
	manifest := g.buildManifest(gvk, schema)

	data, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	yamlOut, err := jsonToYAML(data)
	if err != nil {
		return fmt.Errorf("convert to YAML: %w", err)
	}

	_, err = w.Write(yamlOut)
	return err
}

func (g *OpenAPIGenerator) SupportedTypes() []string {
	var types []string
	seen := make(map[string]bool)
	for _, gvk := range g.gvkIndex {
		if !seen[gvk.Kind] {
			types = append(types, gvk.Kind)
			seen[gvk.Kind] = true
		}
	}
	sort.Strings(types)
	return types
}

func (g *OpenAPIGenerator) buildManifest(gvk openapi.GVK, schema map[string]any) map[string]any {
	manifest := newOrderedMap()

	apiVersion := gvk.Version
	if gvk.Group != "" {
		apiVersion = gvk.Group + "/" + gvk.Version
	}
	manifest.set("apiVersion", apiVersion)
	manifest.set("kind", gvk.Kind)

	name := fmt.Sprintf("example-%s", strings.ToLower(gvk.Kind))
	if override, ok := g.overrides["name"]; ok {
		name = override
	}

	labels := map[string]string{"app.kubernetes.io/name": name}
	manifest.set("metadata", map[string]any{
		"name":   name,
		"labels": labels,
	})

	props, err := openapi.SchemaProperties(g.doc.Raw(), schema)
	if err != nil || props == nil {
		return manifest.toMap()
	}

	specSchema, ok := props["spec"].(map[string]any)
	if !ok {
		g.addDataFields(manifest, gvk, props)
		return manifest.toMap()
	}

	spec := g.walkSchema(specSchema, gvk.Kind, 0)
	if spec != nil {
		g.applyOverrides(spec, gvk.Kind)
		g.injectTemplateLabels(spec, name)
		g.injectTemplateRestartPolicy(spec, gvk.Kind)
		g.fixStrategyDefaults(spec, gvk.Kind)
		g.injectServiceSelector(spec, gvk.Kind, name)
		manifest.set("spec", spec)
	}

	return manifest.toMap()
}

func (g *OpenAPIGenerator) addDataFields(manifest *orderedMap, gvk openapi.GVK, props map[string]any) {
	kind := gvk.Kind
	if _, ok := props["data"]; ok {
		if v := defaults.ValueForField("data", "object", "", kind); v != nil {
			manifest.set("data", v)
		}
	}
	if kind == "Secret" {
		if _, ok := props["type"]; ok {
			manifest.set("type", "Opaque")
		}
	}
}

func (g *OpenAPIGenerator) walkSchema(schema map[string]any, kind string, depth int) map[string]any {
	if depth > 6 {
		return nil
	}

	resolved, err := openapi.ResolveSchema(g.doc.Raw(), schema)
	if err != nil {
		return nil
	}

	props, err := openapi.SchemaProperties(g.doc.Raw(), resolved)
	if err != nil || props == nil {
		return nil
	}

	required := make(map[string]bool)
	for _, r := range openapi.RequiredFields(resolved) {
		required[r] = true
	}

	_, hasSiblingContainers := props["containers"]

	result := make(map[string]any)
	for fieldName, fieldSchema := range props {
		fieldMap, ok := fieldSchema.(map[string]any)
		if !ok {
			continue
		}
		if isExcludedField(fieldName, depth) {
			continue
		}
		if fieldName == "selector" && (kind == "Job" || kind == "CronJob" || kind == "PersistentVolumeClaim" || g.inVCT) {
			continue
		}
		if fieldName == "resources" && hasSiblingContainers {
			continue
		}

		resolvedField, err := openapi.ResolveSchema(g.doc.Raw(), fieldMap)
		if err != nil {
			continue
		}

		isRequired := required[fieldName]
		isImportant := defaults.IsImportantField(fieldName)
		fieldType := openapi.SchemaType(resolvedField)
		isPrimitive := fieldType == "string" || fieldType == "integer" || fieldType == "number" || fieldType == "boolean"

		if !isRequired && !isImportant {
			if depth > 1 {
				continue
			}
			if isPrimitive {
				continue
			}
		}

		val := g.generateValue(fieldName, resolvedField, kind, depth)
		if val != nil {
			result[fieldName] = val
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (g *OpenAPIGenerator) generateValue(fieldName string, schema map[string]any, kind string, depth int) any {
	schemaType := openapi.SchemaType(schema)
	format, _ := schema["format"].(string)

	switch schemaType {
	case "object":
		if v := defaults.ValueForField(fieldName, "object", format, kind); v != nil {
			if m, ok := v.(map[string]any); ok {
				return m
			}
		}
		if isResourceQuantityMap(fieldName, schema) {
			return g.resourceQuantityDefaults(fieldName, kind)
		}
		if hasAdditionalProperties(schema) {
			m := g.mapDefault(fieldName, kind)
			if m == nil {
				return nil
			}
			return m
		}
		sub := g.walkSchema(schema, kind, depth+1)
		if sub == nil {
			return nil
		}
		return sub

	case "array":
		if v := defaults.ValueForField(fieldName, "array", format, kind); v != nil {
			return v
		}
		items := arrayItemSchema(schema)
		if items == nil {
			return nil
		}
		resolved, err := openapi.ResolveSchema(g.doc.Raw(), items)
		if err != nil {
			return nil
		}
		itemType := openapi.SchemaType(resolved)
		if itemType == "object" {
			isVCT := strings.ToLower(fieldName) == "volumeclaimtemplates"
			if isVCT {
				g.inVCT = true
			}
			elem := g.walkSchema(resolved, kind, depth+1)
			if isVCT {
				g.inVCT = false
			}
			if elem != nil {
				return []any{elem}
			}
			return nil
		}
		return nil

	case "string", "integer", "number", "boolean":
		if v, ok := g.overrides[fieldName]; ok {
			if schemaType == "integer" {
				return parseIntOrString(v)
			}
			return v
		}
		if v, ok := defaults.FieldDefault(fieldName, kind); ok {
			return v
		}
		if enums := schemaEnums(schema); len(enums) > 0 {
			return enums[0]
		}
		return defaults.TypeDefault(schemaType, format)
	}
	return nil
}

func (g *OpenAPIGenerator) resourceQuantityDefaults(fieldName, kind string) map[string]any {
	if kind == "PersistentVolumeClaim" || g.inVCT {
		return map[string]any{"storage": "1Gi"}
	}
	lower := strings.ToLower(fieldName)
	switch lower {
	case "limits":
		return map[string]any{"cpu": "500m", "memory": "256Mi"}
	case "requests":
		return map[string]any{"cpu": "250m", "memory": "128Mi"}
	}
	return map[string]any{"cpu": "250m", "memory": "128Mi"}
}

func (g *OpenAPIGenerator) mapDefault(fieldName, kind string) map[string]any {
	if v := defaults.ValueForField(fieldName, "object", "", kind); v != nil {
		if m, ok := v.(map[string]string); ok {
			result := make(map[string]any, len(m))
			for k, v := range m {
				result[k] = v
			}
			return result
		}
	}
	return nil
}

func (g *OpenAPIGenerator) applyOverrides(spec map[string]any, kind string) {
	for k, v := range g.overrides {
		switch k {
		case "name":
			continue
		case "replicas":
			spec["replicas"] = parseIntOrString(v)
		case "image":
			g.setContainerImage(spec, v)
		}
	}
}

func (g *OpenAPIGenerator) setContainerImage(spec map[string]any, image string) {
	if tmpl, ok := spec["template"].(map[string]any); ok {
		if tmplSpec, ok := tmpl["spec"].(map[string]any); ok {
			g.setContainerImage(tmplSpec, image)
			return
		}
	}
	if containers, ok := spec["containers"].([]any); ok && len(containers) > 0 {
		if c, ok := containers[0].(map[string]any); ok {
			c["image"] = image
		}
	}
	if jt, ok := spec["jobTemplate"].(map[string]any); ok {
		if jtSpec, ok := jt["spec"].(map[string]any); ok {
			if tmpl, ok := jtSpec["template"].(map[string]any); ok {
				if tmplSpec, ok := tmpl["spec"].(map[string]any); ok {
					g.setContainerImage(tmplSpec, image)
				}
			}
		}
	}
}

func (g *OpenAPIGenerator) injectTemplateLabels(spec map[string]any, name string) {
	injectLabels := func(tmpl map[string]any) {
		md, ok := tmpl["metadata"].(map[string]any)
		if !ok {
			md = make(map[string]any)
			tmpl["metadata"] = md
		}
		md["labels"] = map[string]any{"app.kubernetes.io/name": name}
	}

	if tmpl, ok := spec["template"].(map[string]any); ok {
		injectLabels(tmpl)
	}
	if jt, ok := spec["jobTemplate"].(map[string]any); ok {
		if jtSpec, ok := jt["spec"].(map[string]any); ok {
			if tmpl, ok := jtSpec["template"].(map[string]any); ok {
				injectLabels(tmpl)
			}
		}
	}
}

func (g *OpenAPIGenerator) injectTemplateRestartPolicy(spec map[string]any, kind string) {
	if kind != "Job" && kind != "CronJob" {
		return
	}
	inject := func(podSpec map[string]any) {
		podSpec["restartPolicy"] = "Never"
	}

	if tmpl, ok := spec["template"].(map[string]any); ok {
		if tmplSpec, ok := tmpl["spec"].(map[string]any); ok {
			inject(tmplSpec)
		}
	}
	if jt, ok := spec["jobTemplate"].(map[string]any); ok {
		if jtSpec, ok := jt["spec"].(map[string]any); ok {
			if tmpl, ok := jtSpec["template"].(map[string]any); ok {
				if tmplSpec, ok := tmpl["spec"].(map[string]any); ok {
					inject(tmplSpec)
				}
			}
		}
	}
}

func (g *OpenAPIGenerator) fixStrategyDefaults(spec map[string]any, kind string) {
	if strategy, ok := spec["strategy"].(map[string]any); ok {
		strategy["type"] = "RollingUpdate"
	}
	if strategy, ok := spec["updateStrategy"].(map[string]any); ok {
		strategy["type"] = "RollingUpdate"
	}
}

func (g *OpenAPIGenerator) injectServiceSelector(spec map[string]any, kind, name string) {
	if kind != "Service" {
		return
	}
	if _, ok := spec["selector"]; ok {
		return
	}
	spec["selector"] = map[string]any{"app.kubernetes.io/name": name}
}

func (g *OpenAPIGenerator) resolveGVK(resourceType string) (openapi.GVK, bool) {
	lower := strings.ToLower(resourceType)
	if gvk, ok := g.gvkIndex[lower]; ok {
		return gvk, true
	}
	if gvk, ok := g.gvkIndex[singularize(lower)]; ok {
		return gvk, true
	}
	return openapi.GVK{}, false
}

func buildGVKIndex(doc *openapi.Document) map[string]openapi.GVK {
	idx := make(map[string]openapi.GVK)
	schemas := doc.ComponentSchemas()
	if schemas == nil {
		return idx
	}

	preferredVersions := map[string]string{
		"Deployment":              "apps/v1",
		"StatefulSet":             "apps/v1",
		"DaemonSet":               "apps/v1",
		"ReplicaSet":              "apps/v1",
		"Job":                     "batch/v1",
		"CronJob":                 "batch/v1",
		"Ingress":                 "networking.k8s.io/v1",
		"NetworkPolicy":           "networking.k8s.io/v1",
		"HorizontalPodAutoscaler": "autoscaling/v2",
	}

	for _, schema := range schemas {
		schemaMap, ok := schema.(map[string]any)
		if !ok {
			continue
		}
		ext, ok := schemaMap["x-kubernetes-group-version-kind"]
		if !ok {
			continue
		}
		arr, ok := ext.([]any)
		if !ok {
			continue
		}
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			gvk := openapi.GVK{
				Group:   stringVal(m, "group"),
				Version: stringVal(m, "version"),
				Kind:    stringVal(m, "kind"),
			}

			lower := strings.ToLower(gvk.Kind)
			existing, exists := idx[lower]
			if exists {
				pref, hasPref := preferredVersions[gvk.Kind]
				gv := gvk.Group + "/" + gvk.Version
				if gvk.Group == "" {
					gv = gvk.Version
				}
				existingGV := existing.Group + "/" + existing.Version
				if existing.Group == "" {
					existingGV = existing.Version
				}
				if hasPref && gv == pref {
					idx[lower] = gvk
				} else if !hasPref && !exists {
					idx[lower] = gvk
				}
				_ = existingGV
			} else {
				idx[lower] = gvk
			}

			for _, alias := range aliasesForKind(gvk.Kind) {
				if _, exists := idx[alias]; !exists {
					idx[alias] = gvk
				}
			}
		}
	}
	return idx
}

func aliasesForKind(kind string) []string {
	aliases := map[string][]string{
		"Pod":                     {"po", "pods"},
		"Deployment":              {"deploy", "deployments"},
		"Service":                 {"svc", "services"},
		"ConfigMap":               {"cm", "configmaps"},
		"Secret":                  {"secrets"},
		"Job":                     {"jobs"},
		"CronJob":                 {"cj", "cronjobs"},
		"Ingress":                 {"ing", "ingresses"},
		"NetworkPolicy":           {"netpol", "networkpolicies"},
		"StatefulSet":             {"sts", "statefulsets"},
		"DaemonSet":               {"ds", "daemonsets"},
		"PersistentVolumeClaim":   {"pvc", "persistentvolumeclaims"},
		"HorizontalPodAutoscaler": {"hpa", "horizontalpodautoscalers"},
	}
	return aliases[kind]
}

func singularize(s string) string {
	irregulars := map[string]string{
		"ingresses":              "ingress",
		"networkpolicies":        "networkpolicy",
		"persistentvolumeclaims": "persistentvolumeclaim",
	}
	if v, ok := irregulars[s]; ok {
		return v
	}
	if strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	return s
}

func isExcludedField(name string, depth int) bool {
	lower := strings.ToLower(name)

	if depth == 0 {
		topLevel := map[string]bool{
			"apiversion": true,
			"kind":       true,
			"metadata":   true,
		}
		if topLevel[lower] {
			return true
		}
	}

	if depth >= 1 {
		deepOnly := map[string]bool{
			"restartpolicy": true,
		}
		if deepOnly[lower] {
			return true
		}
	}

	excluded := map[string]bool{
		"status":            true,
		"managedfields":     true,
		"resourceversion":   true,
		"uid":               true,
		"creationtimestamp": true,
		"deletiontimestamp": true,
		"generation":        true,
		"selflink":          true,
		"finalizers":        true,
		"ownerreferences":   true,
		"clustername":       true,

		"initcontainers":             true,
		"ephemeralcontainers":        true,
		"volumes":                    true,
		"volumemounts":               true,
		"volumedevices":              true,
		"matchexpressions":           true,
		"podfailurepolicy":           true,
		"datasource":                 true,
		"datasourceref":              true,
		"hostaliases":                true,
		"topologyspreadconstraints":  true,
		"overhead":                   true,
		"readinessgates":             true,
		"schedulinggates":            true,
		"resourceclaims":             true,
		"defaultbackend":             true,
		"lifecycle":                  true,
		"livenessprobe":              true,
		"readinessprobe":             true,
		"startupprobe":               true,
		"resizepolicy":               true,
		"env":                        true,
		"securitycontext":            true,
		"os":                         true,
		"affinity":                   true,
		"dnsconfig":                  true,
		"podreplacementpolicy":       true,
		"successfuljobshistorylimit": true,
		"failedjobshistorylimit":     true,
		"concurrencypolicy":          true,
		"suspend":                    true,
		"startingdeadlineseconds":    true,
		"command":                    true,
		"args":                       true,
		"workingdir":                 true,
	}
	return excluded[lower]
}

func isResourceQuantityMap(fieldName string, schema map[string]any) bool {
	lower := strings.ToLower(fieldName)
	if lower == "limits" || lower == "requests" {
		return true
	}
	ap, ok := schema["additionalProperties"].(map[string]any)
	if !ok {
		return false
	}
	ref, _ := ap["$ref"].(string)
	return strings.Contains(ref, "Quantity")
}

func hasAdditionalProperties(schema map[string]any) bool {
	_, ok := schema["additionalProperties"]
	return ok
}

func arrayItemSchema(schema map[string]any) map[string]any {
	items, ok := schema["items"].(map[string]any)
	if ok {
		return items
	}
	return nil
}

func stringVal(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func parseIntOrString(s string) any {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err == nil {
		return i
	}
	return s
}

func schemaEnums(schema map[string]any) []string {
	enums, ok := schema["enum"].([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, e := range enums {
		if s, ok := e.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
