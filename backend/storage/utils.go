package storage

import (
	"encoding/json"
)

// parseJSON parses JSON data into the target object
func parseJSON(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}
