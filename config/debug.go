package config

import (
	"fmt"
	"reflect"
	"strings"
)

// PrintConsolidatedConfig prints out the definitions of all config
func (m *Module) PrintConsolidatedConfig() {
	for _, typ := range m.configDefs {
		if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
			typ = typ.Elem()
		}
		m.printFieldsWithTags(typ, 0)
	}
}

func (m *Module) printFieldsWithTags(typ reflect.Type, indent int) {
	if indent == 0 {
		fmt.Printf("[%s]\n", typ)
	}
	prefix := strings.Repeat(" ", indent)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Tag.Get("json")
		if fieldName == "" {
			fieldName = field.Name
		}
		fieldTyp := field.Type
		if fieldTyp.Kind() == reflect.Ptr && fieldTyp.Elem().Kind() == reflect.Struct {
			fieldTyp = fieldTyp.Elem()
		}
		if fieldTyp.Kind() == reflect.Struct {
			fmt.Printf("%s%s:\n", prefix, fieldName)
			m.printFieldsWithTags(fieldTyp, indent+4)
			return
		}
		fmt.Printf("%s%s: %s\n", prefix, fieldName, goTypeToStr(field.Type))
	}
	fmt.Println()
}

func goTypeToStr(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return goTypeToStr(t.Elem())
	case reflect.Slice:
		typ := goTypeToStr(t.Elem())
		typ += "[]"
		return typ
	case reflect.Struct:
		return t.Name()
	case reflect.Interface:
		return "any"
	case reflect.Map:
		k := goTypeToStr(t.Key())
		e := goTypeToStr(t.Elem())
		return fmt.Sprintf("{ [key: %s]: %s; }", k, e)
	default:
		return t.Name()
	}
}
