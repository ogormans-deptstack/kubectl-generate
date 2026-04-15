package defaults

import (
	"fmt"
	"strings"
)

func ValueForField(fieldName, schemaType, format, kind string) any {
	if v, ok := fieldNameDefaults(fieldName, kind); ok {
		return v
	}
	return typeDefault(schemaType, format)
}

func FieldDefault(fieldName, kind string) (any, bool) {
	return fieldNameDefaults(fieldName, kind)
}

func TypeDefault(schemaType, format string) any {
	return typeDefault(schemaType, format)
}

func fieldNameDefaults(field, kind string) (any, bool) {
	lower := strings.ToLower(field)
	kindLower := strings.ToLower(kind)

	switch lower {
	case "name":
		return fmt.Sprintf("example-%s", kindLower), true
	case "image":
		return "nginx:latest", true
	case "replicas":
		return 3, true
	case "schedule":
		return "*/5 * * * *", true
	case "restartpolicy":
		if kindLower == "job" || kindLower == "cronjob" {
			return "Never", true
		}
		return "Always", true
	case "containerport", "port", "targetport", "number":
		return 80, true
	case "protocol":
		return "TCP", true
	case "type":
		if kindLower == "service" {
			return "ClusterIP", true
		}
		if kindLower == "secret" {
			return "Opaque", true
		}
		return nil, false
	case "accessmodes":
		return []string{"ReadWriteOnce"}, true
	case "storage":
		return "1Gi", true
	case "cpu":
		return "250m", true
	case "memory":
		return "128Mi", true
	case "path":
		return "/", true
	case "pathtype":
		return "Prefix", true
	case "host":
		return "example.com", true
	case "servicename":
		return fmt.Sprintf("example-%s", kindLower), true
	case "minreplicas":
		return 1, true
	case "maxreplicas":
		return 10, true
	case "matchlabels":
		return map[string]string{"app.kubernetes.io/name": fmt.Sprintf("example-%s", kindLower)}, true
	case "metrics":
		if kindLower == "horizontalpodautoscaler" {
			return []any{
				map[string]any{
					"type": "Resource",
					"resource": map[string]any{
						"name": "cpu",
						"target": map[string]any{
							"type":               "Utilization",
							"averageUtilization": 80,
						},
					},
				},
			}, true
		}
		return nil, false
	case "scaletargetref":
		if kindLower == "horizontalpodautoscaler" {
			return map[string]any{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"name":       fmt.Sprintf("example-%s", kindLower),
			}, true
		}
		return nil, false
	case "data":
		if kindLower == "configmap" {
			return map[string]string{"key": "value"}, true
		}
		if kindLower == "secret" {
			return map[string]string{"username": "YWRtaW4=", "password": "cGFzc3dvcmQ="}, true
		}
		return nil, false
	}
	return nil, false
}

func typeDefault(schemaType, format string) any {
	switch schemaType {
	case "string":
		if format == "date-time" {
			return "2024-01-01T00:00:00Z"
		}
		return "example"
	case "integer":
		if format == "int64" {
			return 1
		}
		return 1
	case "boolean":
		return false
	case "number":
		return 1.0
	}
	return nil
}

var importantFields = map[string]bool{
	"containers":           true,
	"selector":             true,
	"template":             true,
	"ports":                true,
	"rules":                true,
	"schedule":             true,
	"jobtemplate":          true,
	"podselector":          true,
	"ingress":              true,
	"egress":               true,
	"servicename":          true,
	"scaletargetref":       true,
	"metrics":              true,
	"resources":            true,
	"limits":               true,
	"requests":             true,
	"data":                 true,
	"type":                 true,
	"accessmodes":          true,
	"matchlabels":          true,
	"minreplicas":          true,
	"maxreplicas":          true,
	"replicas":             true,
	"volumeclaimtemplates": true,
	"image":                true,
	"http":                 true,
	"paths":                true,
	"pathtype":             true,
	"backend":              true,
	"service":              true,
	"spec":                 true,
	"path":                 true,
	"port":                 true,
	"number":               true,
	"host":                 true,
	"restartpolicy":        true,
}

func IsImportantField(fieldName string) bool {
	return importantFields[strings.ToLower(fieldName)]
}
