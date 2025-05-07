package conf

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type customString string

type dynamicHelp struct {
	String string `yaml:"string"`
}

func (d *dynamicHelp) GetUsage(field string) string {
	return "Dynamic help for " + field
}

type defaultHelp struct {
	String string `yaml:"string"`
}

func TestConfig_CollectHelpProperties(t *testing.T) {
	testConfig_CollectHelpProperties(t, &struct {
		String string `yaml:"string" short:"s" usage:"help"`
	}{}, []Property{
		{Path: []*pathChild{{key: "string", usage: "help"}}, Short: "s", Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Cs     customString `yaml:"cs"`
		String string       `yaml:"s"`
		Float  float64      `yaml:"f"`
		Int32  int32        `yaml:"i32"`
		Int    int64        `yaml:"i64"`
	}{}, []Property{
		{Path: []*pathChild{{key: "cs"}}, Type: "string"},
		{Path: []*pathChild{{key: "f"}}, Type: "float64"},
		{Path: []*pathChild{{key: "i32"}}, Type: "int32"},
		{Path: []*pathChild{{key: "i64"}}, Type: "int64"},
		{Path: []*pathChild{{key: "s"}}, Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Cs     []customString `yaml:"cs"`
		String []string       `yaml:"s"`
		Float  []float64      `yaml:"f"`
		Int32  []int32        `yaml:"i32"`
		Int    []int64        `yaml:"i64"`
	}{}, []Property{
		{Path: []*pathChild{{key: "cs"}}, Type: "[]string"},
		{Path: []*pathChild{{key: "f"}}, Type: "[]float64"},
		{Path: []*pathChild{{key: "i32"}}, Type: "[]int32"},
		{Path: []*pathChild{{key: "i64"}}, Type: "[]int64"},
		{Path: []*pathChild{{key: "s"}}, Type: "[]string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Cs     map[string]customString `yaml:"cs"`
		String map[string]string       `yaml:"s"`
		Float  map[string]float64      `yaml:"f"`
		Int32  map[string]int32        `yaml:"i32"`
		Int    map[string]int64        `yaml:"i64"`
	}{}, []Property{
		{Path: []*pathChild{{key: "cs"}}, Type: "map[string]string"},
		{Path: []*pathChild{{key: "f"}}, Type: "map[string]float64"},
		{Path: []*pathChild{{key: "i32"}}, Type: "map[string]int32"},
		{Path: []*pathChild{{key: "i64"}}, Type: "map[string]int64"},
		{Path: []*pathChild{{key: "s"}}, Type: "map[string]string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Cs     map[int]customString `yaml:"cs"`
		String map[int]string       `yaml:"s"`
		Float  map[int]float64      `yaml:"f"`
		Int32  map[int]int32        `yaml:"i32"`
		Int    map[int]int64        `yaml:"i64"`
	}{}, []Property{
		{Path: []*pathChild{{key: "cs"}}, Type: "map[int]string"},
		{Path: []*pathChild{{key: "f"}}, Type: "map[int]float64"},
		{Path: []*pathChild{{key: "i32"}}, Type: "map[int]int32"},
		{Path: []*pathChild{{key: "i64"}}, Type: "map[int]int64"},
		{Path: []*pathChild{{key: "s"}}, Type: "map[int]string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Array []struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"array" usage:"array help: "`
	}{}, []Property{
		{Path: []*pathChild{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "key", usage: "key help"},
		}, Type: "string"},
		{Path: []*pathChild{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "value", usage: "value help"},
		}, Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		Map map[string]struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"map" usage:"map help: "`
	}{}, []Property{
		{Path: []*pathChild{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "key", usage: "key help"},
		}, Type: "string"},
		{Path: []*pathChild{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "value", usage: "value help"},
		}, Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		NoHelp string `yaml:"nohelp"`
	}{}, []Property{
		{Path: []*pathChild{
			{key: "nohelp"},
		}, Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &struct {
		IgnoreMe string `usage:"should be ignored"`
	}{}, nil)

	testConfig_CollectHelpProperties(t, &dynamicHelp{}, []Property{
		{Path: []*pathChild{
			{key: "string", usage: "Dynamic help for String"},
		}, Type: "string"},
	})

	testConfig_CollectHelpProperties(t, &defaultHelp{}, []Property{
		{Path: []*pathChild{
			{key: "string"},
		}, Type: "string", DefaultValue: "Default value"},
	}, WithDefaults(func(d *defaultHelp) {
		d.String = "Default value"
	}))
}

func testConfig_CollectHelpProperties[T any](t *testing.T, dst *T, expected []Property, opts ...Option) {
	t.Run("", func(t *testing.T) {
		r := NewConfig(dst, opts...).collectHelpProperties()
		assert.Equal(t, expected, r.Get())
	})
}

func TestProperties_HelpText(t *testing.T) {
	properties := Properties{
		p: []Property{
			{Path: path{{key: "a-key", usage: "help for key1"}}, Short: "k", Type: "string"},
			{Path: path{{key: "z-key", usage: "help for key3"}}, Type: "int32"},
			{Path: path{{key: "b-key", usage: "help for key2"}}, Type: "int64"},
			{Path: path{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "float64"},
			{Path: path{{key: "m"}, {key: "key", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "float32"},
		},
		options: newDefaultOptions(),
	}

	expected := "  -k,   --a-key              string     help for key1  \n"
	expected += "        --z-key              int32      help for key3  \n"
	expected += "        --b-key              int64      help for key2  \n"
	expected += "        --n.key.[i].value    float64    help: value    \n"
	expected += "        --m.key.[key].value  float32    help: value    \n"

	assert.Equal(t, expected, properties.HelpFlags())
}

func TestProperties_HelpTextWithDefaults(t *testing.T) {
	properties := Properties{
		p: []Property{
			{Path: path{{key: "a-key", usage: "help for key1"}}, Short: "k", Type: "string", DefaultValue: "default"},
			{Path: path{{key: "z-key", usage: "help for key3"}}, Type: "int32", DefaultValue: int32(13)},
			{Path: path{{key: "b-key", usage: "help for key2"}}, Type: "int64", DefaultValue: int64(12)},
			{Path: path{{key: "n"}, {key: "key", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "float64", DefaultValue: float64(13.12)},
			{Path: path{{key: "m"}, {key: "key", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "float32", DefaultValue: float32(12.13)},
		},
		options: newDefaultOptions(),
	}

	expected := "  -k,   --a-key              string   (default: default)  help for key1  \n"
	expected += "        --z-key              int32    (default: 13)       help for key3  \n"
	expected += "        --b-key              int64    (default: 12)       help for key2  \n"
	expected += "        --n.key.[i].value    float64  (default: 13.12)    help: value    \n"
	expected += "        --m.key.[key].value  float32  (default: 12.13)    help: value    \n"

	assert.Equal(t, expected, properties.HelpFlags())
}

func TestProperties_HelpTextYaml(t *testing.T) {
	properties := Properties{
		p: []Property{
			{Path: path{{key: "a", usage: "help for a"}}, Type: "string"},
			{Path: path{{key: "z", usage: "help for z"}}, Type: "string"},
			{Path: path{{key: "b", usage: "help for b"}}, Type: "string"},
			{Path: path{{key: "n"}, {key: "inner", usage: "help: ", isSlice: true}, {key: "value", usage: "value"}}, Type: "int64"},
			{Path: path{{key: "m"}, {key: "inner", usage: "help: ", isMap: true}, {key: "value", usage: "value"}}, Type: "int32"},
		},
		options: newDefaultOptions(),
	}

	expected := `
"a": string # help for a
"b": string # help for b
"m":
  "inner":
    "key":
      "value": int32 # help: value
"n":
  "inner":
    -
      "value": int64 # help: value
"z": string # help for z`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(properties.HelpYaml()))
}
