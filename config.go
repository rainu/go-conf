package conf

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"io"
	"os"
	"regexp"
)

type Config struct {
	options Options

	dest any
}

// NewConfig creates a new Config instance where all parse-results will be reflected in the given destination.
func NewConfig[T any](destination *T, opts ...Option) *Config {
	p := &Config{
		options: newDefaultOptions(),
		dest:    destination,
	}

	// apply options
	for _, opt := range opts {
		opt(&p.options)
	}

	return p
}

// ParseYaml parses the given YAML reader and sets the values in the destination struct.
func (c *Config) ParseYaml(reader io.Reader) error {
	return yaml.NewDecoder(reader, c.options.decodeOptions...).Decode(c.dest)
}

// ParseOsArgs parses the command line arguments (os.Args[1:]) and sets the values in the destination struct.
func (c *Config) ParseOsArgs() error {
	return c.ParseArgs(os.Args[1:]...)
}

// ParseArgs parses the given arguments and sets the values in the destination struct.
func (c *Config) ParseArgs(args ...string) error {
	properties := c.collectHelpProperties()
	reader := newReader(args, &properties, c.options)
	defer reader.Close()

	return c.ParseYaml(reader)
}

// ParseOsEnv parses the environment variables (os.Environ()) and sets the values in the destination struct.
func (c *Config) ParseOsEnv() error {
	return c.ParseEnv(os.Environ()...)
}

// ParseEnv parses the given environment variables and sets the values in the destination struct.
func (c *Config) ParseEnv(env ...string) error {
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

	return c.ParseArgs(args...)
}

// HelpFlags returns the help text for the flags in a table format.
func (c *Config) HelpFlags() string {
	return c.collectHelpProperties().HelpFlags()
}

// HelpYaml returns the help text for the flags in a YAML format.
func (c *Config) HelpYaml() string {
	return c.collectHelpProperties().HelpYaml()
}
