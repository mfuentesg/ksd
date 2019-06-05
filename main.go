package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

type data map[string]interface{}
type secret struct {
	Data map[string]string `json:"data" yaml:"data"`
}

func main() {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (info.Mode()&os.ModeCharDevice) != 0 || info.Size() < 0 {
		fmt.Fprintln(os.Stderr, "the command is intended to work with pipes.")
		fmt.Fprintln(os.Stderr, "usage: kubectl get secret <secret-name> -o <yaml|json> |", os.Args[0])
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "< secret.<yaml|json>")
		os.Exit(1)
	}

	output, err := parse(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stdout, string(output))
}

func parse(rd io.Reader) ([]byte, error) {
	var s secret
	output := read(rd)
	isJSON := isJSONString(output)

	if err := unmarshal(output, &s, isJSON); err != nil {
		return nil, err
	}
	if len(s.Data) <= 0 {
		return output, nil
	}
	if err := decode(&s); err != nil {
		return nil, err
	}

	var d data
	if err := unmarshal(output, &d, isJSON); err != nil {
		return nil, err
	}
	d["data"] = s.Data

	return marshal(d, isJSON)
}

func read(rd io.Reader) []byte {
	var output []byte
	reader := bufio.NewReader(rd)
	for {
		input, err := reader.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		output = append(output, input)
	}
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

func decode(s *secret) error {
	for key, encoded := range s.Data {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return err
		}
		s.Data[key] = string(decoded)
	}
	return nil
}

func isJSONString(s []byte) bool {
	var raw json.RawMessage
	return json.Unmarshal(s, &raw) == nil
}
