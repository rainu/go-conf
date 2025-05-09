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

type fieldPath []*fieldPathNode

type fieldPathNode struct {
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
	infos := fieldInfos{
		options: c.options,
	}

	c.scan(reflect.TypeOf(c.dest), []*fieldPathNode{}, &infos.fi)
	slices.SortFunc(infos.fi, func(a, b fieldInfo) int {
		return strings.Compare(a.Path.key(c.options, "i"), b.Path.key(c.options, "i"))
	})

	//ignore short-hand for ...
	for i := range infos.fi {
		if infos.fi[i].Short == "" {
			continue
		}

		// ... nodes which are in maps
		isMap := slices.ContainsFunc(infos.fi[i].Path, func(node *fieldPathNode) bool {
			return node.isMap
		})
		if isMap {
			infos.fi[i].Short = ""
			continue
		}

		// ... nodes which are >in< slices
		isSlice := slices.ContainsFunc(infos.fi[i].Path, func(node *fieldPathNode) bool {
			return node.isSlice
		})
		if isSlice {
			// only allowed if the slice is the last node
			if infos.fi[i].Path[len(infos.fi[i].Path)-1].isSlice {
				continue
			}
			infos.fi[i].Short = ""
			continue
		}
	}

	return infos
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

		node := fieldPathNode{}
		subPath := slices.Clone(parent)
		subPath = append(subPath, &node)

		node.key = strings.Split(yamlTag, ",")[0]
		node.usage = c.getUsage(t, field)

		shortTag := field.Tag.Get(c.options.shortTag)

		switch field.Type.Kind() {
		case reflect.Struct:
			c.scan(field.Type, subPath, infos)
		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct {
				c.scan(field.Type.Elem(), subPath, infos)
			}
		case reflect.Slice, reflect.Array:
			node.isSlice = true
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				c.scan(elemType, subPath, infos)
			} else {
				// for slices of primitives, we just add the fieldInfo
				info := fieldInfo{
					Path:  subPath.purge(),
					Short: shortTag,
					Type:  "[]" + field.Type.Elem().Kind().String(),
				}
				*infos = append(*infos, info)
			}
		case reflect.Map:
			node.isMap = true
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				c.scan(elemType, subPath, infos)
			} else {
				// for maps of primitives, we just add the fieldInfo
				property := fieldInfo{
					Path:  subPath.purge(),
					Short: shortTag,
					Type:  "map[" + field.Type.Key().Kind().String() + "]" + field.Type.Elem().Kind().String(),
				}
				*infos = append(*infos, property)
			}
		default:
			property := fieldInfo{
				Path:  subPath.purge(),
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

func (p fieldPath) purge() fieldPath {
	//remove empty nodes
	return slices.DeleteFunc(p, func(node *fieldPathNode) bool {
		return node.key == ""
	})
}

func (p fieldInfos) findByShort(key string) *fieldInfo {
	for _, info := range p.fi {
		if info.Short == key {
			return &info
		}
	}
	return nil
}

func (p fieldInfos) findByPath(path []string) *fieldInfo {
	joinedPath := strings.Join(path, string(p.options.keyDelimiter))
	for _, info := range p.fi {
		keyNodes := make([]string, len(info.Path))
		for i, node := range info.Path {
			keyNodes[i] = node.key
		}
		joinedKey := strings.Join(keyNodes, string(p.options.keyDelimiter))

		if joinedKey == joinedPath {
			return &info
		}
	}
	return nil
}
