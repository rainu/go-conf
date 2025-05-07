package main

import (
	"github.com/goccy/go-yaml"
	"reflect"
)

type Options struct {
	keyDelimiter rune
	assignSign   rune
	prefix       string

	usageTag string

	defaultSetter map[reflect.Type]func(any)
	decodeOptions []yaml.DecodeOption
}

func newDefaultOptions() Options {
	return Options{
		keyDelimiter: '.',
		assignSign:   '=',
		prefix:       "--",
		usageTag:     "usage",
	}
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

// WithPrefix sets the prefix for the keys. Default is "--".
func WithPrefix(prefix string) Option {
	return func(o *Options) {
		o.prefix = prefix
	}
}

// WithUsageTag sets the tag for usage. Default is "usage".
func WithUsageTag(tag string) Option {
	return func(o *Options) {
		o.usageTag = tag
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
