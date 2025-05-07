package main

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"os"
	"regexp"
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

func (c *Config) ParseOsArgs() error {
	return c.Parse(os.Args[1:])
}

func (c *Config) Parse(args []string) error {
	properties := c.collectHelpProperties()
	reader := newReader(args, &properties, c.options)
	return yaml.NewDecoder(reader, c.options.decodeOptions...).Decode(c.dest)
}

func (c *Config) ParseOsEnv() error {
	return c.ParseEnv(os.Environ())
}

func (c *Config) ParseEnv(env []string) error {
	if c.options.prefixEnv == "" {
		return nil
	}

	re, err := regexp.Compile(`^` + c.options.prefixEnv + `[^=]*=(.*)$`)
	if err != nil {
		return fmt.Errorf("invalid env variable prefix: %w", err)
	}

	// transform env to args
	args := make([]string, 0, len(env))
	for _, e := range env {
		r := re.FindAllStringSubmatch(e, -1)
		if len(r) == 1 {
			args = append(args, r[0][1])
		}
	}

	return c.Parse(args)
}

func (c *Config) HelpFlags() string {
	return c.collectHelpProperties().HelpFlags()
}

func (c *Config) HelpYaml() string {
	return c.collectHelpProperties().HelpYaml()
}
