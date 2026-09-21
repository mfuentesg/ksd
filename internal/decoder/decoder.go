// Package decoder converts the base64-encoded `data` field of a Kubernetes
// Secret manifest into a plaintext `stringData` field.
package decoder

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Decode reads a Kubernetes Secret manifest (YAML or JSON, auto-detected)
// and returns it with its `data` field base64-decoded into `stringData`.
// Input without a `data` field is returned unchanged. Values that aren't
// valid base64 are passed through as-is.
func Decode(in []byte) ([]byte, error) {
	isJSON := isJSONString(in)

	doc, err := unmarshal(in, isJSON)
	if err != nil {
		return nil, err
	}

	data, ok := stringMap(doc["data"])
	if !ok || len(data) == 0 {
		return in, nil
	}

	doc["stringData"] = decodeValues(data)
	delete(doc, "data")

	return marshal(doc, isJSON)
}

func stringMap(v interface{}) (map[string]interface{}, bool) {
	m, ok := v.(map[string]interface{})
	return m, ok
}

func decodeValues(data map[string]interface{}) map[string]string {
	decoded := make(map[string]string, len(data))
	for key, value := range data {
		str, ok := value.(string)
		if !ok {
			decoded[key] = fmt.Sprintf("%v", value)
			continue
		}
		if raw, err := base64.StdEncoding.DecodeString(str); err == nil {
			decoded[key] = string(raw)
		} else {
			decoded[key] = str
		}
	}
	return decoded
}

// unmarshal decodes into a plain map[string]interface{}, not a named type:
// yaml.v3 propagates the destination's named map type onto nested mappings
// too, which would break the map[string]interface{} type assertion on the
// nested "data" field below.
func unmarshal(in []byte, isJSON bool) (map[string]interface{}, error) {
	var doc map[string]interface{}
	var err error
	if isJSON {
		err = json.Unmarshal(in, &doc)
	} else {
		err = yaml.Unmarshal(in, &doc)
	}
	return doc, err
}

func marshal(doc map[string]interface{}, isJSON bool) ([]byte, error) {
	if isJSON {
		return json.MarshalIndent(doc, "", "    ")
	}

	// yaml.v3's default indent is 4 spaces; explicitly use 2 to match the
	// output produced by the yaml.v2-based tool this replaces.
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isJSONString(in []byte) bool {
	return json.Unmarshal(in, &json.RawMessage{}) == nil
}
