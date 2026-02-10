// utils/borders.go
// last modification: Feb 07 2026
package utils

import (
	"os"
	"reflect"
)

/*
ExpandEnvInStruct recursively expands environment variables in all string fields
of structs, slices, maps, and their nested contents using os.ExpandEnv.
*/
func ExpandEnvInStruct(v any) {
	if v == nil {
		return
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	expandValue(val.Elem())
}

func expandValue(val reflect.Value) {
	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			if f.CanSet() {
				expandValue(f)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			expandValue(val.Index(i))
		}
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			if v.Kind() == reflect.String {
				newStr := os.ExpandEnv(v.String())
				val.SetMapIndex(k, reflect.ValueOf(newStr))
			} else {
				expandValue(v)
			}
			// Keys are usually static strings, we leave them untouched.
			_ = k
		}
	case reflect.String:
		val.SetString(os.ExpandEnv(val.String()))
	}
}