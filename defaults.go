package conf

import "reflect"

type DefaultSetter interface {
	SetDefaults()
}

// ApplyDefaults applies default values (execute all DefaultSetters) to the fields of the destination struct.
func (c *Config) ApplyDefaults() {
	v := reflect.ValueOf(c.dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	c.applyDefaultsRecursive(v.Type(), v)
}

func (c *Config) applyDefaultsRecursive(t reflect.Type, v reflect.Value) {
	addr := v.Addr().Interface()
	if addr != nil {
		if setter, ok := addr.(DefaultSetter); ok {
			setter.SetDefaults()
		}
		if setter := c.options.defaultSetter[t]; setter != nil {
			setter(addr)
		}
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldValue := v.Field(i)

		switch field.Type.Kind() {
		case reflect.Struct:
			c.applyDefaultsRecursive(field.Type, fieldValue)
		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct && !fieldValue.IsNil() {
				c.applyDefaultsRecursive(field.Type.Elem(), fieldValue.Elem())
			}
		default:
			// ignore other types
		}
	}
}

func (c *Config) getDefaultValue(parentType reflect.Type, field reflect.StructField) (any, bool) {
	typeVal := reflect.New(parentType).Interface()

	appliedDefaults := false
	if setter, ok := typeVal.(DefaultSetter); ok {
		setter.SetDefaults()
		appliedDefaults = true
	}
	if setter, ok := c.options.defaultSetter[parentType]; ok {
		setter(typeVal)
		appliedDefaults = true
	}

	if appliedDefaults {
		userDefinedDefaultValue := reflect.ValueOf(typeVal).Elem().FieldByName(field.Name).Interface()

		typeVal = reflect.New(parentType).Interface()
		goDefaultValue := reflect.ValueOf(typeVal).Elem().FieldByName(field.Name).Interface()

		if userDefinedDefaultValue != goDefaultValue {
			return userDefinedDefaultValue, true
		}
	}

	return nil, false
}
