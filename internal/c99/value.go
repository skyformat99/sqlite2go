// Copyright 2017 The C99 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package c99

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"fmt"
	"math"

	"github.com/cznic/ir"
)

var (
	// [0]6.3.1.1-1
	//
	// Every integer type has an integer conversion rank defined as
	// follows:
	intConvRank = [maxTypeKind]int{
		Bool:      1,
		Char:      2,
		SChar:     2,
		UChar:     2,
		Short:     3,
		UShort:    3,
		Int:       4,
		UInt:      4,
		Long:      5,
		ULong:     5,
		LongLong:  6,
		ULongLong: 6,
	}

	isSigned = [maxTypeKind]bool{
		Bool:     true,
		Char:     true,
		SChar:    true,
		Short:    true,
		Int:      true,
		Long:     true,
		LongLong: true,
	}

	isArithmeticType = [maxTypeKind]bool{
		Bool:      true,
		Char:      true,
		Int:       true,
		Long:      true,
		LongLong:  true,
		SChar:     true,
		Short:     true,
		UChar:     true,
		UInt:      true,
		ULong:     true,
		ULongLong: true,
		UShort:    true,

		Float:      true,
		Double:     true,
		LongDouble: true,

		FloatComplex:      true,
		DoubleComplex:     true,
		LongDoubleComplex: true,
	}
)

// [0]6.3.1.8
//
// Many operators that expect operands of arithmetic type cause conversions and
// yield result types in a similar way. The purpose is to determine a common
// real type for the operands and result. For the specified operands, each
// operand is converted, without change of type domain, to a type whose
// corresponding real type is the common real type. Unless explicitly stated
// otherwise, the common real type is also the corresponding real type of the
// result, whose type domain is the type domain of the operands if they are the
// same, and complex otherwise. This pattern is called the usual arithmetic
// conversions:
func usualArithmeticConversions(ctx *context, a, b *Value) (c, d *Value) {
	if !a.isArithmeticType() || !b.isArithmeticType() {
		panic("TODO")
	}

	// First, if the corresponding real type of either operand is long
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is long double.
	if a.Type == LongDoubleComplex || b.Type == LongDoubleComplex {
		return a.convertTo(ctx, LongDoubleComplex), b.convertTo(ctx, LongDoubleComplex)
	}

	if a.Type == LongDouble || b.Type == LongDouble {
		return a.convertTo(ctx, LongDouble), b.convertTo(ctx, LongDouble)
	}

	// Otherwise, if the corresponding real type of either operand is
	// double, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is double.
	if a.Type == DoubleComplex || b.Type == DoubleComplex {
		return a.convertTo(ctx, DoubleComplex), b.convertTo(ctx, DoubleComplex)
	}

	if a.Type == Double || b.Type == Double {
		return a.convertTo(ctx, Double), b.convertTo(ctx, Double)
	}

	// Otherwise, if the corresponding real type of either operand is
	// float, the other operand is converted, without change of type
	// domain, to a type whose corresponding real type is float.)
	if a.Type == FloatComplex || b.Type == FloatComplex {
		return a.convertTo(ctx, FloatComplex), b.convertTo(ctx, FloatComplex)
	}

	if a.Type == Float || b.Type == Float {
		return a.convertTo(ctx, Float), b.convertTo(ctx, Float)
	}

	// Otherwise, the integer promotions are performed on both operands.
	// Then the following rules are applied to the promoted operands:
	if !a.isIntegerType() || !b.isIntegerType() {
		panic("TODO")
	}

	a = a.integerPromotion()
	b = b.integerPromotion()

	// If both operands have the same type, then no further conversion is
	// needed.
	if a.Type == b.Type {
		return a, b
	}

	// Otherwise, if both operands have signed integer types or both have
	// unsigned integer types, the operand with the type of lesser integer
	// conversion rank is converted to the type of the operand with greater
	// rank.
	if a.isSigned() == b.isSigned() {
		t := a.Type
		if intConvRank[b.Type.Kind()] > intConvRank[a.Type.Kind()] {
			t = b.Type
		}
		return a.convertTo(ctx, t), b.convertTo(ctx, t)
	}

	panic(fmt.Errorf("TODO %v %v", a, b))
}

