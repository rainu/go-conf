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

func (d defaultInner1) SetDefaults() {
	d.String = "Should not be called"
}

type defaultInner2 struct {
	String string `yaml:"string"`
}

func (d *defaultInner2) SetDefaults() {
	d.String = "defaultInner2"
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
	)
	assert.True(t, toTest.isFirstParse)

	toTest.ParseArguments()
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
	assert.Equal(t, 2, callCount, "Default values should be applied only once (first parse action)")

	toTest.ParseArguments()
	assert.False(t, toTest.isFirstParse)
	assert.Equal(t, 2, callCount, "Default values should be applied only once (first parse action)")
}

func TestConfig_ApplyDefaults_WithoutPointerReceiver(t *testing.T) {
	c := &defaultInner1{}

	toTest := NewConfig(c)
	assert.True(t, toTest.isFirstParse)

	toTest.ParseArguments()
	assert.Equal(t, defaultInner1{
		Inner: defaultInner2{"defaultInner2"},
	}, *c)
	assert.False(t, toTest.isFirstParse)
}
