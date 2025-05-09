package conf

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFieldInfos_HelpFlags(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{Path: fieldPath{{key: "a-key", usage: "help for key1"}}, Short: "k", Type: "string"},
			{Path: fieldPath{{key: "z-key", usage: "help for key3"}}, Type: "int32"},
			{Path: fieldPath{{key: "b-key", usage: "help for key2"}}, Type: "int64"},
			{Path: fieldPath{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "float64"},
			{Path: fieldPath{{key: "p"}, {key: "slice", usage: "help: ", isSlice: true}}, Type: "[]string"},
			{Path: fieldPath{{key: "m"}, {key: "key", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "float32"},
			{Path: fieldPath{{key: "p"}, {key: "map", usage: "help: ", isMap: true}}, Type: "map[]string"},
		},
		options: newDefaultOptions(),
	}

	expected := "  -k,   --a-key            string       help for key1  \n"
	expected += "        --z-key            int32        help for key3  \n"
	expected += "        --b-key            int64        help for key2  \n"
	expected += "        --n.key.[i].value  float64      help: value    \n"
	expected += "        --p.slice          []string     help:          \n"
	expected += "        --m.key.[k].value  float32      help: value    \n"
	expected += "        --p.map.[k]        map[]string  help:          \n"

	assert.Equal(t, expected, infos.HelpFlags())
}

func TestFieldInfos_HelpFlagsWithDefaults(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{Path: fieldPath{{key: "a-key", usage: "help for key1"}}, Short: "k", Type: "string", DefaultValue: "default"},
			{Path: fieldPath{{key: "z-key", usage: "help for key3"}}, Type: "int32", DefaultValue: int32(13)},
			{Path: fieldPath{{key: "b-key", usage: "help for key2"}}, Type: "int64", DefaultValue: int64(12)},
			{Path: fieldPath{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "float64", DefaultValue: float64(13.12)},
			{Path: fieldPath{{key: "m"}, {key: "key", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "float32", DefaultValue: float32(12.13)},
		},
		options: newDefaultOptions(),
	}

	expected := "  -k,   --a-key            string   help for key1     \n"
	expected += "                                    Default: default  \n"
	expected += "        --z-key            int32    help for key3     \n"
	expected += "                                    Default: 13       \n"
	expected += "        --b-key            int64    help for key2     \n"
	expected += "                                    Default: 12       \n"
	expected += "        --n.key.[i].value  float64  help: value       \n"
	expected += "                                    Default: 13.12    \n"
	expected += "        --m.key.[k].value  float32  help: value       \n"
	expected += "                                    Default: 12.13    \n"

	assert.Equal(t, expected, infos.HelpFlags())
}

func TestFieldInfos_HelpYaml(t *testing.T) {
	infos := fieldInfos{
		fi: []fieldInfo{
			{Path: fieldPath{{key: "a", usage: "help for a"}}, Type: "string"},
			{Path: fieldPath{{key: "z", usage: "help for z"}}, Type: "string"},
			{Path: fieldPath{{key: "b", usage: "help for b"}}, Type: "string"},
			{Path: fieldPath{{key: "n"}, {key: "inner", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "int64"},
			{Path: fieldPath{{key: "m"}, {key: "inner", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "int32"},
		},
		options: newDefaultOptions(),
	}

	expected := `
"a": string # help for a
"b": string # help for b
"m":
  "inner":
    "k":
      "value": int32 # help: value
"n":
  "inner":
    -
      "value": int64 # help: value
"z": string # help for z`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(infos.HelpYaml()))
}
