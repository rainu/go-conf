package yacl

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

	for _, info := range f.fi {
		short := info.short
		if short != "" {
			short = f.options.prefixShort + short
			short += ", "
		}
		long := f.options.prefixLong + info.path.key(f.options, "i", "k")
		if strings.HasPrefix(info.sType, "[]") {
			// we can dismiss the slice key in case there is a slice of primitives
			long = strings.TrimSuffix(long, ".[i]")
		}
		long = strings.ReplaceAll(long, ".[", "[")

		table.Append([]string{
			short,
			long,
			strings.TrimPrefix(info.sType, "*"), // remove pointer prefix
			info.path.Usage(),
		})

		if info.defaultValue != nil {
			table.Append([]string{
				"",
				"",
				"",
				fmt.Sprintf("Default: %v", info.defaultValue),
			})
		}
	}

	table.Render()
	return sb.String()
}

func (f *fieldInfos) HelpYaml() string {
	fakeArgs := make([]string, 0, len(f.fi))

	for _, fInfo := range f.fi {
		arg := f.options.prefixLong
		arg += fInfo.path.key(f.options, "0", "k")
		arg += string(f.options.assignSign)
		arg += strings.TrimPrefix(fInfo.sType, "*") // remove pointer prefix

		help := fInfo.path.Usage()
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
