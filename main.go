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

type decodedSecret struct {
	Key string
	Value string
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

	stdin := read(os.Stdin)
	output, err := parse(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stdout, string(output))
}

func parse(in []byte) ([]byte, error) {
	var s secret
	isJSON := isJSONString(in)

	if err := unmarshal(in, &s, isJSON); err != nil {
		return nil, err
	}
	if len(s.Data) == 0 {
		return in, nil
	}

	s.Decode()

	var d data
	if err := unmarshal(in, &d, isJSON); err != nil {
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

func decodeData(data map[string]string, channel chan decodedSecret) {
	var value string
	for key, encoded := range data {
		// avoid wrong encoded secrets
		if decoded, err := base64.StdEncoding.DecodeString(encoded); err ==  nil {
			value = string(decoded)
		} else {
			value = encoded
		}

		channel <- decodedSecret{
			Key: key,
			Value: value,
		}
	}
	close(channel)
}

func (s *secret) Decode() {
	channel := make(chan decodedSecret, len(s.Data))
	go decodeData(s.Data, channel)
	for secret := range channel {
		s.Data[secret.Key] = secret.Value
	}
}

func isJSONString(s []byte) bool {
	return json.Unmarshal(s, &json.RawMessage{}) == nil
}
