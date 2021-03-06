// Copyright 2017 The C99 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package c99

import (
	"encoding/binary"
	"go/token"
	"reflect"
	"strconv"

	"github.com/cznic/golex/lex"
	"github.com/cznic/ir"
	"github.com/cznic/strutil"
	"github.com/cznic/xc"
)

var (
	dict       = xc.Dict
	printHooks = strutil.PrettyPrintHooks{}
)

func init() {
	for k, v := range xc.PrintHooks {
		printHooks[k] = v
	}
	delete(printHooks, reflect.TypeOf(token.Pos(0)))
	lcRT := reflect.TypeOf(lex.Char{})
	lcH := func(f strutil.Formatter, v interface{}, prefix, suffix string) {
		c := v.(lex.Char)
		r := c.Rune
		s := yySymName(int(r))
		if x := s[0]; x >= '0' && x <= '9' {
			s = strconv.QuoteRune(r)
		}
		f.Format(prefix)
		f.Format("%s", s)
		f.Format(suffix)
	}

	printHooks[lcRT] = lcH
	printHooks[reflect.TypeOf(xc.Token{})] = func(f strutil.Formatter, v interface{}, prefix, suffix string) {
		t := v.(xc.Token)
		if (t == xc.Token{}) {
			return
		}

		lcH(f, t.Char, prefix, "")
		if s := t.S(); len(s) != 0 {
			f.Format(" %q", s)
		}
		f.Format(suffix)
	}
	for _, v := range []interface{}{
		(*ir.Float32Value)(nil),
		(*ir.Float64Value)(nil),
		(*ir.Int32Value)(nil),
		(*ir.Int64Value)(nil),
		(*ir.StringValue)(nil),
		ExprCase(0),
		TypeKind(0),
	} {
		printHooks[reflect.TypeOf(v)] = func(f strutil.Formatter, v interface{}, prefix, suffix string) {
			f.Format(prefix)
			f.Format("%v", v)
			f.Format(suffix)
		}
	}
}

