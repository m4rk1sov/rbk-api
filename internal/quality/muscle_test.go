package quality

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSimilarMusclesJSON_ParsesAndNonEmpty(t *testing.T) {
	root := findRepoRoot(t)

	data, err := os.ReadFile(filepath.Join(root, "similar_muscles.json"))
	if err != nil {
		t.Fatalf("read similar_muscles.json: %v", err)
	}

	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	switch x := v.(type) {
	case map[string]any:
		if len(x) == 0 {
			t.Fatalf("similar_muscles.json parsed as object but is empty")
		}
	case []any:
		if len(x) == 0 {
			t.Fatalf("similar_muscles.json parsed as array but is empty")
		}
	default:
		t.Fatalf("unexpected JSON top-level type: %T", v)
	}
}
