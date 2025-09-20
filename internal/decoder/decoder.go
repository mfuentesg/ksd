package decoder

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// Secret represents a Kubernetes secret structure
type Secret map[string]interface{}

// Decoder handles the decoding of Kubernetes secrets
type Decoder struct {
	config *Config
}

// Config holds decoder configuration
type Config struct {
	OutputFormat string // "json", "yaml", or "auto"
	PrettyPrint  bool
	ShowBinary   bool
	Validate     bool
}

// NewDecoder creates a new decoder with the given configuration
func NewDecoder(config *Config) *Decoder {
	if config == nil {
		config = &Config{
			OutputFormat: "auto",
			PrettyPrint:  true,
			ShowBinary:   false,
			Validate:     true,
		}
	}
	return &Decoder{config: config}
}

// Decode processes the input data and returns decoded secret
func (d *Decoder) Decode(input []byte) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("empty input")
	}

	isJSON := d.isJSONString(input)
	
	var secret Secret
	if err := d.unmarshal(input, &secret, isJSON); err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	if d.config.Validate {
		if err := d.validateSecret(secret); err != nil {
			return nil, fmt.Errorf("invalid secret format: %w", err)
		}
	}

	// Process the secret data
	if err := d.processSecretData(&secret, isJSON); err != nil {
		return nil, fmt.Errorf("failed to process secret data: %w", err)
	}

	// Determine output format
	outputFormat := d.determineOutputFormat(isJSON)
	
	return d.marshal(secret, outputFormat == "json")
}

// processSecretData decodes base64 data and moves it to stringData
func (d *Decoder) processSecretData(secret *Secret, isJSON bool) error {
	dataField := (*secret)["data"]
	if dataField == nil {
		return nil
	}
	
	data, ok := d.cast(dataField, isJSON)
	if !ok || len(data) == 0 {
		return nil
	}

	stringData := d.decodeData(data)
	(*secret)["stringData"] = stringData
	delete(*secret, "data")
	
	return nil
}

// validateSecret checks if the input is a valid Kubernetes secret
func (d *Decoder) validateSecret(secret Secret) error {
	// Check required fields
	if apiVersion, ok := secret["apiVersion"].(string); !ok || apiVersion == "" {
		return fmt.Errorf("missing or invalid apiVersion")
	}
	
	if kind, ok := secret["kind"].(string); !ok || kind != "Secret" {
		return fmt.Errorf("not a Secret resource (kind: %v)", secret["kind"])
	}
	
	return nil
}

// cast converts interface{} to map[string]interface{} handling both JSON and YAML
func (d *Decoder) cast(data interface{}, isJSON bool) (map[string]interface{}, bool) {
	// Try map[string]interface{} first (works for both JSON and some YAML cases)
	if result, ok := data.(map[string]interface{}); ok {
		return result, true
	}

	// Try Secret type (which is map[string]interface{})
	if result, ok := data.(Secret); ok {
		return map[string]interface{}(result), true
	}

	// Try map[interface{}]interface{} (YAML specific)
	if parsed, ok := data.(map[interface{}]interface{}); ok {
		result := make(map[string]interface{}, len(parsed))
		for key, value := range parsed {
			strKey, ok := key.(string)
			if !ok {
				continue
			}
			result[strKey] = value
		}
		return result, true
	}
	
	return nil, false
}

// decodeData decodes base64 encoded values in the secret data
func (d *Decoder) decodeData(data map[string]interface{}) map[string]string {
	decoded := make(map[string]string, len(data))
	
	for key, encoded := range data {
		strVal, ok := encoded.(string)
		if !ok {
			decoded[key] = fmt.Sprintf("%v", encoded)
			continue
		}
		
		// Try to decode as base64
		if decodedBytes, err := base64.StdEncoding.DecodeString(strVal); err == nil {
			// Check if decoded data is valid UTF-8 text
			if utf8.Valid(decodedBytes) {
				decoded[key] = string(decodedBytes)
			} else if d.config.ShowBinary {
				decoded[key] = fmt.Sprintf("<binary data: %d bytes>", len(decodedBytes))
			} else {
				decoded[key] = "<binary data hidden>"
			}
		} else {
			// Not valid base64, keep original value
			decoded[key] = strVal
		}
	}
	
	return decoded
}

// determineOutputFormat determines the output format based on config and input
func (d *Decoder) determineOutputFormat(inputIsJSON bool) string {
	switch d.config.OutputFormat {
	case "json":
		return "json"
	case "yaml":
		return "yaml"
	case "auto":
		if inputIsJSON {
			return "json"
		}
		return "yaml"
	default:
		return "json"
	}
}

// unmarshal parses input data as JSON or YAML
func (d *Decoder) unmarshal(input []byte, out interface{}, asJSON bool) error {
	if asJSON {
		return json.Unmarshal(input, out)
	}
	return yaml.Unmarshal(input, out)
}

// marshal serializes data as JSON or YAML
func (d *Decoder) marshal(data interface{}, asJSON bool) ([]byte, error) {
	if asJSON {
		if d.config.PrettyPrint {
			return json.MarshalIndent(data, "", "    ")
		}
		return json.Marshal(data)
	}
	return yaml.Marshal(data)
}

// isJSONString checks if the input is valid JSON
func (d *Decoder) isJSONString(input []byte) bool {
	return json.Unmarshal(input, &json.RawMessage{}) == nil
}