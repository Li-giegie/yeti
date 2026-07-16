package yeti

import (
	"fmt"
	"reflect"
	"strconv"
)

func toString(a any) ([]string, error) {
	rv, ok := a.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(a)
	}
	if !rv.IsValid() {
		return nil, nil
	}
	switch rv.Kind() {
	case reflect.String:
		return []string{rv.String()}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(rv.Int(), 10)}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(rv.Uint(), 10)}, nil
	case reflect.Float32, reflect.Float64:
		return []string{strconv.FormatFloat(rv.Float(), 'g', -1, 64)}, nil
	case reflect.Bool:
		return []string{strconv.FormatBool(rv.Bool())}, nil
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return nil, nil
		}
		return toString(rv.Elem())
	case reflect.Slice, reflect.Array:
		if rv.IsNil() {
			return nil, nil
		}
		list := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			res, err := toString(rv.Index(i))
			if err != nil {
				return nil, err
			}
			list = append(list, res...)
		}
		return list, nil
	}
	return nil, fmt.Errorf("unsupported type: %T", a)
}
