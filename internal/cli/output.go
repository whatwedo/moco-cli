package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// writeOutput prints the response body in the selected format. An empty body
// (e.g. HTTP 204) produces no output.
func (a *App) writeOutput(resp []byte) error {
	if len(bytes.TrimSpace(resp)) == 0 {
		return nil
	}

	switch a.Output {
	case "raw":
		_, err := fmt.Fprintf(a.Out, "%s\n", resp)
		return err
	case "json", "":
		var buf bytes.Buffer
		if err := json.Indent(&buf, resp, "", "  "); err != nil {
			// Not valid JSON: print unchanged.
			_, err := fmt.Fprintf(a.Out, "%s\n", resp)
			return err
		}
		_, err := fmt.Fprintf(a.Out, "%s\n", buf.Bytes())
		return err
	default:
		return &UsageError{Err: fmt.Errorf("unknown output format %q (allowed: json, raw)", a.Output)}
	}
}

// ParseData parses the raw JSON string of the --data flag into a map. An empty
// string yields an empty map.
func ParseData(data string) (map[string]any, error) {
	body := map[string]any{}
	if data == "" {
		return body, nil
	}
	if err := json.Unmarshal([]byte(data), &body); err != nil {
		return nil, &UsageError{Err: fmt.Errorf("--data is not a valid JSON object: %w", err)}
	}
	return body, nil
}

// EncodeBody serializes the request body. An empty map yields nil so that no
// body data is sent.
func EncodeBody(body map[string]any) ([]byte, error) {
	if len(body) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("could not serialize body: %w", err)
	}
	return raw, nil
}
