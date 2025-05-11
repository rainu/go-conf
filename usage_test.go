package yacl

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func TestFieldInfos_HelpFlags(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{path: fieldPath{{key: "a-key", usage: "help for key1"}}, short: "k", sType: "string"},
			{path: fieldPath{{key: "z-key", usage: "help for key3"}}, sType: "int32"},
			{path: fieldPath{{key: "b-key", usage: "help for key2"}}, sType: "*int64"},
			{path: fieldPath{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, sType: "float64"},
			{path: fieldPath{{key: "p"}, {key: "slice", usage: "help: ", isSlice: true}}, sType: "[]string"},
			{path: fieldPath{{key: "m"}, {key: "key", usage: "help: ", isMap: true, mapKeyType: reflect.TypeOf("")}, {key: "value", usage: "value"}}, sType: "float32"},
			{path: fieldPath{{key: "p"}, {key: "map", usage: "help: ", isMap: true, mapKeyType: reflect.TypeOf("")}}, sType: "map[]string", field: reflect.StructField{Type: reflect.TypeOf(map[string]string{})}},
		},
		options: newDefaultOptions(),
	}

	expected :=
		`  -k, --a-key=string
      	help for key1
      --z-key=int32
      	help for key3
      --b-key=int64
      	help for key2
      --n.key[int].value=float64
      	help: value
      --p.slice=[]string
      	help: 
      --m.key[string].value=float32
      	help: value
      --p.map[string]=string
      	help: 
`

	assert.Equal(t, expected, infos.HelpFlags())
}

func TestFieldInfos_HelpFlagsWithDefaults(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{path: fieldPath{{key: "a-key", usage: "help for key1"}}, short: "k", sType: "string", defaultValue: "default"},
			{path: fieldPath{{key: "z-key", usage: "help for key3"}}, sType: "int32", defaultValue: int32(13)},
			{path: fieldPath{{key: "b-key", usage: "help for key2"}}, sType: "int64", defaultValue: int64(12)},
			{path: fieldPath{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "detail help with\nnewline"}}, sType: "float64", defaultValue: float64(13.12)},
			{path: fieldPath{{key: "m"}, {key: "key", usage: "help: ", isMap: true, mapKeyType: reflect.TypeOf("")}, {key: "value", usage: "value"}}, sType: "float32", defaultValue: float32(12.13)},
		},
		options: newDefaultOptions(),
	}

	expected :=
		`  -k, --a-key=string
      	help for key1
      	Default: default
      --z-key=int32
      	help for key3
      	Default: 13
      --b-key=int64
      	help for key2
      	Default: 12
      --n.key[int].value=float64
      	help: detail help with
      	newline
      	Default: 13.12
      --m.key[string].value=float32
      	help: value
      	Default: 12.13
`

	assert.Equal(t, expected, infos.HelpFlags())
}

func TestFieldInfos_HelpYaml(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{path: fieldPath{{key: "a", usage: "help for a"}}, sType: "string"},
			{path: fieldPath{{key: "z", usage: "help for z"}}, sType: "string"},
			{path: fieldPath{{key: "b", usage: "help for b"}}, sType: "string"},
			{path: fieldPath{{key: "n"}, {key: "inner", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, sType: "int64"},
			{path: fieldPath{{key: "m"}, {key: "inner", usage: "help: ", isMap: true, mapKeyType: reflect.TypeOf("")}, {key: "value", usage: "value"}}, sType: "int32"},
		},
		options: newDefaultOptions(),
	}
	expected := `
"a": string # help for a
"z": string # help for z
"b": string # help for b
"n":
  "inner":
    -
      "value": int64 # help: value
"m":
  "inner":
    "string":
      "value": int32 # help: value
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(infos.HelpYaml()))
}
