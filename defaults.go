package conf

import "reflect"

// ApplyDefaults applies default values (execute all DefaultSetters) to the fields of the destination struct.
func (c *Config) ApplyDefaults() {
	v := reflect.ValueOf(c.dest)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	c.applyDefaultsRecursive(v.Type(), v)
}

func (c *Config) applyDefaultsRecursive(t reflect.Type, v reflect.Value) {
	if setter := c.options.defaultSetter[t]; setter != nil {
		addr := v.Addr().Interface()
		if addr != nil {
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
