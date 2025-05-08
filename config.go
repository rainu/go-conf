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

	isFirstParse bool

	dest any
}

// NewConfig creates a new Config instance where all parse-results will be reflected in the given destination.
func NewConfig[T any](destination *T, opts ...Option) *Config {
	p := &Config{
		options:      newDefaultOptions(),
		isFirstParse: true,
		dest:         destination,
	}

	// apply options
	for _, opt := range opts {
		opt(&p.options)
	}

	return p
}

// ParseYaml parses the given YAML reader and sets the values in the destination struct.
func (c *Config) ParseYaml(reader io.Reader) error {
	if c.isFirstParse {
		c.isFirstParse = false
		if c.options.autoApplyDefaults {
			c.ApplyDefaults()
		}
	}
	return yaml.NewDecoder(reader, c.options.decodeOptions...).Decode(c.dest)
}

// ParseOsArguments parses the command line arguments (os.Args[1:]) and sets the values in the destination struct.
func (c *Config) ParseOsArguments() error {
	return c.ParseArguments(os.Args[1:]...)
}

// ParseArguments parses the given arguments and sets the values in the destination struct.
func (c *Config) ParseArguments(args ...string) error {
	reader := c.ArgumentReader(args...)
	defer reader.Close()

	err := c.ParseYaml(reader)
	if err != nil && err != io.EOF {
		// ignore EOF error (it would be occurred if no args are given)
		return err
	}
	return nil
}

// ArgumentReader creates a new reader that reads the given arguments and transform them into yaml-format.
func (c *Config) ArgumentReader(args ...string) io.ReadCloser {
	properties := c.collectInfos()
	reader := newReader(args, &properties, c.options)
	return reader
}

// ParseOsEnvironment parses the environment variables (os.Environ()) and sets the values in the destination struct.
func (c *Config) ParseOsEnvironment() error {
	return c.ParseEnvironment(os.Environ()...)
}

// ParseEnvironment parses the given environment variables and sets the values in the destination struct.
func (c *Config) ParseEnvironment(env ...string) error {
	reader := c.EnvironmentReader(env...)
	defer reader.Close()

	err := c.ParseYaml(reader)
	if err != nil && err != io.EOF {
		// ignore EOF error (it would be occurred if no environments are given)
		return err
	}
	return nil
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReader) Close() error {
	return nil
}

// EnvironmentReader creates a new reader that reads the given environment variables and transform them into yaml-format.
func (c *Config) EnvironmentReader(env ...string) io.ReadCloser {
	if c.options.prefixEnv == "" {
		return nil
	}

	re, err := regexp.Compile(`^` + c.options.prefixEnv + `[^=]*=(.*)$`)
	if err != nil {
		return &errorReader{err: fmt.Errorf("invalid env variable prefix: %w", err)}
	}

	// transform env to args
	args := make([]string, 0, len(env))
	for _, e := range env {
		r := re.FindAllStringSubmatch(e, -1)
		if len(r) == 1 {
			args = append(args, r[0][1])
		}
	}

	return c.ArgumentReader(args...)
}

// HelpFlags returns the help text for the flags in a table format.
func (c *Config) HelpFlags() string {
	return c.collectInfos().HelpFlags()
}

// HelpYaml returns the help text for the flags in a YAML format.
func (c *Config) HelpYaml() string {
	return c.collectInfos().HelpYaml()
}
