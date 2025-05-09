package conf

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type testConfig struct {
	Bool        bool                 `yaml:"bool"`
	Bool2       bool                 `yaml:"bool2"`
	String      string               `yaml:"string" short:"s" usage:"This is a string"`
	StringArray []string             `yaml:"string-array" short:"a"`
	RawMap      map[string]any       `yaml:"raw-map"`
	CustomArray []testEntry          `yaml:"array"`
	CustomMap   map[string]testEntry `yaml:"map"`

	Entry testEntry `yaml:"entry" usage:"The base entry: "`
}

type testEntry struct {
	Key   string `yaml:"key" usage:"The key of the entry" short:"k"`
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
		"--array.[1].key=name1",
		"--array.[1].value=value1",
		"--array.[0].key=name0",
		"--array.[0].value=value0",
		"--array[2].key=name2",
		"--array[2].value=value2",
		"--map.test1.key=name1",
		"--map.test1.value=value1",
		"--map.[test 2].key=name2",
		"--map.[test 2].value=value2",
		"--map[test 3].key=name3",
		"--map[test 3].value=value3",
		"--raw-map.string=\"*&.<>/{}|\"",
		"--raw-map.number=2",
		"--raw-map.[key with space]=value",
		"--string-array.[0]=value1",
		"--string-array.[1]=value2",
		"--string-array.[2]=\"*&.<>/{}|\"",
	}

	assert.NoError(t, NewConfig(&conf).ParseArguments(args...))
	assert.Equal(t, testConfig{
		Bool:   true,
		Bool2:  true,
		String: "hello",
		StringArray: []string{
			"value1",
			"value2",
			"*&.<>/{}|",
		},
		CustomArray: []testEntry{
			{Key: "name0", Value: "value0"},
			{Key: "name1", Value: "value1"},
			{Key: "name2", Value: "value2"},
		},
		CustomMap: map[string]testEntry{
			"test1":  {Key: "name1", Value: "value1"},
			"test 2": {Key: "name2", Value: "value2"},
			"test 3": {Key: "name3", Value: "value3"},
		},
		RawMap: map[string]any{
			"string":         "*&.<>/{}|",
			"key with space": "value",
			"number":         uint64(2),
		},
	}, conf)
}

func TestConfig_Parse_PrimitiveArray(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"--string-array=value1",
		"--string-array=value2",
		"--string-array=\"*&.<>/{}|\"",
	}

	assert.NoError(t, NewConfig(&conf).ParseArguments(args...))
	assert.Equal(t, testConfig{
		StringArray: []string{
			"value1",
			"value2",
			"*&.<>/{}|",
		},
	}, conf)
}

func TestConfig_Parse_ShortPrimitiveArray(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"-a=value1",
		"-a=value2",
		"-a=\"*&.<>/{}|\"",
	}

	assert.NoError(t, NewConfig(&conf).ParseArguments(args...))
	assert.Equal(t, testConfig{
		StringArray: []string{
			"value1",
			"value2",
			"*&.<>/{}|",
		},
	}, conf)
}

func TestConfig_Parse_PrimitiveArray_Single(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"--string-array=value1",
	}

	assert.NoError(t, NewConfig(&conf).ParseArguments(args...))
	assert.Equal(t, testConfig{
		StringArray: []string{
			"value1",
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

	assert.NoError(t, NewConfig(&conf).ParseEnvironment(env...))
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

func TestConfig_ParseEnv_InvalidPrefix(t *testing.T) {
	conf := testConfig{}

	assert.Error(t, NewConfig(&conf, WithPrefixEnv("(")).ParseEnvironment())
}

func TestConfig_Parse_Empty(t *testing.T) {
	conf := testConfig{}

	c := NewConfig(&conf)

	assert.NoError(t, c.ParseArguments())
	assert.NoError(t, c.ParseEnvironment())
}

func TestConfig_Parse_WithDefaults(t *testing.T) {
	conf := testConfig{}

	args := []string{
		"-s=hello",
		"--array.[0].key=name0",
		"--array.[1].key=name1",
		"--map.test1.key=name1",
		"--map.test1.value=value1",
		"--map.[test 2].key=name2",
	}
	p := NewConfig(&conf, WithDefaults(func(t *testEntry) {
		t.Value = "DEFAULT"
	}))

	assert.NoError(t, p.ParseArguments(args...))
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
		Entry: testEntry{
			Value: "DEFAULT", //should be applied by AutoApplyDefaults
		},
	}, conf)
}

func TestConfig_HelpFlags(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf,
		WithDefaults(SetDefaults),
		WithUsage(func(t *testConfig, f string) string {
			if f == "Bool" {
				return "Bool usage"
			}
			return ""
		}),
	)

	expected := "        --bool            bool                  Bool usage                              \n"
	expected += "        --bool2           bool                                                          \n"
	expected += "  -s,   --string          string                This is a string                        \n"
	expected += "  -a,   --string-array    []string                                                      \n"
	expected += "        --raw-map[k]      map[string]interface                                          \n"
	expected += "        --array[i].key    string                The key of the entry                    \n"
	expected += "        --array[i].value  string                The value of the entry                  \n"
	expected += "                                                Default: DEFAULT                        \n"
	expected += "        --map[k].key      string                The key of the entry                    \n"
	expected += "        --map[k].value    string                The value of the entry                  \n"
	expected += "                                                Default: DEFAULT                        \n"
	expected += "  -k,   --entry.key       string                The base entry: The key of the entry    \n"
	expected += "        --entry.value     string                The base entry: The value of the entry  \n"
	expected += "                                                Default: DEFAULT                        \n"

	assert.Equal(t, expected, toTest.HelpFlags())
}

