package conf

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type customString string

type dynamicHelp struct {
	String string `yaml:"string"`
}

func (d *dynamicHelp) GetUsage(field string) string {
	return "Dynamic help for " + field
}

type defaultHelp struct {
	String string `yaml:"string"`
}

func (d *defaultHelp) SetDefaults() {
	d.String = "Default value"
}

type defaultWithoutPointerHelp struct {
	String string `yaml:"string"`
}

func (d defaultWithoutPointerHelp) SetDefaults() {
	d.String = "Default value"
}

func TestConfig_collectInfos(t *testing.T) {
	testConfig_collectInfos(t, &struct {
		String string `yaml:"string" short:"s" usage:"help"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{{key: "string", usage: "help"}}, short: "s", sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     customString `yaml:"cs"`
		String string       `yaml:"s"`
		Float  float64      `yaml:"f"`
		Int32  int32        `yaml:"i32"`
		Int    int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{{key: "cs"}}, sType: "string"},
		{path: []*fieldPathNode{{key: "s"}}, sType: "string"},
		{path: []*fieldPathNode{{key: "f"}}, sType: "float64"},
		{path: []*fieldPathNode{{key: "i32"}}, sType: "int32"},
		{path: []*fieldPathNode{{key: "i64"}}, sType: "int64"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     []customString `yaml:"cs"`
		String []string       `yaml:"s"`
		Float  []float64      `yaml:"f"`
		Int32  []int32        `yaml:"i32"`
		Int    []int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{{key: "cs", isSlice: true}}, sType: "[]string"},
		{path: []*fieldPathNode{{key: "s", isSlice: true}}, sType: "[]string"},
		{path: []*fieldPathNode{{key: "f", isSlice: true}}, sType: "[]float64"},
		{path: []*fieldPathNode{{key: "i32", isSlice: true}}, sType: "[]int32"},
		{path: []*fieldPathNode{{key: "i64", isSlice: true}}, sType: "[]int64"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     map[string]customString `yaml:"cs"`
		String map[string]string       `yaml:"s"`
		Float  map[string]float64      `yaml:"f"`
		Int32  map[string]int32        `yaml:"i32"`
		Int    map[string]int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{{key: "cs", isMap: true}}, sType: "map[string]string"},
		{path: []*fieldPathNode{{key: "s", isMap: true}}, sType: "map[string]string"},
		{path: []*fieldPathNode{{key: "f", isMap: true}}, sType: "map[string]float64"},
		{path: []*fieldPathNode{{key: "i32", isMap: true}}, sType: "map[string]int32"},
		{path: []*fieldPathNode{{key: "i64", isMap: true}}, sType: "map[string]int64"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     map[int]customString `yaml:"cs"`
		String map[int]string       `yaml:"s"`
		Float  map[int]float64      `yaml:"f"`
		Int32  map[int]int32        `yaml:"i32"`
		Int    map[int]int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{{key: "cs", isMap: true}}, sType: "map[int]string"},
		{path: []*fieldPathNode{{key: "s", isMap: true}}, sType: "map[int]string"},
		{path: []*fieldPathNode{{key: "f", isMap: true}}, sType: "map[int]float64"},
		{path: []*fieldPathNode{{key: "i32", isMap: true}}, sType: "map[int]int32"},
		{path: []*fieldPathNode{{key: "i64", isMap: true}}, sType: "map[int]int64"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"array" usage:"array help: "`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "value", usage: "value help"},
		}, sType: "string"},
		{path: []*fieldPathNode{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "key", usage: "key help"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Map map[string]struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"map" usage:"map help: "`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "value", usage: "value help"},
		}, sType: "string"},
		{path: []*fieldPathNode{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "key", usage: "key help"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		NoHelp string `yaml:"nohelp"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "nohelp"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		IgnoreMe string `usage:"should be ignored"`
	}{}, nil)

	testConfig_collectInfos(t, &dynamicHelp{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "string", usage: "Dynamic help for String"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &defaultHelp{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "string"},
		}, sType: "string", defaultValue: "Default value"},
	})

	// will not work, but will also not panic
	testConfig_collectInfos(t, &defaultWithoutPointerHelp{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "string"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &defaultHelp{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "string"},
		}, sType: "string", defaultValue: "Default value via external function"},
	}, WithDefaults(func(d *defaultHelp) {
		d.String = "Default value via external function"
	}))

	testConfig_collectInfos(t, &struct {
		Map map[string]struct {
			Value string `yaml:"value" short:"v"`
		} `yaml:"map"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "map", isMap: true},
			{key: "value"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Value string `yaml:"value" short:"v"`
		} `yaml:"array"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "array", isSlice: true},
			{key: "value"},
		}, sType: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Entry struct {
			Values []string `yaml:"value" short:"v"`
		} `yaml:"entry"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "entry"},
			{key: "value", isSlice: true},
		}, sType: "[]string", short: "v"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Array []struct {
				Value string `yaml:"value" short:"v"`
			} `yaml:"array"`
		} `yaml:"array"`
	}{}, []fieldInfo{
		{path: []*fieldPathNode{
			{key: "array", isSlice: true},
			{key: "array", isSlice: true},
			{key: "value"},
		}, sType: "string"},
	})
}

func testConfig_collectInfos[T any](t *testing.T, dst *T, expected []fieldInfo, opts ...Option) {
	t.Run("", func(t *testing.T) {
		r := NewConfig(dst, opts...).collectInfos()

		// ignore field content
		for i := range r.fi {
			r.fi[i].field = reflect.StructField{}
		}
		assert.Equal(t, expected, r.fi)
	})
}
