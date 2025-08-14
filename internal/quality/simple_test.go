package quality

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"gopkg.in/yaml.v3"
)

// findRepoRoot walks up from the current directory to locate the module root (where go.mod lives).
func findRepoRoot(tb testing.TB) string {
	tb.Helper()
	
	start, err := os.Getwd()
	if err != nil {
		tb.Fatalf("os.Getwd(): %v", err)
	}
	
	dir := start
	for i := 0; i < 20; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	tb.Fatalf("could not locate module root (no go.mod found) starting from %s", start)
	return ""
}

// readYAML loads a YAML file from the repository root into a generic map.
func readYAML(tb testing.TB, relPath string) map[string]any {
	tb.Helper()
	
	root := findRepoRoot(tb)
	data, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		tb.Fatalf("read %s: %v", relPath, err)
	}
	
	var out map[string]any
	if err := yaml.Unmarshal(data, &out); err != nil {
		tb.Fatalf("yaml.Unmarshal(%s): %v", relPath, err)
	}
	return out
}

func TestSwaggerYAML_BasicShape(t *testing.T) {
	spec := readYAML(t, "swagger.yaml")
	
	ov, ok := spec["openapi"].(string)
	if !ok || ov == "" {
		t.Fatalf("missing or invalid 'openapi' version")
	}
	if !strings.HasPrefix(ov, "3") {
		t.Fatalf("expected OpenAPI v3.x, got %q", ov)
	}
	
	info, ok := spec["info"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'info' object")
	}
	if title, _ := info["title"].(string); strings.TrimSpace(title) == "" {
		t.Fatalf("missing or empty 'info.title'")
	}
	
	paths, ok := spec["paths"].(map[string]any)
	if !ok || len(paths) == 0 {
		t.Fatalf("missing or empty 'paths' object")
	}
}

func TestSwaggerYAML_NoDuplicateOperationIDs(t *testing.T) {
	spec := readYAML(t, "swagger.yaml")
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'paths'")
	}
	
	seen := make(map[string]string) // operationId -> path+method
	var found int
	
	methods := map[string]struct{}{
		"get": {}, "put": {}, "post": {}, "delete": {}, "patch": {}, "options": {}, "head": {}, "trace": {},
	}
	
	for p, raw := range paths {
		pm, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		for m, vr := range pm {
			if _, ok := methods[strings.ToLower(m)]; !ok {
				continue
			}
			op, ok := vr.(map[string]any)
			if !ok {
				continue
			}
			if id, ok := op["operationId"].(string); ok && id != "" {
				found++
				key := strings.ToLower(p) + " " + strings.ToUpper(m)
				if prev, exists := seen[id]; exists {
					t.Fatalf("duplicate operationId %q found at %s; previously seen at %s", id, key, prev)
				}
				seen[id] = key
			}
		}
	}
	
	if found == 0 {
		t.Skip("no operationId fields found; skipping duplicate check")
	}
}

func TestSwaggerYAML_InternalRefsResolvable(t *testing.T) {
	spec := readYAML(t, "swagger.yaml")
	
	var refs []string
	//collectRefs := func(v any) {}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, val := range x {
				if k == "$ref" {
					if s, ok := val.(string); ok && strings.HasPrefix(s, "#/") {
						refs = append(refs, s)
					}
				}
				walk(val)
			}
		case []any:
			for _, it := range x {
				walk(it)
			}
		}
	}
	walk(spec)
	
	if len(refs) == 0 {
		t.Skip("no internal $ref found; skipping resolution check")
	}
	
	// Resolve internal JSON pointers like "#/components/schemas/Thing"
	for _, r := range refs {
		ptr := strings.TrimPrefix(r, "#")
		if !strings.HasPrefix(ptr, "/") {
			t.Fatalf("malformed $ref %q (expected JSON pointer starting with '/')", r)
		}
		if _, ok := getByJSONPointer(spec, ptr); !ok {
			t.Fatalf("unresolvable $ref: %s", r)
		}
	}
}

// JSON pointer resolution for maps/slices per RFC 6901.
func getByJSONPointer(root any, pointer string) (any, bool) {
	cur := root
	parts := strings.Split(pointer, "/")
	// First element is always empty due to leading "/"
	for i := 1; i < len(parts); i++ {
		raw := parts[i]
		// Unescape
		token := strings.ReplaceAll(strings.ReplaceAll(raw, "~1", "/"), "~0", "~")
		
		switch node := cur.(type) {
		case map[string]any:
			next, ok := node[token]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			// arrays not commonly used in OpenAPI pointers, but handle for completeness
			// attempt to parse numeric index
			//idx := -1
			if token == "-" {
				return nil, false
			}
			for j := 0; j < len(node); j++ {
				// best-effort: OpenAPI refs usually don't use array indices; bail out
				_ = j
			}
			// If pointer requests an index we don't support, report missing
			return nil, false
		default:
			return nil, false
		}
	}
	return cur, true
}
