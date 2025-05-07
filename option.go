package conf

import (
	"github.com/goccy/go-yaml"
	"reflect"
)

type Options struct {
	keyDelimiter rune
	assignSign   rune

	prefixLong  string
	prefixShort string
	prefixEnv   string

	usageTag string
	shortTag string

	defaultSetter map[reflect.Type]func(any)
	decodeOptions []yaml.DecodeOption
}

func newDefaultOptions() Options {
	opts := Options{}

	WithKeyDelimiter('.')(&opts)
	WithAssignSign('=')(&opts)
	WithPrefixLong("--")(&opts)
	WithPrefixShort("-")(&opts)
	WithPrefixEnv("CFG_")(&opts)
	WithUsageTag("usage")(&opts)
	WithShortTag("short")(&opts)

	return opts
}

type Option func(*Options)

// WithKeyDelimiter sets the delimiter for keys. Default is '.'.
func WithKeyDelimiter(delimiter rune) Option {
	return func(o *Options) {
		o.keyDelimiter = delimiter
	}
}

// WithAssignSign sets the sign for assignment. Default is '='.
func WithAssignSign(sign rune) Option {
	return func(o *Options) {
		o.assignSign = sign
	}
}

// WithPrefixLong sets the prefix for the keys (long variant). Default is "--".
func WithPrefixLong(prefix string) Option {
	return func(o *Options) {
		o.prefixLong = prefix
	}
}

// WithPrefixShort sets the prefix for the keys (short variant). Default is "-".
func WithPrefixShort(prefix string) Option {
	return func(o *Options) {
		o.prefixShort = prefix
	}
}

// WithPrefixEnv sets the prefix for environment variables. Default is "CFG_".
func WithPrefixEnv(prefix string) Option {
	return func(o *Options) {
		o.prefixEnv = prefix
	}
}

// WithUsageTag sets the tag for usage. Default is "usage".
func WithUsageTag(tag string) Option {
	return func(o *Options) {
		o.usageTag = tag
	}
}

// WithShortTag sets the tag for short usage. Default is "short".
func WithShortTag(tag string) Option {
	return func(o *Options) {
		o.shortTag = tag
	}
}

// WithDecoderOptions sets the decoder options for the parser.
func WithDecoderOptions(options ...yaml.DecodeOption) Option {
	return func(o *Options) {
		o.decodeOptions = append(o.decodeOptions, options...)
	}
}

// WithDefaults register a function which is responsible for setting default values for the given type.
func WithDefaults[T any](defaultSetter func(*T)) Option {
	return func(o *Options) {
		if o.defaultSetter == nil {
			o.defaultSetter = make(map[reflect.Type]func(any))
		}
		o.defaultSetter[reflect.TypeOf((*T)(nil)).Elem()] = func(v any) {
			defaultSetter(v.(*T))
		}

		o.decodeOptions = append(o.decodeOptions, yaml.CustomUnmarshaler(func(dst *T, bytes []byte) error {
			defaultSetter(dst)
			return yaml.Unmarshal(bytes, dst)
		}))
	}
}
