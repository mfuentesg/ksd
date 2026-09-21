package decoder

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDecode_JSON(t *testing.T) {
	in, err := os.ReadFile("../../testdata/secret.json")
	require.NoError(t, err)

	out, err := Decode(in)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &got))

	assert.NotContains(t, got, "data")
	assert.Equal(t, map[string]interface{}{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}, got["stringData"])
	assert.Equal(t, "Secret", got["kind"])
}

func TestDecode_YAML(t *testing.T) {
	in, err := os.ReadFile("../../testdata/secret.yaml")
	require.NoError(t, err)

	out, err := Decode(in)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, yaml.Unmarshal(out, &got))

	assert.NotContains(t, got, "data")
	assert.Equal(t, map[string]interface{}{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}, got["stringData"])
	assert.Equal(t, "Secret", got["kind"])
}

func TestDecode_NoDataField(t *testing.T) {
	in := []byte(`{"apiVersion":"v1","kind":"ConfigMap"}`)

	out, err := Decode(in)

	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func TestDecode_NonStringValues(t *testing.T) {
	in := []byte(`{"data":{"password":"c2VjcmV0","number":123,"boolean":true}}`)

	out, err := Decode(in)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &got))

	assert.Equal(t, map[string]interface{}{
		"password": "secret",
		"number":   "123",
		"boolean":  "true",
	}, got["stringData"])
}

func TestDecode_InvalidBase64Passthrough(t *testing.T) {
	in := []byte(`{"data":{"invalid":"not-base64!@#"}}`)

	out, err := Decode(in)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(out, &got))

	assert.Equal(t, map[string]interface{}{"invalid": "not-base64!@#"}, got["stringData"])
}

func TestDecode_EmptyInput(t *testing.T) {
	out, err := Decode([]byte(""))

	require.NoError(t, err)
	assert.Empty(t, out)
}

func TestDecode_InvalidInput(t *testing.T) {
	_, err := Decode([]byte(`{invalid`))

	assert.Error(t, err)
}

func TestDecode_EmptyDataField(t *testing.T) {
	in := []byte(`{"apiVersion":"v1","kind":"Secret","data":{}}`)

	out, err := Decode(in)

	require.NoError(t, err)
	assert.Equal(t, in, out)
}

func BenchmarkDecode(b *testing.B) {
	in := []byte(`{
		"apiVersion": "v1",
		"kind": "Secret",
		"data": {
			"password": "c2VjcmV0",
			"username": "YWRtaW4="
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decode(in)
	}
}