func TestConfig_HelpFlags_Sorted(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf,
		WithDefaults(SetDefaults),
		WithUsage(func(t *testConfig, f string) string {
			if f == "Bool" {
				return "Bool usage"
			}
			return ""
		}),
	)

	expected := "        --array[i].key    string                The key of the entry                    \n"
	expected += "        --array[i].value  string                The value of the entry                  \n"
	expected += "                                                Default: DEFAULT                        \n"
	expected += "        --bool            bool                  Bool usage                              \n"
	expected += "        --bool2           bool                                                          \n"
	expected += "  -k,   --entry.key       string                The base entry: The key of the entry    \n"
	expected += "        --entry.value     string                The base entry: The value of the entry  \n"
	expected += "                                                Default: DEFAULT                        \n"
	expected += "        --map[k].key      string                The key of the entry                    \n"
	expected += "        --map[k].value    string                The value of the entry                  \n"
	expected += "                                                Default: DEFAULT                        \n"
	expected += "        --raw-map[k]      map[string]interface                                          \n"
	expected += "  -s,   --string          string                This is a string                        \n"
	expected += "  -a,   --string-array    []string                                                      \n"

	assert.Equal(t, expected, toTest.HelpFlags(WithSorter(PathSorter)))
}

func TestConfig_HelpFlags_Filtered(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf,
		WithDefaults(SetDefaults),
		WithUsage(func(t *testConfig, f string) string {
			if f == "Bool" {
				return "Bool usage"
			}
			return ""
		}),
	)

	expected := "    --bool  bool  Bool usage  \n"

	testFilter := func(a FieldInfo) bool {
		return a.Path() != "bool"
	}

	assert.Equal(t, expected, toTest.HelpFlags(WithFilter(testFilter)))
}

func TestConfig_HelpYaml(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf, WithDefaults(SetDefaults))

	expected := `
"bool": bool
"bool2": bool
"string": string # This is a string
"string-array":
  - []string
"raw-map":
  "k": map[string]interface
"array":
  -
    "key": string # The key of the entry
    "value": string # The value of the entry
"map":
  "k":
    "key": string # The key of the entry
    "value": string # The value of the entry
"entry":
  "key": string # The base entry: The key of the entry
  "value": string # The base entry: The value of the entry
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(toTest.HelpYaml()))
}

func TestConfig_HelpYaml_Sorted(t *testing.T) {
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
  "k":
    "key": string # The key of the entry
    "value": string # The value of the entry
"raw-map":
  "k": map[string]interface
"string": string # This is a string
"string-array":
  - []string
`

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(toTest.HelpYaml(WithSorter(PathSorter))))
}

func TestConfig_HelpYaml_Filtered(t *testing.T) {
	conf := testConfig{}

	toTest := NewConfig(&conf, WithDefaults(SetDefaults))

	expected := `
"bool": bool
`

	testFilter := func(a FieldInfo) bool {
		return a.Path() != "bool"
	}

	assert.Equal(t, strings.TrimSpace(expected), strings.TrimSpace(toTest.HelpYaml(WithFilter(testFilter))))
}

type parent struct {
	Child `yaml:",inline" usage:"Test"`
}

type Child struct {
	String string            `yaml:"string"`
	Array  []string          `yaml:"array" usage:"array"`
	Map    map[string]string `yaml:"map" usage:"map"`
}

func TestConfig_ShadowStructs(t *testing.T) {
	c := &parent{}
	config := NewConfig(c)

	assert.NoError(t, config.ParseArguments(
		"--string=hello",
		"--array.[0]=0",
		"--array[1]=1",
		"--map.[one]=1",
		"--map[two]=2",
	))
	assert.Equal(t, parent{
		Child{
			String: "hello",
			Array:  []string{"0", "1"},
			Map:    map[string]string{"one": "1", "two": "2"},
		},
	}, *c)

	eArgs := "    --string  string                    \n"
	eArgs += "    --array   []string           array  \n"
	eArgs += "    --map[k]  map[string]string  map    \n"

	assert.Equal(t, eArgs, config.HelpFlags())

	eYaml := `
"string": string
"array":
  - []string # array
"map":
  "k": map[string]string # map
`
	assert.Equal(t, strings.TrimSpace(eYaml), strings.TrimSpace(config.HelpYaml()))
}
