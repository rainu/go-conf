package conf

import (
	"github.com/stretchr/testify/assert"
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
		{Path: []*fieldPathNode{{key: "string", usage: "help"}}, Short: "s", Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     customString `yaml:"cs"`
		String string       `yaml:"s"`
		Float  float64      `yaml:"f"`
		Int32  int32        `yaml:"i32"`
		Int    int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{{key: "cs"}}, Type: "string"},
		{Path: []*fieldPathNode{{key: "f"}}, Type: "float64"},
		{Path: []*fieldPathNode{{key: "i32"}}, Type: "int32"},
		{Path: []*fieldPathNode{{key: "i64"}}, Type: "int64"},
		{Path: []*fieldPathNode{{key: "s"}}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     []customString `yaml:"cs"`
		String []string       `yaml:"s"`
		Float  []float64      `yaml:"f"`
		Int32  []int32        `yaml:"i32"`
		Int    []int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{{key: "cs", isSlice: true}}, Type: "[]string"},
		{Path: []*fieldPathNode{{key: "f", isSlice: true}}, Type: "[]float64"},
		{Path: []*fieldPathNode{{key: "i32", isSlice: true}}, Type: "[]int32"},
		{Path: []*fieldPathNode{{key: "i64", isSlice: true}}, Type: "[]int64"},
		{Path: []*fieldPathNode{{key: "s", isSlice: true}}, Type: "[]string"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     map[string]customString `yaml:"cs"`
		String map[string]string       `yaml:"s"`
		Float  map[string]float64      `yaml:"f"`
		Int32  map[string]int32        `yaml:"i32"`
		Int    map[string]int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{{key: "cs", isMap: true}}, Type: "map[string]string"},
		{Path: []*fieldPathNode{{key: "f", isMap: true}}, Type: "map[string]float64"},
		{Path: []*fieldPathNode{{key: "i32", isMap: true}}, Type: "map[string]int32"},
		{Path: []*fieldPathNode{{key: "i64", isMap: true}}, Type: "map[string]int64"},
		{Path: []*fieldPathNode{{key: "s", isMap: true}}, Type: "map[string]string"},
	})

	testConfig_collectInfos(t, &struct {
		Cs     map[int]customString `yaml:"cs"`
		String map[int]string       `yaml:"s"`
		Float  map[int]float64      `yaml:"f"`
		Int32  map[int]int32        `yaml:"i32"`
		Int    map[int]int64        `yaml:"i64"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{{key: "cs", isMap: true}}, Type: "map[int]string"},
		{Path: []*fieldPathNode{{key: "f", isMap: true}}, Type: "map[int]float64"},
		{Path: []*fieldPathNode{{key: "i32", isMap: true}}, Type: "map[int]int32"},
		{Path: []*fieldPathNode{{key: "i64", isMap: true}}, Type: "map[int]int64"},
		{Path: []*fieldPathNode{{key: "s", isMap: true}}, Type: "map[int]string"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"array" usage:"array help: "`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "key", usage: "key help"},
		}, Type: "string"},
		{Path: []*fieldPathNode{
			{key: "array", isSlice: true, usage: "array help: "},
			{key: "value", usage: "value help"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Map map[string]struct {
			Value string `yaml:"value" usage:"value help"`
			Key   string `yaml:"key" usage:"key help"`
		} `yaml:"map" usage:"map help: "`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "key", usage: "key help"},
		}, Type: "string"},
		{Path: []*fieldPathNode{
			{key: "map", isMap: true, usage: "map help: "},
			{key: "value", usage: "value help"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		NoHelp string `yaml:"nohelp"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "nohelp"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		IgnoreMe string `usage:"should be ignored"`
	}{}, nil)

	testConfig_collectInfos(t, &dynamicHelp{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "string", usage: "Dynamic help for String"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &defaultHelp{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "string"},
		}, Type: "string", DefaultValue: "Default value"},
	})

	// will not work, but will also not panic
	testConfig_collectInfos(t, &defaultWithoutPointerHelp{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "string"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &defaultHelp{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "string"},
		}, Type: "string", DefaultValue: "Default value via external function"},
	}, WithDefaults(func(d *defaultHelp) {
		d.String = "Default value via external function"
	}))

	testConfig_collectInfos(t, &struct {
		Map map[string]struct {
			Value string `yaml:"value" short:"v"`
		} `yaml:"map"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "map", isMap: true},
			{key: "value"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Value string `yaml:"value" short:"v"`
		} `yaml:"array"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "array", isSlice: true},
			{key: "value"},
		}, Type: "string"},
	})

	testConfig_collectInfos(t, &struct {
		Entry struct {
			Values []string `yaml:"value" short:"v"`
		} `yaml:"entry"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "entry"},
			{key: "value", isSlice: true},
		}, Type: "[]string", Short: "v"},
	})

	testConfig_collectInfos(t, &struct {
		Array []struct {
			Array []struct {
				Value string `yaml:"value" short:"v"`
			} `yaml:"array"`
		} `yaml:"array"`
	}{}, []fieldInfo{
		{Path: []*fieldPathNode{
			{key: "array", isSlice: true},
			{key: "array", isSlice: true},
			{key: "value"},
		}, Type: "string"},
	})
}

func testConfig_collectInfos[T any](t *testing.T, dst *T, expected []fieldInfo, opts ...Option) {
	t.Run("", func(t *testing.T) {
		r := NewConfig(dst, opts...).collectInfos()
		assert.Equal(t, expected, r.fi)
	})
}
