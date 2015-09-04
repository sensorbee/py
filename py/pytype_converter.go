package py

import (
	"fmt"
	"pfi/sensorbee/sensorbee/data"
)

func getNewPyDic(m map[string]interface{}) Object {
	return Object{}
}

func newPyObj(v data.Value) Object {
	var result interface{}
	switch v.Type() {
	case data.TypeBool:
		result, _ = data.AsBool(v)
	case data.TypeInt:
		result, _ = data.AsInt(v)
	case data.TypeFloat:
		result, _ = data.AsFloat(v)
	case data.TypeString:
		result, _ = data.AsString(v)
	case data.TypeBlob:
		result, _ = data.AsBlob(v)
	case data.TypeTimestamp:
		result, _ = data.ToInt(v)
	case data.TypeArray:
		innerArray, _ := data.AsArray(v)
		result = newPyArray(innerArray)
	case data.TypeMap:
		innerMap, _ := data.AsMap(v)
		result = newPyMap(innerMap)
	case data.TypeNull:
		result = nil
	default:
		//do nothing
	}
	fmt.Println(result)
	return Object{}
}

func newPyArray(a data.Array) []interface{} {
	result := make([]interface{}, len(a))
	for i, v := range a {
		value := newPyObj(v)
		result[i] = value
	}
	return result
}

func newPyMap(m data.Map) interface{} {
	result := map[string]interface{}{}
	for k, v := range m {
		value := newPyObj(v)
		result[k] = value
	}
	return result
}
