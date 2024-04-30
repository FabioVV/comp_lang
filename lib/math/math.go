package math

import (
	Object "github/FabioVV/interp_lang/object"
)

var Math = map[string]*Object.Lib{
	"PI": {
		Fn: &Object.Float{Value: 3.141592},
	},
}
