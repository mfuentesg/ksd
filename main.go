package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type secret map[string]interface{}

var version string

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		_, _ = fmt.Fprintf(os.Stdout, "ksd version %s\n", version)
		return
	}
	info, err := os.Stdin.Stat()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		_, _ = fmt.Fprintln(os.Stderr, "the command is intended to work with pipes.")
		_, _ = fmt.Fprintln(os.Stderr, "usage: kubectl get secret <secret-name> -o <yaml|json> |", os.Args[0])
		_, _ = fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "< secret.<yaml|json>")
		os.Exit(1)
	}

	stdin := read(os.Stdin)
	output, err := parse(stdin)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprint(os.Stdout, string(output))
}

func cast(data interface{}, isJSON bool) (map[string]interface{}, bool) {
	if isJSON {
		d, ok := data.(map[string]interface{})
		return d, ok
	}

	parsed, ok := data.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	d := make(map[string]interface{}, len(parsed))
	for key, value := range parsed {
		strKey, ok := key.(string)
		if !ok {
			continue
		}
		d[strKey] = value
	}
	return d, true
}

func parse(in []byte) ([]byte, error) {
	isJSON := isJSONString(in)

	var s secret
	if err := unmarshal(in, &s, isJSON); err != nil {
		return nil, err
	}

	data, ok := cast(s["data"], isJSON)
	if !ok || len(data) == 0 {
		return in, nil
	}
	s["stringData"] = decode(data)
	delete(s, "data")
	return marshal(s, isJSON)
}

func read(rd io.Reader) []byte {
	output, _ := io.ReadAll(rd)
	return output
}

func unmarshal(in []byte, out interface{}, asJSON bool) error {
	if asJSON {
		return json.Unmarshal(in, out)
	}
	return yaml.Unmarshal(in, out)
}

func marshal(d interface{}, asJSON bool) ([]byte, error) {
	if asJSON {
		return json.MarshalIndent(d, "", "    ")
	}
	return yaml.Marshal(d)
}

func decode(data map[string]interface{}) map[string]string {
	decoded := make(map[string]string, len(data))
	for key, encoded := range data {
		strVal, ok := encoded.(string)
		if !ok {
			continue
		}
		if decodedVal, err := base64.StdEncoding.DecodeString(strVal); err == nil {
			decoded[key] = string(decodedVal)
		} else {
			decoded[key] = strVal
		}
	}
	return decoded
}

func isJSONString(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}
