package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// binPath is built once in TestMain and exercised as a subprocess by every
// test below, so these tests exercise the real stdin/stdout/exit-code
// contract rather than calling unexported main.go internals directly.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ksd-test")
	if err != nil {
		panic(err)
	}

	binPath = filepath.Join(dir, "ksd")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("building ksd for tests: " + err.Error() + "\n" + string(out))
	}

	// os.Exit below skips deferred calls, so clean up explicitly first.
	code := m.Run()
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup: %v\n", err)
	}
	os.Exit(code)
}

func run(t *testing.T, stdin []byte, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("running ksd: %v", err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

func TestVersion(t *testing.T) {
	stdout, _, exitCode := run(t, nil, "version")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "ksd version \n", stdout)
}

func TestNoStdinPrintsUsage(t *testing.T) {
	stdout, stderr, exitCode := run(t, nil)

	assert.Equal(t, 1, exitCode)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "the command is intended to work with pipes.")
}

func TestDecodeJSON(t *testing.T) {
	in, err := os.ReadFile("testdata/secret.json")
	require.NoError(t, err)

	stdout, stderr, exitCode := run(t, in)

	require.Equal(t, 0, exitCode, "stderr: %s", stderr)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got))
	assert.Equal(t, map[string]interface{}{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}, got["stringData"])
}

func TestDecodeYAML(t *testing.T) {
	in, err := os.ReadFile("testdata/secret.yaml")
	require.NoError(t, err)

	stdout, stderr, exitCode := run(t, in)

	require.Equal(t, 0, exitCode, "stderr: %s", stderr)

	var got map[string]interface{}
	require.NoError(t, yaml.Unmarshal([]byte(stdout), &got))
	assert.NotContains(t, got, "data")
	assert.Equal(t, map[string]interface{}{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}, got["stringData"])
}

func TestDecodeInvalidInput(t *testing.T) {
	_, stderr, exitCode := run(t, []byte("{invalid"))

	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stderr, "could not decode secret:")
}
