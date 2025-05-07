package main

import (
	"github.com/goccy/go-yaml"
)

type Config struct {
	options Options

	dest any
}

func NewConfig[T any](dest *T, opts ...Option) *Config {
	p := &Config{
		options: newDefaultOptions(),
		dest:    dest,
	}

	// apply options
	for _, opt := range opts {
		opt(&p.options)
	}

	return p
}

func (c *Config) Parse(args []string) error {
	properties := c.collectHelpProperties()
	reader := newReader(args, &properties, c.options)
	return yaml.NewDecoder(reader, c.options.decodeOptions...).Decode(c.dest)
}

func (c *Config) HelpFlags() string {
	return c.collectHelpProperties().HelpFlags()
}

func (c *Config) HelpYaml() string {
	return c.collectHelpProperties().HelpYaml()
}
