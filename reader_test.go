package yacl

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	args := []string{
		"--int=42",
		"--string=hello: from another world",
		"--string-array.[0]=value1",
		"--string-array.[1]=value2",
		"--int-array.[0]=1",
		"--int-array.[1]=2",
		"--mystring=hello",
		"--bool=true",
		"--bool-flag",
		"--float=3.14",
		"--inner.name=name",
		"--inner.value=value",
		"--array.[0].name=name0",
		"--array.[0].value=value0",
		"--array.[1].name=name1",
		"--array.[1].value=value1",
		"--array.[2].array.[0].name=name0",
		"--array.[2].array.[0].value=value0",
		"--array.[2].array.[1].name=name1",
		"--array.[2].array.[1].value=value1",
		"--inner-map.test1.name=name1",
		"--inner-map.test1.value=value1",
		"--inner-map.[space key].name=name2",
		"--inner-map.[space key].value=value2",
		"--map.test.key=value",
		"--raw-map.string=value",
		"--raw-map.number=2",
		"--raw-map.[key with space]=value",
	}
	expected := `
"array":
  -
    "name": name0
    "value": value0
  -
    "name": name1
    "value": value1
  -
    "array":
      -
        "name": name0
        "value": value0
      -
        "name": name1
        "value": value1
"bool": true
"bool-flag": true
"float": '3.14'
"inner-map":
  "space key":
    "name": name2
    "value": value2
  "test1":
    "name": name1
    "value": value1
"inner":
  "name": name
  "value": value
"int": 42
"int-array":
  - 1
  - 2
"map":
  "test":
    "key": value
"mystring": hello
"raw-map":
  "key with space": value
  "number": 2
  "string": value
"string": 'hello: from another world'
"string-array":
  - value1
  - value2
`

	result, err := io.ReadAll(newReader(args, nil, newDefaultOptions()))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(result)))
}

func TestReader_Quoting(t *testing.T) {
	args := []string{
		"--string1=hello: from another world",
		"--string2=hello:\nfrom another world",
	}
	expected := `
"string1": 'hello: from another world'
"string2": 'hello:\nfrom another world'
`

	result, err := io.ReadAll(newReader(args, nil, newDefaultOptions()))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(result)))
}

func TestReader_Short(t *testing.T) {
	testStruct := struct {
		String string `yaml:"string" short:"s"`
		Inner  struct {
			Int  int  `yaml:"int" short:"i"`
			Bool bool `yaml:"bool" short:"B"`
		} `yaml:"inner"`
		Bool  bool  `yaml:"bool" short:"b"`
		BoolP *bool `yaml:"boolP" short:"p"`
	}{}
	args := []string{
		"-s=string",
		"-i=42",
		"-b=true",
		"-B",
		"-p",
	}
	infos := NewConfig(&testStruct).collectInfos()

	expected := `
"bool": true
"boolP": true
"inner":
  "bool": true
  "int": 42
"string": string
`

	result, err := io.ReadAll(newReader(args, infos, newDefaultOptions()))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(result)))
}

func TestReader_Short_Split(t *testing.T) {
	testStruct := struct {
		String string `yaml:"string" short:"s"`
		Inner  struct {
			Int  int  `yaml:"int" short:"i"`
			Bool bool `yaml:"bool" short:"b"`
		} `yaml:"inner"`
		Bool bool `yaml:"bool" short:"B"`
	}{}
	args := []string{
		"-s", "string",
		"-B",
		"-b", "true",
		"-i", "42",
	}
	infos := NewConfig(&testStruct).collectInfos()

	expected := `
"bool": true
"inner":
  "bool": true
  "int": 42
"string": string
`

	result, err := io.ReadAll(newReader(args, infos, newDefaultOptions()))
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(string(result)))
}

func Test_collectLines(t *testing.T) {
	tests := []struct {
		given    []string
		expected []line
	}{
		{
			given: []string{"--key1=value1", "--key2=value2"},
			expected: []line{
				{path: []string{"key1"}, value: "value1"},
				{path: []string{"key2"}, value: "value2"},
			},
		},
		{
			given: []string{"--deep.key1=value1", "--deep.key2=value2"},
			expected: []line{
				{path: []string{"deep", "key1"}, value: "value1"},
				{path: []string{"deep", "key2"}, value: "value2"},
			},
		},
		{
			given: []string{"--array.[0].key=value1", "--array.[1].key=value2"},
			expected: []line{
				{path: []string{"array", "[0]", "key"}, value: "value1"},
				{path: []string{"array", "[1]", "key"}, value: "value2"},
			},
		},
		{
			given: []string{"--map.[key with space].key=value"},
			expected: []line{
				{path: []string{"map", "[key with space]", "key"}, value: "value"},
			},
		},
		{
			given: []string{"--map.[key-with.].key=value"},
			expected: []line{
				{path: []string{"map", "[key-with.]", "key"}, value: "value"},
			},
		},
		{
			given: []string{"-ignore", "me", "--not=me"},
			expected: []line{
				{path: []string{"not"}, value: "me"},
			},
		},
		{
			given: []string{"--key=value=with=equals"},
			expected: []line{
				{path: []string{"key"}, value: "value=with=equals"},
			},
		},
		{
			given: []string{"--raw-map.string=value"},
			expected: []line{
				{path: []string{"raw-map", "string"}, value: "value"},
			},
		},
		{
			given: []string{"--raw-map.number=2"},
			expected: []line{
				{path: []string{"raw-map", "number"}, value: "2"},
			},
		},
		{
			given: []string{"--raw-map.[key with space]=value"},
			expected: []line{
				{path: []string{"raw-map", "[key with space]"}, value: "value"},
			},
		},
		{
			given: []string{"--array=v2", "--array=v1"},
			expected: []line{
				{path: []string{"array"}, value: "v2"},
				{path: []string{"array"}, value: "v1"},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d", t.Name(), i), func(t *testing.T) {
			r := newReader(tt.given, nil, newDefaultOptions())
			assert.Equal(t, tt.expected, r.collectLines())
		})
	}
}
