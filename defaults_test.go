package conf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type defaultS1 struct {
	Inner defaultInner1 `yaml:"inner"`

	String string `yaml:"string"`
}

type defaultInner1 struct {
	Inner defaultInner2 `yaml:"inner"`

	String string `yaml:"string"`
}

type defaultInner2 struct {
	String string `yaml:"string"`
}

func TestConfig_ApplyDefaults(t *testing.T) {
	c := &defaultS1{}

	callCount := 0
	toTest := NewConfig(c,
		WithDefaults(func(d *defaultS1) {
			d.String = "defaultS1"
			if c == d {
				callCount++
			}
		}),
		WithDefaults(func(d *defaultInner1) {
			d.String = "defaultInner1"
			if &c.Inner == d {
				callCount++
			}
		}),
		WithDefaults(func(d *defaultInner2) {
			d.String = "defaultInner2"
			if &c.Inner.Inner == d {
				callCount++
			}
		}),
	)
	assert.True(t, toTest.isFirstParse)

	toTest.ParseArgs()
	assert.Equal(t, defaultS1{
		Inner: defaultInner1{
			Inner: defaultInner2{
				String: "defaultInner2",
			},
			String: "defaultInner1",
		},
		String: "defaultS1",
	}, *c)
	assert.False(t, toTest.isFirstParse)
	assert.Equal(t, 3, callCount, "Default values should be applied only once (first parse action)")

	toTest.ParseArgs()
	assert.False(t, toTest.isFirstParse)
	assert.Equal(t, 3, callCount, "Default values should be applied only once (first parse action)")
}
