package conf

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"reflect"
	"strings"
)

type UsageProvider interface {
	GetUsage(field string) string
}

func (c *Config) getUsage(t reflect.Type, field reflect.StructField) (usage string) {
	valAsInterface := reflect.New(t).Interface()
	if provider, ok := valAsInterface.(UsageProvider); ok {
		usage = provider.GetUsage(field.Name)
	}
	if provide, ok := c.options.usageProvider[t]; ok {
		usage = provide(valAsInterface, field.Name)
	}

	if usage == "" {
		usage = field.Tag.Get(c.options.usageTag)
	}

	return
}

func (f fieldPath) Usage() string {
	var sb strings.Builder

	for _, pc := range f {
		sb.WriteString(pc.usage)
	}

	return sb.String()
}

func (f *fieldInfos) HelpFlags() string {
	var sb strings.Builder

	table := tablewriter.NewWriter(&sb)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoWrapText(false)

	for _, property := range f.fi {
		short := property.short
		if short != "" {
			short = f.options.prefixShort + short
			short += ", "
		}
		long := f.options.prefixLong + property.path.key(f.options, "i", "k")
		if strings.HasPrefix(property.sType, "[]") {
			// we can dismiss the slice key in case there is a slice of primitives
			long = strings.TrimSuffix(long, ".[i]")
		}
		long = strings.ReplaceAll(long, ".[", "[")

		table.Append([]string{
			short,
			long,
			property.sType,
			property.path.Usage(),
		})

		if property.defaultValue != nil {
			table.Append([]string{
				"",
				"",
				"",
				fmt.Sprintf("Default: %v", property.defaultValue),
			})
		}
	}

	table.Render()
	return sb.String()
}

func (f *fieldInfos) HelpYaml() string {
	fakeArgs := make([]string, 0, len(f.fi))

	for _, property := range f.fi {
		arg := f.options.prefixLong
		arg += property.path.key(f.options, "0", "k")
		arg += string(f.options.assignSign)
		arg += property.sType

		help := property.path.Usage()
		if help != "" {
			arg += " # " + help
		}
		fakeArgs = append(fakeArgs, arg)
	}

	r := newReaderWithoutSort(fakeArgs, nil, f.options)
	defer r.Close()

	c, _ := io.ReadAll(r)

	return string(c)
}
