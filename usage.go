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

func (p fieldPath) Usage() string {
	var sb strings.Builder

	for _, pc := range p {
		sb.WriteString(pc.usage)
	}

	return sb.String()
}

func (p fieldInfos) HelpFlags() string {
	var sb strings.Builder

	table := tablewriter.NewWriter(&sb)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoWrapText(false)

	for _, property := range p.fi {
		short := property.Short
		if short != "" {
			short = p.options.prefixShort + short
			short += ", "
		}
		long := p.options.prefixLong + property.Path.key(p.options, "i")
		if strings.HasPrefix(property.Type, "[]") {
			// we can dismiss the slice key in case there is a slice of primitives
			long = strings.TrimSuffix(long, ".[i]")
		}
		long = strings.ReplaceAll(long, ".[", "[")

		table.Append([]string{
			short,
			long,
			property.Type,
			property.Path.Usage(),
		})

		if property.DefaultValue != nil {
			table.Append([]string{
				"",
				"",
				"",
				fmt.Sprintf("Default: %v", property.DefaultValue),
			})
		}
	}

	table.Render()
	return sb.String()
}

func (p fieldInfos) HelpYaml() string {
	fakeArgs := make([]string, 0, len(p.fi))

	for _, property := range p.fi {
		arg := p.options.prefixLong
		arg += property.Path.key(p.options, "0")
		arg += string(p.options.assignSign)
		arg += property.Type

		help := property.Path.Usage()
		if help != "" {
			arg += " # " + help
		}
		fakeArgs = append(fakeArgs, arg)
	}

	r := newReader(fakeArgs, nil, p.options)
	defer r.Close()

	c, _ := io.ReadAll(r)

	return string(c)
}
