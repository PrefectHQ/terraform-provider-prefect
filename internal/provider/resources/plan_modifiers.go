package resources

import "encoding/json"

// IsEmptyJSONObject returns true if s parses as a JSON object with no keys.
func IsEmptyJSONObject(s string) bool {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return false
	}

	// obj is nil when s is "null"; we only want actual empty objects.
	return obj != nil && len(obj) == 0
}
