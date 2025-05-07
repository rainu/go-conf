[![Go](https://github.com/rainu/go-conf/actions/workflows/build.yml/badge.svg)](https://github.com/rainu/go-conf/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/rainu/go-conf/branch/main/graph/badge.svg)](https://codecov.io/gh/rainu/go-conf)
[![Go Report Card](https://goreportcard.com/badge/github.com/rainu/go-conf)](https://goreportcard.com/report/github.com/rainu/go-conf)
[![Go Reference](https://pkg.go.dev/badge/github.com/rainu/go-conf.svg)](https://pkg.go.dev/github.com/rainu/go-conf)

# go-conf

This library can parse command line flags, environment variables, and configuration files.

# How to use it

```go
package main

import (
	conf "github.com/rainu/go-conf"
)

type MyConfig struct {
	String      string             `yaml:"string" short:"s" usage:"This is a string"`
	StringArray []string           `yaml:"string-array"`
	RawMap      map[string]any     `yaml:"raw-map"`
	CustomArray []MyEntry          `yaml:"array"`
	CustomMap   map[string]MyEntry `yaml:"map"`

	Entry MyEntry `yaml:"entry" usage:"The base entry: "`
}

type MyEntry struct {
	Key   string `yaml:"key" usage:"The key of the entry"`
	Value string `yaml:"value"`
}

func SetDefaults(t *MyEntry) {
	t.Value = "DEFAULT"
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c,
		conf.WithDefaults(SetDefaults), // function to set default values for "MyEntry"
	)

	config.Parse([]string{
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
	})

	// print help
	println(config.HelpFlags())

	// print yaml example
	println(config.HelpYaml())
}

```

Output of `config.HelpFlags()`:
```text
        --array.[i].key    string                                    The key of the entry                    
        --array.[i].value  string                (default: DEFAULT)  The value of the entry                  
        --entry.key        string                                    The base entry: The key of the entry    
        --entry.value      string                (default: DEFAULT)  The base entry: The value of the entry  
        --map.[key].key    string                                    The key of the entry                    
        --map.[key].value  string                (default: DEFAULT)  The value of the entry                  
        --raw-map          map[string]interface                                                              
  -s,   --string           string                                    This is a string                        
        --string-array     []string                                                                          
```

Output of `config.HelpYaml()`:
```yaml
"array":
  -
    "key": string # The key of the entry
    "value": string # The value of the entry
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
```