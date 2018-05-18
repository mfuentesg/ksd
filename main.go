package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var isJSON bool

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
		fmt.Fprintln(os.Stderr, "The command is intended to work with pipes.")
		fmt.Fprintln(os.Stderr, "Usage: kubectl get secret <secret-name> -o <yaml|json> |", os.Args[0])
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "< secret.<yml|json>")
		os.Exit(1)
	}

	output, err := parse(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not decode the incoming secret: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprint(os.Stdout, output)
}

func parse(rd io.Reader) (string, error) {
	output := read(os.Stdin)
	var s secret
	if err := unmarshal(output, &s); err != nil {
		return "", err
	}
	if len(s.Data) <= 0 {
		return string(output), nil
	}
	if err := decode(&s); err != nil {
		return "", err
	}

	var d data
	if err := unmarshal(output, &d); err != nil {
		return "", err
	}
	d["data"] = s.Data
	return string(marshal(d)), nil
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

func unmarshal(in []byte, to interface{}) error {
	err := json.Unmarshal(in, &to)
	isJSON = err == nil

	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(in, &to); err != nil {
		return err
	}
	return nil
}

func marshal(d interface{}) []byte {
	var s []byte
	if isJSON {
		s, _ = json.MarshalIndent(d, "", "  ")
	} else {
		s, _ = yaml.Marshal(d)
	}
	return s
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