// Value represents the type and optionally the value of an expression.
type Value struct {
	Type Type
	ir.Value
}

func (v *Value) isArithmeticType() bool { return isArithmeticType[v.Type.Kind()] }
func (v *Value) isIntegerType() bool    { return intConvRank[v.Type.Kind()] != 0 }
func (v *Value) isSigned() bool         { return isSigned[v.Type.Kind()] }

func (v *Value) add(ctx *context, w *Value) (r *Value) {
	v, w = usualArithmeticConversions(ctx, v, w)
	if v.Value == nil || w.Value == nil {
		return &Value{Type: v.Type}
	}

	switch x := v.Value.(type) {
	case *ir.Int64Value:
		return (&Value{v.Type, &ir.Int64Value{Value: x.Value + w.Value.(*ir.Int64Value).Value}}).normalize(ctx)
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (v *Value) convertTo(ctx *context, t Type) *Value {
	if v.Type == t {
		return v
	}

	if v.Value == nil {
		return &Value{Type: t}
	}

	switch x := v.Type.Kind(); x {
	case Int:
		switch t.Kind() {
		case Long:
			return &Value{t, v.Value}
		default:
			panic(fmt.Errorf("%v -> %v", v, t))
		}
	default:
		panic(fmt.Errorf("%v -> %v", v, t))
	}
}

func (v *Value) eq(ctx *context, w *Value) (r *Value) {
	r = &Value{Type: Int}
	if v.Value == nil || w.Value == nil {
		return r
	}

	v, w = usualArithmeticConversions(ctx, v, w)
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		var val int64
		if x.Value == w.Value.(*ir.Int64Value).Value {
			val = 1
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r
}

func (v *Value) ge(ctx *context, w *Value) (r *Value) {
	r = &Value{Type: Int}
	r = &Value{Type: Int}
	if v.Value == nil || w.Value == nil {
		return r
	}

	v, w = usualArithmeticConversions(ctx, v, w)
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case v.isSigned():
			if x.Value >= w.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			panic("TODO")
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r
}

func (v *Value) gt(ctx *context, w *Value) (r *Value) {
	r = &Value{Type: Int}
	r = &Value{Type: Int}
	if v.Value == nil || w.Value == nil {
		return r
	}

	v, w = usualArithmeticConversions(ctx, v, w)
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case v.isSigned():
			if x.Value > w.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			panic("TODO")
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r
}

func (v *Value) lt(ctx *context, w *Value) (r *Value) {
	r = &Value{Type: Int}
	r = &Value{Type: Int}
	if v.Value == nil || w.Value == nil {
		return r
	}

	v, w = usualArithmeticConversions(ctx, v, w)
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		var val int64
		switch {
		case v.isSigned():
			if x.Value < w.Value.(*ir.Int64Value).Value {
				val = 1
			}
		default:
			panic("TODO")
		}
		r.Value = &ir.Int64Value{Value: val}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return r
}

func (v *Value) normalize(ctx *context) *Value {
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		val := x.Value
		switch sz := ctx.model[v.Type.Kind()].Size; sz {
		case 4:
			switch {
			case v.isSigned():
				switch {
				case val < 0:
					panic("TODO")
				default:
					x.Value = val & math.MaxUint32
				}
			default:
				panic("TODO")
			}
		default:
			panic(fmt.Errorf("TODO %v", sz))
		}
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
	return v
}

// [0]6.3.1.1-2
//
// If an int can represent all values of the original type, the value is
// converted to an int; otherwise, it is converted to an unsigned int. These
// are called the integer promotions. All other types are unchanged by the
// integer promotions.
func (v *Value) integerPromotion() *Value {
	switch v.Type.Kind() {
	case Int, Long:
		return v
	default:
		panic(v)
	}
}

func (v *Value) isNonzero() bool {
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		return x.Value != 0
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}

func (v *Value) isZero() bool {
	switch x := v.Value.(type) {
	case *ir.Int64Value:
		return x.Value == 0
	default:
		panic(fmt.Errorf("TODO %T", x))
	}
}
