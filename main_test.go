package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	texts := [...]string{
		"this is a plain text",
		`{"this": "is", "a": "text",\n"with": "multiple", "line": "s"}`,
		`version:"2"\ndata:\n\tplain: "yml"`,
		"",
		"text",
		"\t",
		"\n",
		"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		"0",
		"0x00000",
	}

	for _, text := range texts {
		reader := strings.NewReader(text)
		assert.Equal(t, text, string(read(reader)))
	}
}

func BenchmarkRead(b *testing.B) {
	texts := [...]string{
		"this is a plain text",
		`{"this": "is", "a": "text",\n"with": "multiple", "line": "s"}`,
		`version:"2"\ndata:\n\tplain: "yml"`,
		"",
		"text",
		"\t",
		"\n",
		"\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		"0",
		"0x00000",
	}
	for _, text := range texts {
		reader := strings.NewReader(text)
		for n := 0; n < b.N; n++ {
			read(reader)
		}
	}
	b.ReportAllocs()
}

func TestMarshal(t *testing.T) {
	test := map[string]string{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	if byt, err := marshal(test, true); err != nil {
		t.Errorf("wrong marshal: %v got %s ", err, string(byt))
	}

	expected := "{\n    \"app\": \"a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==\",\n    \"password\": \"c2VjcmV0\"\n}"
	byt, _ := marshal(test, true)
	assert.Equal(t, expected, string(byt))

	testYml := map[string]interface{}{
		"data": map[string]string{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
	}

	expected = "data:\n  app: a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==\n  password: c2VjcmV0\n"
	byt, _ = marshal(testYml, false)
	assert.Equal(t, expected, string(byt))
}

func BenchmarkMarshal(b *testing.B) {
	test := map[string]string{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = marshal(test, true)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var j map[string]interface{}
	jsonCase, _ := os.ReadFile("./mock.json")
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[string]interface{}{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
		"kind": "Secret",
		"metadata": map[string]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}

	err := unmarshal(jsonCase, &j, true)
	assert.Nil(t, err)
	assert.NoError(t, err)
	assert.Equal(t, expected, j)
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	jsonCase, _ := os.ReadFile("./mock.json")
	var j map[string]interface{}
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = unmarshal(jsonCase, &j, true)
	}
}

func TestUnmarshalYaml(t *testing.T) {
	var y map[string]interface{}
	yamlCase, _ := os.ReadFile("./mock.yml")
	expected := map[string]interface{}{
		"apiVersion": "v1",
		"data": map[interface{}]interface{}{
			"password": "c2VjcmV0",
			"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
		},
		"kind": "Secret",
		"metadata": map[interface{}]interface{}{
			"name":      "kubernetes secret decoder",
			"namespace": "ksd",
		},
		"type": "Opaque",
	}
	err := unmarshal(yamlCase, &y, false)
	assert.Nil(t, err)
	assert.NoError(t, err)
	assert.Equal(t, expected, y)
}

func BenchmarkUnmarshalYaml(b *testing.B) {
	var y map[string]interface{}
	yamlCase, _ := os.ReadFile("./mock.yml")
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_ = unmarshal(yamlCase, &y, false)
	}
}

func TestSecret_Decode(t *testing.T) {
	data := map[string]interface{}{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}
	result := decode(data)
	expected := map[string]string{
		"password": "secret",
		"app":      "kubernetes secret decoder",
	}
	assert.Equal(t, expected, result)
}

func BenchmarkSecret_Decode(b *testing.B) {
	data := map[string]interface{}{
		"password": "c2VjcmV0",
		"app":      "a3ViZXJuZXRlcyBzZWNyZXQgZGVjb2Rlcg==",
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		decode(data)
	}
}

func TestIsJSONString(t *testing.T) {
	yamlCase, _ := os.ReadFile("./mock.yml")
	wrongTests := [...][]byte{
		nil,
		[]byte(""),
		[]byte("k"),
		[]byte("-"),
		[]byte(`"test": "case"`),
		yamlCase,
	}
	for _, test := range wrongTests {
		if isJSONString(test) {
			t.Errorf("%v must not be a json string", string(test))
		}
	}
	jsonCase, _ := os.ReadFile("./mock.json")
	successCases := [...][]byte{
		[]byte("null"),
		[]byte(`{"valid":"json"}`),
		[]byte(`{"nested": {"json": "string"}}`),
		jsonCase,
	}
	for _, test := range successCases {
		assert.True(t, isJSONString(test))
	}
}

func BenchmarkIsJSONString(b *testing.B) {
	jsonCase, _ := os.ReadFile("./mock.json")
	successCases := [...][]byte{
		[]byte("null"),
		[]byte(`{"valid":"json"}`),
		[]byte(`{"nested": {"json": "string"}}`),
		jsonCase,
	}

	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		for _, test := range successCases {
			isJSONString(test)
		}
	}
}

func TestParse(t *testing.T) {
	_, err := parse([]byte(`{"a"`))
	assert.NotNil(t, err)
	assert.Error(t, err)

	// Return same string without data part
	expected := `{"key": "value"}`
	s, err := parse([]byte(`{"key": "value"}`))
	assert.Nil(t, err)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(s))

	_, err = parse([]byte(`{"data": {"password": "c2VjcmV0"}}`))
	assert.Nil(t, err)
	assert.NoError(t, err)
}

func BenchmarkParse(b *testing.B) {
	reader := []byte(`{"data": {"password": "c2VjcmV0"}}`)
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		_, _ = parse(reader)
	}
}
