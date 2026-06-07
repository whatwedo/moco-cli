package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Spec is the minimal model of the MOCO OpenAPI spec needed by the generator.
type Spec struct {
	Paths      map[string]*PathItem `yaml:"paths"`
	Components struct {
		Schemas map[string]*Schema `yaml:"schemas"`
	} `yaml:"components"`
}

// PathItem bundles the operations of a single path.
type PathItem struct {
	Parameters []*Parameter `yaml:"parameters"`
	Get        *Operation   `yaml:"get"`
	Post       *Operation   `yaml:"post"`
	Put        *Operation   `yaml:"put"`
	Patch      *Operation   `yaml:"patch"`
	Delete     *Operation   `yaml:"delete"`
}

// MethodOp pairs an HTTP method with its operation.
type MethodOp struct {
	Method string
	Op     *Operation
}

// operations returns the method/operation pairs in a stable order.
func (p *PathItem) operations() []MethodOp {
	var out []MethodOp
	for _, m := range []MethodOp{
		{"GET", p.Get}, {"POST", p.Post}, {"PUT", p.Put}, {"PATCH", p.Patch}, {"DELETE", p.Delete},
	} {
		if m.Op != nil {
			out = append(out, m)
		}
	}
	return out
}

// Operation describes a single API operation.
type Operation struct {
	Tags        []string     `yaml:"tags"`
	Summary     string       `yaml:"summary"`
	Parameters  []*Parameter `yaml:"parameters"`
	RequestBody *RequestBody `yaml:"requestBody"`
}

// Parameter describes a path, query or header parameter.
type Parameter struct {
	Ref         string  `yaml:"$ref"`
	In          string  `yaml:"in"`
	Name        string  `yaml:"name"`
	Required    bool    `yaml:"required"`
	Description string  `yaml:"description"`
	Schema      *Schema `yaml:"schema"`
}

// RequestBody describes the request body of an operation.
type RequestBody struct {
	Required bool                  `yaml:"required"`
	Content  map[string]*MediaType `yaml:"content"`
}

// MediaType references the schema of a content type.
type MediaType struct {
	Schema *Schema `yaml:"schema"`
}

// Schema is a simplified JSON schema.
type Schema struct {
	Ref         string             `yaml:"$ref"`
	Type        string             `yaml:"type"`
	Format      string             `yaml:"format"`
	Description string             `yaml:"description"`
	Properties  map[string]*Schema `yaml:"properties"`
	Required    []string           `yaml:"required"`
	Items       *Schema            `yaml:"items"`
}

// LoadSpec reads and parses the OpenAPI spec from a local path or an https URL.
// The MOCO spec is MOCO's own work and is fetched at generation time rather than
// vendored into this repository.
func LoadSpec(src string) (*Spec, error) {
	data, err := readSource(src)
	if err != nil {
		return nil, err
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("could not parse spec: %w", err)
	}
	return &spec, nil
}

// readSource returns the spec bytes from an https URL or a local file path.
func readSource(src string) ([]byte, error) {
	if strings.HasPrefix(src, "https://") {
		c := &http.Client{Timeout: 60 * time.Second}
		resp, err := c.Get(src)
		if err != nil {
			return nil, fmt.Errorf("could not fetch spec from %s: %w", src, err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("could not fetch spec from %s: %s", src, resp.Status)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(src)
}

// resolve follows a $ref of the form "#/Name" (a MOCO quirk) or
// "#/components/schemas/Name" to the referenced schema.
func (s *Spec) resolve(schema *Schema) *Schema {
	seen := map[string]bool{}
	for schema != nil && schema.Ref != "" {
		name := refName(schema.Ref)
		if name == "" || seen[name] {
			return schema
		}
		seen[name] = true
		schema = s.Components.Schemas[name]
	}
	return schema
}

// refName extracts the schema name from a $ref pointer.
func refName(ref string) string {
	ref = strings.TrimPrefix(ref, "#/")
	ref = strings.TrimPrefix(ref, "components/schemas/")
	return ref
}

// sortedSchemaProps returns the property names of a resolved body schema in a
// stable (alphabetical) order.
func (s *Spec) sortedSchemaProps(schema *Schema) []string {
	schema = s.resolve(schema)
	if schema == nil {
		return nil
	}
	names := make([]string, 0, len(schema.Properties))
	for name := range schema.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