var (
	idDefine  = dict.SID("define")
	idDefined = dict.SID("defined")
	idElif    = dict.SID("elif")
	idElse    = dict.SID("else")
	idEndif   = dict.SID("endif")
	idError   = dict.SID("error")
	idIf      = dict.SID("if")
	idIfdef   = dict.SID("ifdef")
	idIfndef  = dict.SID("ifndef")
	idInclude = dict.SID("include")
	idOne     = dict.SID("1")
	idPragma  = dict.SID("pragma")
	idUndef   = dict.SID("undef")
	idVaArgs  = dict.SID("__VA_ARGS__")
	idWarning = dict.SID("warning")
	idZero    = dict.SID("0")

	keywords = map[int]rune{
		dict.SID("_Bool"):    BOOL,
		dict.SID("_Complex"): COMPLEX,
		dict.SID("auto"):     AUTO,
		dict.SID("break"):    BREAK,
		dict.SID("case"):     CASE,
		dict.SID("char"):     CHAR,
		dict.SID("const"):    CONST,
		dict.SID("continue"): CONTINUE,
		dict.SID("default"):  DEFAULT,
		dict.SID("do"):       DO,
		dict.SID("double"):   DOUBLE,
		dict.SID("else"):     ELSE,
		dict.SID("enum"):     ENUM,
		dict.SID("extern"):   EXTERN,
		dict.SID("float"):    FLOAT,
		dict.SID("for"):      FOR,
		dict.SID("goto"):     GOTO,
		dict.SID("if"):       IF,
		dict.SID("inline"):   INLINE,
		dict.SID("int"):      INT,
		dict.SID("long"):     LONG,
		dict.SID("register"): REGISTER,
		dict.SID("restrict"): RESTRICT,
		dict.SID("return"):   RETURN,
		dict.SID("short"):    SHORT,
		dict.SID("signed"):   SIGNED,
		dict.SID("sizeof"):   SIZEOF,
		dict.SID("static"):   STATIC,
		dict.SID("struct"):   STRUCT,
		dict.SID("switch"):   SWITCH,
		dict.SID("typedef"):  TYPEDEF,
		dict.SID("union"):    UNION,
		dict.SID("unsigned"): UNSIGNED,
		dict.SID("void"):     VOID,
		dict.SID("volatile"): VOLATILE,
		dict.SID("while"):    WHILE,
	}

	tokConstVals = map[rune]int{
		ADDASSIGN: dict.SID("+="),
		ANDAND:    dict.SID("&&"),
		ANDASSIGN: dict.SID("&="),
		ARROW:     dict.SID("->"),
		AUTO:      dict.SID("auto"),
		BOOL:      dict.SID("_Bool"),
		BREAK:     dict.SID("break"),
		CASE:      dict.SID("case"),
		CHAR:      dict.SID("char"),
		COMPLEX:   dict.SID("_Complex"),
		CONST:     dict.SID("const"),
		CONTINUE:  dict.SID("continue"),
		DDD:       dict.SID("..."),
		DEC:       dict.SID("--"),
		DEFAULT:   dict.SID("default"),
		DIVASSIGN: dict.SID("/="),
		DO:        dict.SID("do"),
		DOUBLE:    dict.SID("double"),
		ELSE:      dict.SID("else"),
		ENUM:      dict.SID("enum"),
		EQ:        dict.SID("=="),
		EXTERN:    dict.SID("extern"),
		FLOAT:     dict.SID("float"),
		FOR:       dict.SID("for"),
		GEQ:       dict.SID(">="),
		GOTO:      dict.SID("goto"),
		IF:        dict.SID("if"),
		INC:       dict.SID("++"),
		INLINE:    dict.SID("inline"),
		INT:       dict.SID("int"),
		LEQ:       dict.SID("<="),
		LONG:      dict.SID("long"),
		LSH:       dict.SID("<<"),
		LSHASSIGN: dict.SID("<<="),
		MODASSIGN: dict.SID("%="),
		MULASSIGN: dict.SID("*="),
		NEQ:       dict.SID("!="),
		ORASSIGN:  dict.SID("|="),
		OROR:      dict.SID("||"),
		PPPASTE:   dict.SID("##"),
		REGISTER:  dict.SID("register"),
		RESTRICT:  dict.SID("restrict"),
		RETURN:    dict.SID("return"),
		RSH:       dict.SID(">>"),
		RSHASSIGN: dict.SID(">>="),
		SHORT:     dict.SID("short"),
		SIGNED:    dict.SID("signed"),
		SIZEOF:    dict.SID("sizeof"),
		STATIC:    dict.SID("static"),
		STRUCT:    dict.SID("struct"),
		SUBASSIGN: dict.SID("-="),
		SWITCH:    dict.SID("switch"),
		TYPEDEF:   dict.SID("typedef"),
		TYPEOF:    dict.SID("typeof"),
		UNION:     dict.SID("union"),
		UNSIGNED:  dict.SID("unsigned"),
		VOID:      dict.SID("void"),
		VOLATILE:  dict.SID("volatile"),
		WHILE:     dict.SID("while"),
		XORASSIGN: dict.SID("^="),
	}

	tokHasVal = map[rune]struct{}{
		CHARCONST:         {},
		FLOATCONST:        {},
		IDENTIFIER:        {},
		INTCONST:          {},
		LONGCHARCONST:     {},
		LONGSTRINGLITERAL: {},
		NON_REPL:          {},
		PPNUMBER:          {},
		STRINGLITERAL:     {},
		TYPEDEF_NAME:      {},
	}
)

func isUCNDigit(r rune) bool {
	return int(r) < len(ucnDigits)<<bitShift && ucnDigits[uint(r)>>bitShift]&(1<<uint(r&bitMask)) != 0
}

func isUCNNonDigit(r rune) bool {
	return int(r) < len(ucnNonDigits)<<bitShift && ucnNonDigits[uint(r)>>bitShift]&(1<<uint(r&bitMask)) != 0
}

func rune2class(r rune) (c int) {
	switch {
	case r == lex.RuneEOF:
		return ccEOF
	case r < 128:
		return int(r)
	case isUCNDigit(r):
		return ccUCNDigit
	case isUCNNonDigit(r):
		return ccUCNNonDigit
	default:
		return ccOther
	}
}

func decodeToken(b []byte, pos token.Pos) ([]byte, token.Pos, xc.Token) {
	r, n := binary.Uvarint(b)
	b = b[n:]
	d, n := binary.Uvarint(b)
	b = b[n:]
	np := pos + token.Pos(d)
	c := lex.NewChar(np, rune(r))
	var v uint64
	if _, ok := tokHasVal[c.Rune]; ok {
		v, n = binary.Uvarint(b)
		b = b[n:]
	}
	return b, np, xc.Token{Char: c, Val: int(v)}
}

// TokSrc returns t in its source form.
func TokSrc(t xc.Token) string {
	if x, ok := tokConstVals[t.Rune]; ok {
		return string(dict.S(x))
	}

	if _, ok := tokHasVal[t.Rune]; ok {
		return string(t.S())
	}

	return string(t.Rune)
}
