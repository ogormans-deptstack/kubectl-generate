package openapi

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/openapi"
	cachedopenapi "k8s.io/client-go/openapi/cached"
)

type SchemaFetcher struct {
	client openapi.Client
}

func NewSchemaFetcher(client openapi.Client) *SchemaFetcher {
	return &SchemaFetcher{client: cachedopenapi.NewClient(client)}
}

func (f *SchemaFetcher) FetchSchema(gvk schema.GroupVersionKind) (*Document, map[string]any, error) {
	paths, err := f.client.Paths()
	if err != nil {
		return nil, nil, fmt.Errorf("fetch OpenAPI paths: %w", err)
	}

	resourcePath := resourcePathFromGV(gvk.GroupVersion())
	gv, ok := paths[resourcePath]
	if !ok {
		return nil, nil, fmt.Errorf("no OpenAPI schema for path %s", resourcePath)
	}

	data, err := gv.Schema(runtime.ContentTypeJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch schema for %s: %w", resourcePath, err)
	}

	doc, err := ParseDocument(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse schema for %s: %w", resourcePath, err)
	}

	resourceSchema, err := doc.SchemaForGVK(gvk.Group, gvk.Version, gvk.Kind)
	if err != nil {
		return nil, nil, err
	}

	return doc, resourceSchema, nil
}

func (f *SchemaFetcher) ListGVKs() ([]GVK, error) {
	paths, err := f.client.Paths()
	if err != nil {
		return nil, fmt.Errorf("fetch OpenAPI paths: %w", err)
	}

	var allGVKs []GVK
	for _, gv := range paths {
		data, err := gv.Schema(runtime.ContentTypeJSON)
		if err != nil {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		components, _ := raw["components"].(map[string]any)
		if components == nil {
			continue
		}
		schemas, _ := components["schemas"].(map[string]any)
		for _, s := range schemas {
			schemaMap, ok := s.(map[string]any)
			if !ok {
				continue
			}
			allGVKs = append(allGVKs, extractGVKs(schemaMap)...)
		}
	}
	return allGVKs, nil
}

func (f *SchemaFetcher) FetchAll() (*Document, error) {
	paths, err := f.client.Paths()
	if err != nil {
		return nil, fmt.Errorf("fetch OpenAPI paths: %w", err)
	}

	merged := map[string]any{
		"components": map[string]any{
			"schemas": map[string]any{},
		},
	}
	mergedSchemas := merged["components"].(map[string]any)["schemas"].(map[string]any)

	for _, gv := range paths {
		data, err := gv.Schema(runtime.ContentTypeJSON)
		if err != nil {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		components, _ := raw["components"].(map[string]any)
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
		return nil, fmt.Errorf("marshal merged schemas: %w", err)
	}
	return ParseDocument(data)
}

func resourcePathFromGV(gv schema.GroupVersion) string {
	if len(gv.Group) == 0 {
		return fmt.Sprintf("api/%s", gv.Version)
	}
	return fmt.Sprintf("apis/%s/%s", gv.Group, gv.Version)
}
