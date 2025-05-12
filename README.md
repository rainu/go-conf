[![Go](https://github.com/rainu/go-yacl/actions/workflows/build.yml/badge.svg)](https://github.com/rainu/go-yacl/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/rainu/go-yacl/branch/main/graph/badge.svg)](https://codecov.io/gh/rainu/go-yacl)
[![Go Report Card](https://goreportcard.com/badge/github.com/rainu/go-yacl)](https://goreportcard.com/report/github.com/rainu/go-yacl)
[![Go Reference](https://pkg.go.dev/badge/github.com/rainu/go-yacl.svg)](https://pkg.go.dev/github.com/rainu/go-yacl)

# go-yacl

**Y**et **a**nother **c**onfiguration **l**ibrary for Go.

This library parses command line flags, environment variables, and configuration files (yaml) into a struct.

With the go's default [flag package](https://pkg.go.dev/flag) there is only the possibility to define "flat" flags: there is no possibility to define dynamic structures.

This library tackle that issue. They can parse arguments like this:

```
$> myApp --map[key]=value --entry[0].key=key1 --entry[0].value=val1
```

The "key-mechanic" is converting command line arguments to yaml-content and then use [yaml parser](github.com/goccy/go-yaml) to parse the content into a struct.
Therefore, you have to use the yaml-tags for defining the key names.

# How to use it

## Basic usage

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	Help bool               `yaml:"help" short:"h" usage:"Show help"`
	Map  map[string]string  `yaml:"map"`
	Entries []struct{
		Key   string `yaml:"key"`
		Value string `yaml:"value"`
	} `yaml:"entry"`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	err := config.ParseArguments("--help")
	if err != nil {
		panic(err)
	}
	if c.Help {
		println(config.HelpFlags())
		return
	}
}
```

## Parse environment

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	Help bool `yaml:"help" short:"h" usage:"Show help"`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	err := config.ParseEnvironment("CFG_0=--help")
	if err != nil {
		panic(err)
	}
	if c.Help {
		println(config.HelpFlags())
		return
	}
}
```

## Parse yaml file

```go
package main

import (
	"github.com/rainu/go-yacl"
	"os"
)

type MyConfig struct {
	Help bool `yaml:"help" short:"h" usage:"Show help"`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)

	yamlFile, err := os.Open("/path/to/config.yaml")
	if err != nil {
		panic(err)
	}
	defer yamlFile.Close()

	err = config.ParseYaml(yamlFile)
	if err != nil {
		panic(err)
	}
	if c.Help {
		println(config.HelpFlags())
		return
	}
}
```

## Define default values

For applying default values you have to define a function which is responsible for setting those values.
**Attention**: This function will be called **after** the unmarshalling! So you have to check if a field is already filled
before setting a default value. Otherwise, you will override the provided value! 
Because of that it is recommended to use pointers for primitive types. Otherwise, you cannot really be sure if the field is intended to be empty.

To define suche a function there are two ways to do this:

### Register a function

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	String *string `yaml:"string"`
}

func Default(m *MyConfig) {
	// check if the field is already set
	if m.String == nil {
        m.String = yacl.P("default")
    }
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c, yacl.WithDefaults(Default))
	err := config.ParseArguments()
	if err != nil {
		panic(err)
	}
	println(c.String) // should be "default"
}
```

### Define a **pointer receiver** function

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	String *string `yaml:"string"`
}

// implements the interface >yacl.DefaultSetter<
func (m *MyConfig) SetDefaults() {
	// check if the field is already set
	if m.String == nil {
		m.String = yacl.P("default")
	}
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	err := config.ParseArguments()
	if err != nil {
		panic(err)
	}
	println(c.String) // should be "default"
}
```

## Define usage

There are three ways to define the usage

### Usage tag

```go
type MyConfig struct {
	String string `yaml:"string" usage:"Put a string here"`
}
```

### Register a function

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	String string `yaml:"string"`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c, yacl.WithUsage(func(t *MyConfig, f string) string {
        if f == "String" {
            return "Put a string here"
        }
		return ""
	}))
	println(config.HelpFlags())
}
```

### Define a **pointer receiver** function

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	String string `yaml:"string"`
}

// implements the interface >yacl.UsageProvider<
func (m *MyConfig) GetUsage(field string) string {
    if field == "String" {
        return "Put a string here"
    }
    return ""
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	println(config.HelpFlags())
}
```

## Nested structs

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	Inner struct {
		Bool bool `yaml:"bool" usage:"bool"`
	} `yaml:"inner" usage:"Inner: "`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	err := config.ParseArguments("--inner.bool=true")
	if err != nil {
		panic(err)
	}
	println(config.HelpFlags())
}
```

## Shadow structs

```go
package main

import (
	"github.com/rainu/go-yacl"
)

type MyConfig struct {
	Inner struct {
		Bool bool `yaml:"bool" usage:"bool"`
	} `yaml:",inline"`
}

func main() {
	c := MyConfig{}

	config := yacl.NewConfig(&c)
	err := config.ParseArguments("--inner.bool=true")
	if err != nil {
		panic(err)
	}
	println(config.HelpFlags())
}
```

## More options

For more options, have a look into the [option.go](./option.go) file.
