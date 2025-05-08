[![Go](https://github.com/rainu/go-conf/actions/workflows/build.yml/badge.svg)](https://github.com/rainu/go-conf/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/rainu/go-conf/branch/main/graph/badge.svg)](https://codecov.io/gh/rainu/go-conf)
[![Go Report Card](https://goreportcard.com/badge/github.com/rainu/go-conf)](https://goreportcard.com/report/github.com/rainu/go-conf)
[![Go Reference](https://pkg.go.dev/badge/github.com/rainu/go-conf.svg)](https://pkg.go.dev/github.com/rainu/go-conf)

# go-conf

This library parses command line flags, environment variables, and configuration files (yaml) into a struct.

The "key-mechanic" is to convert command line arguments to yaml-content and then use [yaml parser](github.com/goccy/go-yaml) to parse the content into a struct.
Therefore, you have to use the yaml-tags for defining the key names.

# How to use it

## Basic usage

```go
package main

import (
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	Help bool `yaml:"help" short:"h" usage:"Show help"`
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)
	err := config.ParseArgs("--help")
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
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	Help bool `yaml:"help" short:"h" usage:"Show help"`
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)
	err := config.ParseEnv("CFG_0=--help")
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
	"github.com/rainu/go-conf"
	"os"
)

type MyConfig struct {
	Help bool `yaml:"help" short:"h" usage:"Show help"`
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)

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

For applying default values there are two ways to do this:

### Register a function

```go
package main

import (
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	String string `yaml:"string" short:"s" usage:"String value"`
}

func Default(m *MyConfig) {
	m.String = "default"
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c, conf.WithDefaults(Default))
	err := config.ParseArgs()
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
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	String string `yaml:"string" short:"s" usage:"String value"`
}

// implements the interface >conf.DefaultSetter<
func (m *MyConfig) SetDefaults() {
	m.String = "default"
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)
	err := config.ParseArgs()
	if err != nil {
		panic(err)
	}
	println(c.String) // should be "default"
}
```

## Define dynamic usage

Sometimes you want to define a dynamic usage for a flag. Therefore, your struct must declare a function which is responsible for getting the usage.

```go
package main

import (
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	String string `yaml:"string" short:"s"`
}

func (c *MyConfig) GetUsage(field string) string {
	if field == "String" {
		return "Dynamic help for string"
	}
	return ""
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)
	println(config.HelpFlags())
}
```

## Nested structs

```go
package main

import (
	"github.com/rainu/go-conf"
)

type MyConfig struct {
	Inner struct {
		Bool bool `yaml:"bool" usage:"bool"`
	} `yaml:"inner" usage:"Inner: "`
}

func main() {
	c := MyConfig{}

	config := conf.NewConfig(&c)
	err := config.ParseArgs("--inner.bool=true")
	if err != nil {
		panic(err)
	}
	println(config.HelpFlags())
}
```

## More options

For more options, have a look into the [options.go](./options.go) file.