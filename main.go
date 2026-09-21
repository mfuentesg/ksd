package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mfuentesg/ksd/internal/decoder"
)

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

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	output, err := decoder.Decode(stdin)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not decode secret: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprint(os.Stdout, string(output))
}
