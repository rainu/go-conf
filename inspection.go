package yacl

import (
	"reflect"
	"slices"
	"strings"
)

type FieldInfo interface {
	// Path returns the path to the field in the destination struct.
	Path() string

	// Field returns the corresponding field in the destination struct.
	Field() reflect.StructField
}

type FieldInfos interface {
	// Infos returns a list of all relevant fields which are defined in the destination struct.
	Infos() []FieldInfo
}

type fieldInfo struct {
	path         fieldPath
	short        string
	defaultValue any
	sType        string
	field        reflect.StructField
}

type fieldPath []*fieldPathNode

type fieldPathNode struct {
	key        string
	isMap      bool
	mapKeyType reflect.Type
	isSlice    bool
	usage      string
}

type fieldInfos struct {
	fi      []fieldInfo
	options Options
}

// CollectInfos returns a list of all fields which are defined in the destination struct.
func (c *Config) CollectInfos() FieldInfos {
	return c.collectInfos()
}

func (c *Config) collectInfos() *fieldInfos {
	infos := fieldInfos{
		options: c.options,
	}

	c.scan(reflect.TypeOf(c.dest), []*fieldPathNode{}, &infos.fi)

	//ignore short-hand for ...
	for i := range infos.fi {
		if infos.fi[i].short == "" {
			continue
		}

		// ... nodes which are in maps
		isMap := slices.ContainsFunc(infos.fi[i].path, func(node *fieldPathNode) bool {
			return node.isMap
		})
		if isMap {
			infos.fi[i].short = ""
			continue
		}

		// ... nodes which are >in< slices
		isSlice := slices.ContainsFunc(infos.fi[i].path, func(node *fieldPathNode) bool {
			return node.isSlice
		})
		if isSlice {
			// only allowed if the slice is the last node
			if infos.fi[i].path[len(infos.fi[i].path)-1].isSlice {
				continue
			}
			infos.fi[i].short = ""
			continue
		}
	}

	return &infos
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
			} else {
				// for pointers to primitives, we just add the fieldInfo
				info := fieldInfo{
					path:  subPath.purge(),
					short: shortTag,
					sType: "*" + field.Type.Elem().Kind().String(),
					field: field,
				}
				*infos = append(*infos, info)
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
					path:  subPath.purge(),
					short: shortTag,
					sType: "[]" + field.Type.Elem().Kind().String(),
					field: field,
				}
				*infos = append(*infos, info)
			}
		case reflect.Map:
			node.isMap = true
			node.mapKeyType = field.Type.Key()
			if field.Type.Elem().Kind() == reflect.Struct ||
				(field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct) {
				elemType := field.Type.Elem()
				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
				}
				c.scan(elemType, subPath, infos)
			} else {
				// for maps of primitives, we just add the fieldInfo
				fInfo := fieldInfo{
					path:  subPath.purge(),
					short: shortTag,
					sType: "map[" + field.Type.Key().Kind().String() + "]" + field.Type.Elem().Kind().String(),
					field: field,
				}
				*infos = append(*infos, fInfo)
			}
		default:
			fInfo := fieldInfo{
				path:  subPath.purge(),
				short: shortTag,
				sType: field.Type.Kind().String(),
				field: field,
			}
			if defValue, ok := c.getDefaultValue(t, field); ok {
				fInfo.defaultValue = defValue
			}

			*infos = append(*infos, fInfo)
		}
	}
}

func (f fieldPath) key(opts Options, sliceKey string) string {
	var sb strings.Builder

	for i, pc := range f {
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
			sb.WriteRune('[')
			sb.WriteString(pc.mapKeyType.String())
			sb.WriteRune(']')
		}
	}

	return sb.String()
}

func (f fieldPath) purge() fieldPath {
	//remove empty nodes
	return slices.DeleteFunc(f, func(node *fieldPathNode) bool {
		return node.key == ""
	})
}

func (f *fieldInfos) Infos() []FieldInfo {
	result := make([]FieldInfo, 0, len(f.fi))
	for i := range f.fi {
		result = append(result, &f.fi[i])
	}
	return result
}

func (f *fieldInfos) findByShort(key string) *fieldInfo {
	for _, info := range f.fi {
		if info.short == key {
			return &info
		}
	}
	return nil
}

func (f *fieldInfos) findByPath(path []string) *fieldInfo {
	joinedPath := strings.Join(path, string(f.options.keyDelimiter))
	for _, info := range f.fi {
		keyNodes := make([]string, len(info.path))
		for i, node := range info.path {
			keyNodes[i] = node.key
		}
		joinedKey := strings.Join(keyNodes, string(f.options.keyDelimiter))

		if joinedKey == joinedPath {
			return &info
		}
	}
	return nil
}

func (f *fieldInfo) Path() string {
	return f.path.key(newDefaultOptions(), "")
}

func (f *fieldInfo) Field() reflect.StructField {
	return f.field
}
