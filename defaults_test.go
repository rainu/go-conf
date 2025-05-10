package yacl

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
	if d.String == "" {
		d.String = "Should not be called"
	}
}

type defaultInner2 struct {
	String string `yaml:"string"`
}

func (d *defaultInner2) SetDefaults() {
	if d.String == "" {
		d.String = "defaultInner2"
	}
}

func TestConfig_ApplyDefaults(t *testing.T) {
	c := &defaultS1{}

	toTest := NewConfig(c,
		WithDefaults(func(d *defaultS1) {
			if d.String == "" {
				d.String = "defaultS1"
			}
		}),
		WithDefaults(func(d *defaultInner1) {
			if d.String == "" {
				d.String = "defaultInner1"
			}
		}),
	)

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
}

func TestConfig_ApplyDefaults_WithoutPointerReceiver(t *testing.T) {
	c := &defaultInner1{}

	toTest := NewConfig(c)

	toTest.ParseArguments()
	assert.Equal(t, defaultInner1{
		Inner: defaultInner2{"defaultInner2"},
	}, *c)
}

func TestConfig_ApplyDefaultsOnDynamicElements(t *testing.T) {
	c := &struct {
		A []struct {
			S *string       `yaml:"s"`
			I defaultInner2 `yaml:"i"`
		} `yaml:"a"`
		M map[string]struct {
			S string        `yaml:"s"`
			I defaultInner2 `yaml:"i"`
		} `yaml:"m"`
	}{}
	conf := NewConfig(c)
	assert.NoError(t, conf.ParseArguments(
		"--a.[0].s=string",
		"--a.[1].i.string=myValue",
		`--m.1.s=string`,
		`--m.2.i.string=myValue`,
	))

	assert.Equal(t, "string", *c.A[0].S)
	assert.Equal(t, "defaultInner2", c.A[0].I.String)
	assert.Equal(t, "myValue", c.A[1].I.String)
	assert.Equal(t, "defaultInner2", c.M["1"].I.String)
	assert.Equal(t, "myValue", c.M["2"].I.String)
}
