package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io"
	"reflect"
	"slices"
	"strings"
)

type Property struct {
	Path         path
	Short        string
	DefaultValue any
	Type         string
}

type path []*pathChild

type pathChild struct {
	key     string
	isMap   bool
	isSlice bool
	usage   string
}

type Properties struct {
	p       []Property
	options Options
}

type UsageProvider interface {
	GetUsage(field string) string
}

func (c *Config) collectHelpProperties() Properties {
	properties := Properties{
		options: c.options,
	}

	c.collect(reflect.TypeOf(c.dest), []*pathChild{}, &properties.p)
	slices.SortFunc(properties.p, func(a, b Property) int {
		return strings.Compare(a.Path.key(c.options, "i"), b.Path.key(c.options, "i"))
	})

	return properties
}

func (c *Config) collect(t reflect.Type, parent path, properties *[]Property) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		pc := pathChild{}
		subPath := slices.Clone(parent)
		subPath = append(subPath, &pc)

		pc.key = strings.Split(yamlTag, ",")[0]
		pc.usage = field.Tag.Get(c.options.usageTag)

		if pc.usage == "" {
			if provider, ok := reflect.New(t).Interface().(UsageProvider); ok {
				pc.usage = provider.GetUsage(field.Name)
			}
		}

		shortTag := field.Tag.Get("short")

		switch field.Type.Kind() {
		case reflect.Struct:
			c.collect(field.Type, subPath, properties)
		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct {
				c.collect(field.Type.Elem(), subPath, properties)
			}
		case reflect.Slice, reflect.Array:
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				pc.isSlice = true
				c.collect(elemType, subPath, properties)
			} else {
				// for slices of primitives, we just add the property
				property := Property{
					Path:  subPath,
					Short: shortTag,
					Type:  "[]" + field.Type.Elem().Kind().String(),
				}
				*properties = append(*properties, property)
			}
		case reflect.Map:
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				pc.isMap = true
				c.collect(elemType, subPath, properties)
			} else {
				// for maps of primitives, we just add the property
				property := Property{
					Path:  subPath,
					Short: shortTag,
					Type:  "map[" + field.Type.Key().Kind().String() + "]" + field.Type.Elem().Kind().String(),
				}
				*properties = append(*properties, property)
			}
		default:
			property := Property{
				Path:  subPath,
				Short: shortTag,
				Type:  field.Type.Kind().String(),
			}
			if defValue, ok := c.getDefaultValue(t, field); ok {
				property.DefaultValue = defValue
			}

			*properties = append(*properties, property)
		}
	}
}

func (c *Config) getDefaultValue(parentType reflect.Type, field reflect.StructField) (any, bool) {
	typeVal := reflect.New(parentType).Interface()
	if c.options.defaultSetter[parentType] != nil {
		c.options.defaultSetter[parentType](typeVal)

		userDefinedDefaultValue := reflect.ValueOf(typeVal).Elem().FieldByName(field.Name).Interface()

		typeVal = reflect.New(parentType).Interface()
		goDefaultValue := reflect.ValueOf(typeVal).Elem().FieldByName(field.Name).Interface()

		if userDefinedDefaultValue != goDefaultValue {
			return userDefinedDefaultValue, true
		}
	}

	return nil, false
}

func (p path) key(opts Options, sliceKey string) string {
	var sb strings.Builder

	for i, pc := range p {
		if i > 0 {
			sb.WriteRune(opts.keyDelimiter)
		}
		sb.WriteString(pc.key)
		if pc.isSlice {
			sb.WriteRune(opts.keyDelimiter)
			sb.WriteRune('[')
			sb.WriteString(sliceKey)
			sb.WriteRune(']')
		} else if pc.isMap {
			sb.WriteRune(opts.keyDelimiter)
			sb.WriteString("[key]")
		}
	}

	return sb.String()
}

func (p path) Usage() string {
	var sb strings.Builder

	for _, pc := range p {
		sb.WriteString(pc.usage)
	}

	return sb.String()
}

func (p Properties) Get() []Property {
	return p.p
}

func (p Properties) HelpFlags() string {
	var sb strings.Builder

	table := tablewriter.NewWriter(&sb)
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAutoWrapText(false)

	for _, property := range p.p {
		short := property.Short
		if short != "" {
			short = p.options.prefixShort + short
			short += ", "
		}
		defValue := ""
		if property.DefaultValue != nil {
			defValue = fmt.Sprintf("(default: %v)", property.DefaultValue)
		}

		table.Append([]string{
			short,
			p.options.prefixLong + property.Path.key(p.options, "i"),
			property.Type,
			defValue,
			property.Path.Usage(),
		})
	}

	table.Render()
	return sb.String()
}

func (p Properties) HelpYaml() string {
	fakeArgs := make([]string, 0, len(p.p))

	for _, property := range p.p {
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

	r := newReader(fakeArgs, p.options)
	defer r.Close()

	c, _ := io.ReadAll(r)

	return string(c)
}
