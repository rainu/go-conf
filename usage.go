package yacl

import (
	"fmt"
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

	maxShortLen := 0
	intend := "  "
	shortIntend := ""
	shortLongDelimiter := ", "

	for _, info := range f.fi {
		if maxShortLen < len(info.short) {
			maxShortLen = len(info.short)
		}
	}
	if maxShortLen > 0 {
		maxShortLen += len(f.options.prefixShort) + len(intend)
		shortIntend = strings.Repeat(" ", maxShortLen)
	}

	for _, info := range f.fi {
		short := info.short
		if short != "" {
			short = f.options.prefixShort + short
			short += shortLongDelimiter
		} else {
			short = shortIntend
		}

		long := f.options.prefixLong + info.path.key(f.options, "int")
		if strings.HasPrefix(info.sType, "[]") {
			// we can dismiss the slice key in case there is a slice of primitives
			long = strings.TrimSuffix(long, ".[int]")
		}
		long = strings.ReplaceAll(long, ".[", "[")
		long += string(f.options.assignSign)

		if strings.HasPrefix(info.sType, "map[") {
			// only show the value-type of the map
			valueType := info.Field().Type.Elem().Kind().String()
			if valueType == "interface" {
				valueType = "any"
			}

			long += valueType
		} else {
			long += strings.TrimPrefix(info.sType, "*") // remove pointer prefix
		}

		sb.WriteString(intend)
		sb.WriteString(short)
		sb.WriteString(long)
		sb.WriteString("\n")
		sb.WriteString(intend)
		sb.WriteString(shortIntend)
		sb.WriteString("\t")
		sb.WriteString(strings.ReplaceAll(info.path.Usage(), "\n", "\n"+intend+shortIntend+"\t"))
		if info.defaultValue != nil {
			sb.WriteString("\n")
			sb.WriteString(intend)
			sb.WriteString(shortIntend)
			sb.WriteString("\t")
			sb.WriteString(fmt.Sprintf("Default: %v", info.defaultValue))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *fieldInfos) HelpYaml() string {
	fakeArgs := make([]string, 0, len(f.fi))

	for _, fInfo := range f.fi {
		arg := f.options.prefixLong
		arg += fInfo.path.key(f.options, "0")
		arg += string(f.options.assignSign)

		if strings.HasPrefix(fInfo.sType, "map[") {
			// only show the value-type of the map
			valueType := fInfo.Field().Type.Elem().Kind().String()
			if valueType == "interface" {
				valueType = "any"
			}

			arg += valueType
		} else {
			arg += strings.TrimPrefix(fInfo.sType, "*") // remove pointer prefix
		}

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
