package configa

import (
	"fmt"
	"maps"

	"gopkg.in/yaml.v3"
)

// mergeYAML deep-merges overlay onto base and returns the combined YAML bytes.
// For map keys that exist in both files, the merge recurses into nested maps.
// For scalars and arrays, the overlay value wins entirely.
func mergeYAML(base, overlay []byte) ([]byte, error) {
	var baseMap, overlayMap map[string]any
	if err := yaml.Unmarshal(base, &baseMap); err != nil {
		return nil, fmt.Errorf("configa: parse base yaml for merge: %w", err)
	}
	if err := yaml.Unmarshal(overlay, &overlayMap); err != nil {
		return nil, fmt.Errorf("configa: parse overlay yaml for merge: %w", err)
	}
	if baseMap == nil {
		baseMap = make(map[string]any)
	}
	merged := deepMerge(baseMap, overlayMap)
	out, err := yaml.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("configa: marshal merged yaml: %w", err)
	}
	return out, nil
}

// deepMerge returns a new map that is overlay applied on top of base.
// When both values are maps, they are merged recursively.
// In all other cases the overlay value replaces the base value.
func deepMerge(base, overlay map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	maps.Copy(result, base)
	for k, ov := range overlay {
		if bv, exists := result[k]; exists {
			bMap, bIsMap := bv.(map[string]any)
			oMap, oIsMap := ov.(map[string]any)
			if bIsMap && oIsMap {
				result[k] = deepMerge(bMap, oMap)
				continue
			}
		}
		result[k] = ov
	}
	return result
}
