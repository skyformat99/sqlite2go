// Code generated by "stringer -output enum_string.go -type=TypeKind,condValue enum.go type.go"; DO NOT EDIT.

package c99

import "fmt"

const _TypeKind_name = "BoolCharIntLongLongLongSCharShortUCharUIntULongULongLongUShortFloatDoubleLongDoubleFloatComplexDoubleComplexLongDoubleComplexmaxTypeKind"

var _TypeKind_index = [...]uint8{0, 4, 8, 11, 15, 23, 28, 33, 38, 42, 47, 56, 62, 67, 73, 83, 95, 108, 125, 136}

func (i TypeKind) String() string {
	i -= 1
	if i < 0 || i >= TypeKind(len(_TypeKind_index)-1) {
		return fmt.Sprintf("TypeKind(%d)", i+1)
	}
	return _TypeKind_name[_TypeKind_index[i]:_TypeKind_index[i+1]]
}

const _condValue_name = "condZerocondIfOffcondIfOncondIfSkipmaxCond"

var _condValue_index = [...]uint8{0, 8, 17, 25, 35, 42}

func (i condValue) String() string {
	if i < 0 || i >= condValue(len(_condValue_index)-1) {
		return fmt.Sprintf("condValue(%d)", i)
	}
	return _condValue_name[_condValue_index[i]:_condValue_index[i+1]]
}
