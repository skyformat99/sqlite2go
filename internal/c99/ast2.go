// Copyright 2017 The C99 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package c99

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"

	"github.com/cznic/ir"
)

// Node represents an AST node.
type Node interface {
	Pos() token.Pos
}

func (n *ConstExpr) eval(ctx *context) *Value {
	if n.Value == nil {
		n.Value = n.Expr.eval(ctx)
	}
	return n.Value
}

func (n *Expr) eval(ctx *context) *Value {
	if n.Value != nil {
		return n.Value
	}

	switch n.Case {
	case ExprPreInc: // "++" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprPreDec: // "--" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprSizeOfType: // "sizeof" '(' TypeName ')'
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprSizeofExpr: // "sizeof" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprNot: // '!' Expr
		n.Value = &Value{Type: Int}
		a := n.Expr.eval(ctx)
		if a.isZero() {
			n.Value.Value = &ir.Int64Value{Value: 1}
			break
		}

		if a.isNonzero() {
			n.Value.Value = &ir.Int64Value{Value: 0}
		}
	case ExprAddrof: // '&' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprPExprList: // '(' ExprList ')'
		n.Value = n.ExprList.eval(ctx)
	case ExprCompLit: // '(' TypeName ')' '{' InitializerList CommaOpt '}'
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprCast: // '(' TypeName ')' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprDeref: // '*' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprUnaryPlus: // '+' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprUnaryMinus: // '-' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprCpl: // '~' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprChar: // CHARCONST
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprNe: // Expr "!=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprModAssign: // Expr "%=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLAnd: // Expr "&&" Expr
		n.Value = &Value{Type: Int}
		a := n.Expr.eval(ctx)
		if a.isZero() {
			n.Value.Value = &ir.Int64Value{Value: 0}
			break
		}

		b := n.Expr2.eval(ctx)
		if b.isZero() {
			n.Value.Value = &ir.Int64Value{Value: 0}
			break
		}

		if a.isNonzero() && b.isNonzero() {
			n.Value.Value = &ir.Int64Value{Value: 1}
		}
	case ExprAndAssign: // Expr "&=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprMulAssign: // Expr "*=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprPostInt: // Expr "++"
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprAddAssign: // Expr "+=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprPostDec: // Expr "--"
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprSubAssign: // Expr "-=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprPSelect: // Expr "->" IDENTIFIER
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprDivAssign: // Expr "/=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLsh: // Expr "<<" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLshAssign: // Expr "<<=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLe: // Expr "<=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprEq: // Expr "==" Expr
		n.Value = n.Expr.eval(ctx).eq(ctx, n.Expr2.eval(ctx))
	case ExprGe: // Expr ">=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprRsh: // Expr ">>" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprRshAssign: // Expr ">>=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprXorAssign: // Expr "^=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprOrAssign: // Expr "|=" Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLOr: // Expr "||" Expr
		n.Value = &Value{Type: Int}
		a := n.Expr.eval(ctx)
		if a.isNonzero() {
			n.Value.Value = &ir.Int64Value{Value: 1}
			break
		}

		b := n.Expr2.eval(ctx)
		if b.isNonzero() {
			n.Value.Value = &ir.Int64Value{Value: 1}
			break
		}

		if a.isZero() && b.isZero() {
			n.Value.Value = &ir.Int64Value{Value: 0}
		}
	case ExprMod: // Expr '%' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprAnd: // Expr '&' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprCall: // Expr '(' ArgumentExprListOpt ')'
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprMul: // Expr '*' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprAdd: // Expr '+' Expr
		n.Value = n.Expr.eval(ctx).add(ctx, n.Expr2.eval(ctx))
	case ExprSub: // Expr '-' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprSelect: // Expr '.' IDENTIFIER
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprDiv: // Expr '/' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLt: // Expr '<' Expr
		n.Value = n.Expr.eval(ctx).lt(ctx, n.Expr2.eval(ctx))
	case ExprAssign: // Expr '=' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprGt: // Expr '>' Expr
		n.Value = n.Expr.eval(ctx).gt(ctx, n.Expr2.eval(ctx))
	case ExprCond: // Expr '?' ExprList ':' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprIndex: // Expr '[' ExprList ']'
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprXor: // Expr '^' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprOr: // Expr '|' Expr
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprFloat: // FLOATCONST
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprIdent: // IDENTIFIER
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprInt: // INTCONST
		s0 := string(dict.S(n.Token.Val))
		s := s0
	loop:
		for i := len(s) - 1; i > 0; i-- {
			switch s0[i] {
			case 'l', 'L', 'u', 'U':
				s = s[:i]
			default:
				break loop
			}
		}
		decadic := s == "0" || !strings.HasPrefix(s, "0")
		v, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
		}

		// [0]6.4.4.1
		switch suff := strings.ToUpper(s0[len(s):]); {
		case suff == "" && decadic:
			n.Value = ctx.newIntConstValue(n, v, Int, Long, LongLong)
		case suff == "L" && decadic:
			n.Value = ctx.newIntConstValue(n, v, Long, LongLong)
		default:
			panic(fmt.Errorf("%v: TODO %q %q decadic: %v\n%s", ctx.fset.Position(n.Pos()), s, suff, decadic, PrettyString(n)))
		}
	case ExprLChar: // LONGCHARCONST
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprLString: // LONGSTRINGLITERAL
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	case ExprString: // STRINGLITERAL
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	default:
		panic(fmt.Errorf("%v: TODO\n%s", ctx.fset.Position(n.Pos()), PrettyString(n)))
	}
	return n.Value
}

func (n *ExprList) eval(ctx *context) *Value {
	if n.Value == nil {
		for l := n; l != nil; l = l.ExprList {
			n.Value = l.Expr.eval(ctx)
		}
	}
	return n.Value
}
