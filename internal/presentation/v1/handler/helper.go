package handler

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func toURLFormat(v any) (string, error) {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("Error marshaling JSON: %w\n", err)
	}

	// URL encode the JSON string
	encoded := url.QueryEscape(string(jsonBytes))
	return encoded, nil
}
