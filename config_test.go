package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type testConfig struct {
	Bool        bool                 `yaml:"bool"`
	Bool2       bool                 `yaml:"bool2"`
	String      string               `yaml:"string" short:"s" usage:"This is a string"`
	StringArray []string             `yaml:"string-array"`
	RawMap      map[string]any       `yaml:"raw-map"`
	CustomArray []testEntry          `yaml:"array"`
	CustomMap   map[string]testEntry `yaml:"map"`

	Entry testEntry `yaml:"entry" usage:"The base entry: "`
}

type testEntry struct {
	Key   string `yaml:"key" usage:"The key of the entry"`
	Value string `yaml:"value"`
}

func (t testEntry) GetUsage(field string) string {
	if field == "Value" {
		return "The value of the entry"
	}
	return ""
}

func SetDefaults(t *testEntry) {
	t.Value = "DEFAULT"
}

func TestConfig_Parse_DefaultConfig(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"--bool",
		"--bool2=true",
		"--string=hello",
		"--array.[0].key=name0",
		"--array.[0].value=value0",
		"--array.[1].key=name1",
		"--array.[1].value=value1",
		"--map.test1.key=name1",
		"--map.test1.value=value1",
		"--map.[test 2].key=name2",
		"--map.[test 2].value=value2",
		"--raw-map.string=value",
		"--raw-map.number=2",
		"--raw-map.[key with space]=value",
	}

	assert.NoError(t, NewConfig(&conf).Parse(args))
	assert.Equal(t, testConfig{
		Bool:   true,
		Bool2:  true,
		String: "hello",
		CustomArray: []testEntry{
			{Key: "name0", Value: "value0"},
			{Key: "name1", Value: "value1"},
		},
		CustomMap: map[string]testEntry{
			"test1":  {Key: "name1", Value: "value1"},
			"test 2": {Key: "name2", Value: "value2"},
		},
		RawMap: map[string]any{
			"string":         "value",
			"key with space": "value",
			"number":         uint64(2),
		},
	}, conf)
}

func TestConfig_ParseEnv(t *testing.T) {
	conf := testConfig{}

	env := []string{
		"CFG_0=--bool",
		"CFG_1=--bool2=true",
		"CFG_2=--string=hello",
		"CFG_3=--array.[0].key=name0",
		"CFG_4=--array.[0].value=value0",
		"CFG_5=--array.[1].key=name1",
		"CFG_6=--array.[1].value=value1",
		"CFG_7=--map.test1.key=name1",
		"CFG_8=--map.test1.value=value1",
		"CFG_9=--map.[test 2].key=name2",
		"CFG_10=--map.[test 2].value=value2",
		"CFG_11=--raw-map.string=value",
		"CFG_12=--raw-map.number=2",
		"CFG_13=--raw-map.[key with space]=value",
	}

	assert.NoError(t, NewConfig(&conf).ParseEnv(env))
	assert.Equal(t, testConfig{
		Bool:   true,
		Bool2:  true,
		String: "hello",
		CustomArray: []testEntry{
			{Key: "name0", Value: "value0"},
			{Key: "name1", Value: "value1"},
		},
		CustomMap: map[string]testEntry{
			"test1":  {Key: "name1", Value: "value1"},
			"test 2": {Key: "name2", Value: "value2"},
		},
		RawMap: map[string]any{
			"string":         "value",
			"key with space": "value",
			"number":         uint64(2),
		},
	}, conf)
}

func TestConfig_Parse_WithDefaults(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"--string=hello",
		"--array.[0].key=name0",
		"--array.[1].key=name1",
		"--map.test1.key=name1",
		"--map.test1.value=value1",
		"--map.[test 2].key=name2",
	}
	p := NewConfig(&conf, WithDefaults(func(t *testEntry) {
		t.Value = "DEFAULT"
	}))

	assert.NoError(t, p.Parse(args))
	assert.Equal(t, testConfig{
		String: "hello",
		CustomArray: []testEntry{
			{Key: "name0", Value: "DEFAULT"},
			{Key: "name1", Value: "DEFAULT"},
		},
		CustomMap: map[string]testEntry{
			"test1":  {Key: "name1", Value: "value1"},
			"test 2": {Key: "name2", Value: "DEFAULT"},
		},
	}, conf)
}

func TestConfig_HelpFlags(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf, WithDefaults(SetDefaults))

	expected := "        --array.[i].key    string                                    The key of the entry                    \n"
	expected += "        --array.[i].value  string                (default: DEFAULT)  The value of the entry                  \n"
	expected += "        --bool             bool                                                                              \n"
	expected += "        --bool2            bool                                                                              \n"
	expected += "        --entry.key        string                                    The base entry: The key of the entry    \n"
	expected += "        --entry.value      string                (default: DEFAULT)  The base entry: The value of the entry  \n"
	expected += "        --map.[key].key    string                                    The key of the entry                    \n"
	expected += "        --map.[key].value  string                (default: DEFAULT)  The value of the entry                  \n"
	expected += "        --raw-map          map[string]interface                                                              \n"
	expected += "  -s,   --string           string                                    This is a string                        \n"
	expected += "        --string-array     []string                                                                          \n"

	assert.Equal(t, expected, toTest.HelpFlags())
}

func TestConfig_HelpYaml(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf, WithDefaults(SetDefaults))

	expected := `
"array":
  -
    "key": string # The key of the entry
    "value": string # The value of the entry
"bool": bool
"bool2": bool
"entry":
  "key": string # The base entry: The key of the entry
  "value": string # The base entry: The value of the entry
"map":
  "key":
    "key": string # The key of the entry
    "value": string # The value of the entry
"raw-map": map[string]interface
"string": string # This is a string
"string-array": []string
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(toTest.HelpYaml()))
}
