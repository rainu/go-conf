package conf

import (
	"reflect"
	"slices"
	"strings"
)

type fieldInfo struct {
	Path         fieldPath
	Short        string
	DefaultValue any
	Type         string
}

type fieldPath []*fieldPathElement

type fieldPathElement struct {
	key     string
	isMap   bool
	isSlice bool
	usage   string
}

type fieldInfos struct {
	fi      []fieldInfo
	options Options
}

func (c *Config) collectInfos() fieldInfos {
	properties := fieldInfos{
		options: c.options,
	}

	c.scan(reflect.TypeOf(c.dest), []*fieldPathElement{}, &properties.fi)
	slices.SortFunc(properties.fi, func(a, b fieldInfo) int {
		return strings.Compare(a.Path.key(c.options, "i"), b.Path.key(c.options, "i"))
	})

	return properties
}

func (c *Config) scan(t reflect.Type, parent fieldPath, infos *[]fieldInfo) {
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

		pc := fieldPathElement{}
		subPath := slices.Clone(parent)
		subPath = append(subPath, &pc)

		pc.key = strings.Split(yamlTag, ",")[0]
		pc.usage = c.getUsage(t, field)

		shortTag := field.Tag.Get(c.options.shortTag)

		switch field.Type.Kind() {
		case reflect.Struct:
			c.scan(field.Type, subPath, infos)
		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct {
				c.scan(field.Type.Elem(), subPath, infos)
			}
		case reflect.Slice, reflect.Array:
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				pc.isSlice = true
				c.scan(elemType, subPath, infos)
			} else {
				// for slices of primitives, we just add the fieldInfo
				info := fieldInfo{
					Path:  subPath,
					Short: shortTag,
					Type:  "[]" + field.Type.Elem().Kind().String(),
				}
				*infos = append(*infos, info)
			}
		case reflect.Map:
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				pc.isMap = true
				c.scan(elemType, subPath, infos)
			} else {
				// for maps of primitives, we just add the fieldInfo
				property := fieldInfo{
					Path:  subPath,
					Short: shortTag,
					Type:  "map[" + field.Type.Key().Kind().String() + "]" + field.Type.Elem().Kind().String(),
				}
				*infos = append(*infos, property)
			}
		default:
			property := fieldInfo{
				Path:  subPath,
				Short: shortTag,
				Type:  field.Type.Kind().String(),
			}
			if defValue, ok := c.getDefaultValue(t, field); ok {
				property.DefaultValue = defValue
			}

			*infos = append(*infos, property)
		}
	}
}

func (p fieldPath) key(opts Options, sliceKey string) string {
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
			sb.WriteString("[k]")
		}
	}

	return sb.String()
}

func (p fieldInfos) findByShort(key string) *fieldInfo {
	for _, property := range p.fi {
		if property.Short == key {
			return &property
		}
	}
	return nil
}
