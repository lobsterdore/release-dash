package templatefns

import (
	"html/template"
	"math"
	"reflect"
)

var TemplateFnsMap = template.FuncMap{
	"divisibleby": func(arg interface{}, value interface{}) bool {
		var v float64
		switch value.(type) {
		case int, int8, int16, int32, int64:
			v = float64(reflect.ValueOf(value).Int())
		case uint, uint8, uint16, uint32, uint64:
			v = float64(reflect.ValueOf(value).Uint())
		case float32, float64:
			v = reflect.ValueOf(value).Float()
		default:
			return false
		}

		var a float64
		switch arg.(type) {
		case int, int8, int16, int32, int64:
			a = float64(reflect.ValueOf(arg).Int())
		case uint, uint8, uint16, uint32, uint64:
			a = float64(reflect.ValueOf(arg).Uint())
		case float32, float64:
			a = reflect.ValueOf(arg).Float()
		default:
			return false
		}

		return math.Mod(v, a) == 0
	},
	"dividetoint": func(dividend int, divisor int) int {
		return int(math.RoundToEven(float64(dividend) / float64(divisor)))
	},
}
