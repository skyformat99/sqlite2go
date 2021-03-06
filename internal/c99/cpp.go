// Copyright 2017 The C99 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// [0]: http://www.open-std.org/jtc1/sc22/wg14/www/docs/n1256.pdf
// [1]: https://www.spinellis.gr/blog/20060626/cpp.algo.pdf

package c99

import (
	"encoding/binary"
	"fmt"
	"go/token"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cznic/golex/lex"
	"github.com/cznic/ir"
	"github.com/cznic/mathutil"
	"github.com/cznic/xc"
)

const (
	maxIncludeLevel = 200 // gcc, std is 15.
)

var (
	_ tokenReader = (*cppReader)(nil)
	_ tokenReader = (*tokenBuffer)(nil)
	_ tokenWriter = (*tokenBuffer)(nil)
)

type tokenWriter interface {
	write(...xc.Token)
}

type tokenReader interface {
	read() xc.Token
	unget(xc.Token)
	ungets(...xc.Token)
}

type tokenBuffer struct {
	toks []xc.Token
	ungetBuffer

	last rune
}

func (b *tokenBuffer) write(t ...xc.Token) { b.toks = append(b.toks, t...) }

func (b *tokenBuffer) read() (t xc.Token) {
	if len(b.ungetBuffer) != 0 {
		return b.ungetBuffer.read()
	}

	if len(b.toks) == 0 {
		t.Rune = lex.RuneEOF
		return
	}

	t = b.toks[0]
	b.toks = b.toks[1:]
	if t.Rune == '#' && (b.last == '\n' || b.last == 0) {
		t.Rune = DIRECTIVE
	}
	b.last = t.Rune
	return t
}

type cppReader struct {
	decBuf []byte
	decPos token.Pos
	tu     [][]uint32
	ungetBuffer

	last rune
}

func (c *cppReader) unget(t xc.Token) { c.ungetBuffer = append(c.ungetBuffer, t) }

func (c *cppReader) read() (t xc.Token) {
	if len(c.ungetBuffer) != 0 {
		return c.ungetBuffer.read()
	}

more:
	if len(c.decBuf) == 0 {
		if len(c.tu) == 0 {
			t.Rune = lex.RuneEOF
			return t
		}

		if len(c.tu[0]) == 0 {
			c.tu = c.tu[1:]
			goto more
		}

		c.decBuf = dict.S(int(c.tu[0][0]))
		c.tu[0] = c.tu[0][1:]
		c.decPos = 0
	}

	c.decBuf, c.decPos, t = decodeToken(c.decBuf, c.decPos)
	if t.Rune == '#' && (c.last == '\n' || c.last == 0) {
		t.Rune = DIRECTIVE
	}
	c.last = t.Rune
	return t
}

type cond []condValue

func (c cond) on() bool              { return condOn[c.tos()] }
func (c cond) pop() cond             { return c[:len(c)-1] }
func (c cond) push(n condValue) cond { return append(c, n) }
func (c cond) tos() condValue        { return c[len(c)-1] }

type macro struct {
	def  xc.Token
	fp   []int
	repl []xc.Token

	fnLike   bool
	variadic bool
}

func newMacro(def xc.Token, repl []xc.Token) *macro {
	// dbg("#define %s %s", dict.S(def.Val), toksDump(repl))
	return &macro{def: def, repl: repl}
}

func (m *macro) param(ap [][]xc.Token, nm int, out *[]xc.Token) bool {
	*out = nil
	if nm == idVaArgs {
		if !m.variadic {
			return false
		}

		if i := len(m.fp); i < len(ap) {
			o := *out
			for i, v := range ap[i:] {
				if i != 0 {
					t := o[len(o)-1]
					t.Rune = ','
					t.Val = 0
					o = append(o, t)
					t.Rune = ' '
					o = append(o, t)
				}
				o = append(o, v...)
			}
			*out = o
		}
		return true
	}

	for i, v := range m.fp {
		if v == nm {
			*out = ap[i]
			return true
		}
	}
	return false
}

type nullReader struct{}

func (nullReader) Read([]byte) (int, error) { return 0, io.EOF }

type cpp struct {
	*context
	hideSet      map[int]int // name: hidden if != 0.
	includeLevel int
	lx           *lexer
	macros       map[int]*macro // name ID: macro
}

func newCPP(ctx *context) *cpp {
	lx, err := newLexer(ctx, "", 0, nullReader{})
	if err != nil {
		panic(err)
	}

	lx.context = ctx
	r := &cpp{
		context: ctx,
		hideSet: map[int]int{},
		lx:      lx,
		macros:  map[int]*macro{},
	}
	return r
}

func (c *cpp) parse(src ...Source) (tokenReader, error) {
	var (
		encBuf  []byte
		encBuf1 [30]byte // Rune, position, optional value ID.
		tokBuf  []xc.Token
		tu      [][]uint32
	)
	for _, v := range src {
		if pf := v.Cached(); pf != nil {
			tu = append(tu, pf)
			continue
		}

		sz, err := v.Size()
		if err != nil {
			return nil, err
		}

		if sz > mathutil.MaxInt {
			return nil, fmt.Errorf("%v: file too big: %v", v.Name(), sz)
		}

		r, err := v.ReadCloser()
		if err != nil {
			return nil, err
		}

		lx, err := newLexer(c.context, v.Name(), int(sz), r)
		if err != nil {
			return nil, err
		}

		if err := func() (err error) {
			defer func() {
				switch e := recover(); x := e.(type) {
				case nil:
					// nop
				case error:
					err = newPanicError(fmt.Errorf("%s: PANIC: %v\n%s", lx.lastPosition(), errString(x), debugStack()))
				default:
					err = newPanicError(fmt.Errorf("%s: PANIC: %v\n%s", lx.lastPosition(), e, debugStack()))
				}
				if e := r.Close(); e != nil && err == nil {
					err = e
				}
			}()

			var pf []uint32
			var t xc.Token
			var toks []xc.Token
			for {
				ch := lx.cppScan()
				if ch.Rune == ccEOF {
					break
				}

				tokBuf = tokBuf[:0]
				for {
					t.Char = ch
					t.Val = 0
					if ch.Rune == '\n' {
						toks = append(trimSpace(tokBuf), t)
						break
					}

					if _, ok := tokHasVal[ch.Rune]; ok {
						t.Val = dict.ID(lx.TokenBytes(nil))
					}
					tokBuf = append(tokBuf, t)

					if ch = lx.cppScan(); ch.Rune == ccEOF {
						if !c.tweaks.injectFinalNL {
							c.errPos(lx.last.Pos(), "file is missing final NL")
						}
						break

					}
				}

				var encPos token.Pos
				encBuf = encBuf[:0]
				for _, t := range toks {
					n := binary.PutUvarint(encBuf1[:], uint64(t.Rune))
					pos := t.Pos()
					n += binary.PutUvarint(encBuf1[n:], uint64(pos-encPos))
					encPos = pos
					if t.Val != 0 {
						n += binary.PutUvarint(encBuf1[n:], uint64(t.Val))
					}
					encBuf = append(encBuf, encBuf1[:n]...)
				}
				id := dict.ID(encBuf)
				if id > math.MaxUint32 {
					panic("internal error")
				}

				pf = append(pf, uint32(id))
			}
			v.Cache(pf)
			tu = append(tu, pf)
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	return &cppReader{tu: tu}, nil
}
func (c *cpp) eval(r tokenReader, w tokenWriter) (err error) {
	defer func() {
		switch e := recover(); x := e.(type) {
		case nil:
		case error:
			err = newPanicError(fmt.Errorf("%T: PANIC: %v\n%s", c, errString(x), debugStack()))
		default:
			err = newPanicError(fmt.Errorf("%T: PANIC: %v\n%s", c, e, debugStack()))
		}
	}()
	if cond := c.expand(r, w, cond(nil).push(condZero)); len(cond) != 1 || cond.tos() != condZero {
		panic(cond)
	}
	return nil
}

// [1]pg 1.
//
// expand(TS ) /* recur, substitute, pushback, rescan */
// {
// 	if TS is {} then
//		// ---------------------------------------------------------- A
// 		return {};
//
// 	else if TS is T^HS • TS’ and T is in HS then
//		//----------------------------------------------------------- B
// 		return T^HS • expand(TS’);
//
// 	else if TS is T^HS • TS’ and T is a "()-less macro" then
//		// ---------------------------------------------------------- C
// 		return expand(subst(ts(T ),{},{},HS ∪{T},{}) • TS’ );
//
// 	else if TS is T^HS •(•TS’ and T is a "()’d macro" then
//		// ---------------------------------------------------------- D
// 		check TS’ is actuals • )^HS’ • TS’’ and actuals are "correct for T"
// 		return expand(subst(ts(T ),fp(T ),actuals,(HS ∩HS’) ∪{T },{}) • TS’’);
//
//	// ------------------------------------------------------------------ E
// 	note TS must be T^HS • TS’
// 	return T^HS • expand(TS’);
// }
func (c *cpp) expand(r tokenReader, w tokenWriter, cond cond) cond {
	for {
		t := r.read()
		switch t.Rune {
		case lex.RuneEOF:
			// -------------------------------------------------- A
			return cond
		case DIRECTIVE:
			cond = c.directive(r, w, cond)
			t.Rune = '\n'
			t.Val = 0
			w.write(t)
		case IDENTIFIER:
			if !cond.on() {
				break
			}

			nm := t.Val
			if c.hideSet[nm] != 0 {
				// ------------------------------------------ B
				w.write(t)
				continue
			}

			m := c.macros[nm]
			if m != nil && !m.fnLike {
				// ------------------------------------------ C
				t.Rune = SENTINEL
				r.unget(t)
				toks := c.subst(m, nil)
				c.hideSet[nm]++
				r.ungets(c.sanitize(toks)...)
				continue
			}

			if m != nil && m.fnLike {
				// ------------------------------------------ D
				var sentinels []xc.Token
			again:
				switch t2 := r.read(); t2.Rune {
				case SENTINEL, '\n':
					sentinels = append(sentinels, t2)
					goto again
				case '(':
					// ok
				case lex.RuneEOF:
					r.ungets(sentinels...)
					w.write(t)
					continue
				case ' ':
					goto again
				default:
					r.ungets(sentinels...)
					w.write(t)
					w.write(t2)
					continue
				}

				ap := c.actuals(m, r)
				t.Rune = SENTINEL
				sentinels = append([]xc.Token{t}, sentinels...)
				toks := append(c.subst(m, ap), sentinels...)
				c.hideSet[nm]++
				r.ungets(c.sanitize(toks)...)
				continue
			}

			w.write(t)
		case SENTINEL:
			if !cond.on() {
				panic("internal error")
			}

			c.hideSet[t.Val]--
			if c.hideSet[t.Val] < 0 {
				panic(PrettyString(t))
			}
		default:
			// -------------------------------------------------- E
			if !cond.on() {
				break
			}

			w.write(t)
		}
	}
}

func (c *cpp) sanitize(toks []xc.Token) []xc.Token {
	for i, v := range toks {
		if v.Rune == IDENTIFIER && c.hideSet[v.Val] != 0 {
			toks[i].Rune = NON_REPL
		}
	}
	return toks
}

func (c *cpp) actuals(m *macro, r tokenReader) (out [][]xc.Token) {
	var lvl, n int
	for {
		t := r.read()
		if t.Rune < 0 {
			c.err(t, "unexpected EOF")
			return nil
		}

		switch t.Rune {
		case ',':
			if lvl == 0 {
				n++
				continue
			}
		case ')':
			if lvl == 0 {
				for i, v := range out {
					out[i] = trimSpace(v)
				}
				for len(out) < len(m.fp) {
					out = append(out, nil)
				}
				return out
			}

			lvl--
		case '(':
			lvl++
		}

		for len(out) <= n {
			out = append(out, []xc.Token{})
		}
		if t.Rune == '\n' {
			t.Rune = ' '
		}
		out[n] = append(out[n], t)
	}
}

func (c *cpp) expands(toks []xc.Token) (out []xc.Token) {
	//defer func(hs, in string) { dbg("Z expands(%v)\t%q\t%q", hs, in, toksDump(out)) }(hsDump(c.hideSet), toksDump(toks))
	var r, w tokenBuffer
	r.toks = toks
	c.expand(&r, &w, cond(nil).push(condZero))
	return w.toks
}

// [1]pg 2.
//
// subst(IS,FP,AP,HS,OS ) /* substitute args, handle stringize and paste */
// {
// 	if IS is {} then
//		// ---------------------------------------------------------- A
// 		return hsadd(HS,OS);
//
// 	else if IS is #•T•IS’ and T is FP[i] then
//		// ---------------------------------------------------------- B
// 		return subst(IS’,FP,AP,HS,OS • stringize(select(i,AP)));
//
// 	else if IS is ## • T • IS’ and T is FP[i] then
//	{
//		// ---------------------------------------------------------- C
// 		if select(i,AP ) is {} then /* only if actuals can be empty */
//			// -------------------------------------------------- D
// 			return subst(IS’,FP,AP,HS,OS);
// 		else
//			// -------------------------------------------------- E
// 			return subst(IS’,FP,AP,HS,glue(OS,select(i,AP)));
// 	}
//
// 	else if IS is ## • T^HS’ • IS’ then
//		// ---------------------------------------------------------- F
// 		return subst(IS’,FP,AP,HS,glue(OS,T^HS’ ));
//
// 	else if IS is T•##^HS’ • IS’ and T is FP[i] then
//	{
//		// ---------------------------------------------------------- G
// 		if select(i,AP ) is {} then /* only if actuals can be empty */
//		{
//			// -------------------------------------------------- H
// 			if IS’ is T’ • IS’’ and T’ is FP[j] then
//				// ------------------------------------------ I
// 				return subst(IS’’,FP,AP,HS,OS • select(j,AP));
// 			else
//				// ------------------------------------------ J
// 				return subst(IS’,FP,AP,HS,OS);
// 		}
//		else
//			// -------------------------------------------------- K
// 			return subst(##^HS’ • IS’,FP,AP,HS,OS • select(i,AP));
//
//	}
//
// 	else if IS is T•IS’ and T is FP[i] then
//		// ---------------------------------------------------------- L
// 		return subst(IS’,FP,AP,HS,OS • expand(select(i,AP)));
//
//	// ------------------------------------------------------------------ M
// 	note IS must be T^HS’ • IS’
// 	return subst(IS’,FP,AP,HS,OS • THS’);
// }
func (c *cpp) subst(m *macro, ap [][]xc.Token) (out []xc.Token) {
	//defer func(hs, in string) { dbg("Z subst(%v)\t%q\t%q", hs, in, toksDump(out)) }(hsDump(c.hideSet), toksDump(repl))
	repl := m.repl
	var arg []xc.Token
	for {
		if len(repl) == 0 {
			// -------------------------------------------------- A
			return trimSpace(out)
		}

		if repl[0].Rune == '#' && len(repl) > 1 && repl[1].Rune == IDENTIFIER && m.param(ap, repl[1].Val, &arg) {
			// -------------------------------------------------- B
			out = append(out, c.stringize(arg))
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == '#' && len(repl) > 2 && repl[1].Rune == ' ' && repl[2].Rune == IDENTIFIER && m.param(ap, repl[2].Val, &arg) {
			// -------------------------------------------------- B
			out = append(out, c.stringize(arg))
			repl = repl[3:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 1 && repl[1].Rune == IDENTIFIER && m.param(ap, repl[1].Val, &arg) {
			// -------------------------------------------------- C
			if len(arg) == 0 {
				// ------------------------------------------ D
				panic("TODO")
			}

			// -------------------------------------------------- E
			_, out = c.glue(out, arg)
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 2 && repl[1].Rune == ' ' && repl[2].Rune == IDENTIFIER && m.param(ap, repl[2].Val, &arg) {
			// -------------------------------------------------- C
			if len(arg) == 0 {
				// ------------------------------------------ D
				repl = repl[3:]
				continue
			}

			// -------------------------------------------------- E
			_, out = c.glue(out, arg)
			repl = repl[3:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 1 && repl[1].Rune != ' ' {
			// -------------------------------------------------- F
			_, out = c.glue(out, repl[1:2])
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == PPPASTE && len(repl) > 2 && repl[1].Rune == ' ' {
			// -------------------------------------------------- F
			_, out = c.glue(out, repl[2:3])
			repl = repl[3:]
			continue
		}

		if len(repl) > 1 && repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) && repl[1].Rune == PPPASTE {
			// -------------------------------------------------- G
			if len(arg) == 0 {
				// ------------------------------------------ H
				panic(c.position(repl[0]))
			}

			// -------------------------------------------------- K
			out = append(out, arg...)
			repl = repl[1:]
			continue
		}

		if len(repl) > 2 && repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) && repl[1].Rune == ' ' && repl[2].Rune == PPPASTE {
			// -------------------------------------------------- G
			if len(arg) == 0 {
				// ------------------------------------------ H
				if len(repl) > 3 && repl[3].Rune == IDENTIFIER && m.param(ap, repl[3].Val, &arg) {
					// ---------------------------------- I
					panic("TODO")
				}

				// ------------------------------------------ J
				repl = repl[3:]
				continue
			}

			// -------------------------------------------------- K
			out = append(out, arg...)
			repl = repl[2:]
			continue
		}

		if repl[0].Rune == IDENTIFIER && m.param(ap, repl[0].Val, &arg) {
			// -------------------------------------------------- L
			out = append(out, c.expands(arg)...)
			repl = repl[1:]
			continue
		}

		// ---------------------------------------------------------- M
		out = append(out, repl[0])
		repl = repl[1:]
	}
}

// paste last of left side with first of right side
//
// [1] pg. 3
func (c *cpp) glue(ls, rs []xc.Token) (n int, out []xc.Token) {
	//defer func(a, b string) { dbg("glue\t%q\t%q\t%q", a, b, toksDump(out)) }(toksDump(ls), toksDump(rs))
	for len(ls) != 0 && ls[len(ls)-1].Rune == ' ' {
		ls = ls[:len(ls)-1]
	}

	for len(rs) != 0 && rs[0].Rune == ' ' {
		rs = rs[1:]
		n++
	}
	if len(rs) == 0 {
		panic("TODO")
	}

	if len(ls) == 0 {
		return n, rs
	}

	l := ls[len(ls)-1]
	ls = ls[:len(ls)-1]
	r := rs[0]
	rs = rs[1:]
	n++

	switch l.Rune {
	case '#':
		switch r.Rune {
		case '#':
			l.Rune = PPPASTE
		default:
			panic(PrettyString([]xc.Token{l, r}))
		}
	default:
		l.Val = dict.SID(TokSrc(l) + TokSrc(r))
	}
	return n, append(append(ls, l), rs...)
}

// Givenatoken sequence, stringize returns a single string literal token
// containing the concatenated spellings of the tokens.
//
// [1] pg. 3
func (c *cpp) stringize(s []xc.Token) xc.Token {
	var a []string
	for _, v := range s {
		switch v.Rune {
		case CHARCONST, LONGCHARCONST, LONGSTRINGLITERAL, STRINGLITERAL:
			s := fmt.Sprintf("%q", TokSrc(v))
			a = append(a, s[1:len(s)-1])
		default:
			a = append(a, TokSrc(v))
		}
	}
	if v := dict.SID(fmt.Sprintf(`"%s"`, strings.Join(a, ""))); v != 0 {
		var t xc.Token
		if len(s) != 0 {
			t = s[0]
		}
		t.Rune = STRINGLITERAL
		t.Val = v
		return t
	}

	return xc.Token{}
}

func (c *cpp) directive(r tokenReader, w tokenWriter, cond cond) cond {
	line := c.line(r)
	if len(line) == 0 {
		return cond
	}

	switch t := line[0]; t.Rune {
	case lex.RuneEOF:
		// nop
	case IDENTIFIER:
		switch t.Val {
		case idDefine:
			if !cond.on() {
				break
			}

			if len(line) == 1 {
				c.err(t, "empty define not allowed")
				break
			}

			c.define(line[1:])
		case idElif:
			switch cond.tos() {
			case condIfOff:
				if c.constExpr(line[1:]) {
					return cond.pop().push(condIfOn)
				}
			case condIfOn:
				return cond.pop().push(condIfSkip)
			case condIfSkip:
				// nop
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cond.tos()))
			}
		case idElse:
			switch cond.tos() {
			case condIfOff:
				return cond.pop().push(condIfOn)
			case condIfOn:
				return cond.pop().push(condIfOff)
			case condIfSkip:
				// nop
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cond.tos()))
			}
		case idError:
			if !cond.on() {
				break
			}

			panic(fmt.Errorf("%v", c.position(t)))
		case idIf:
			if !cond.on() {
				return cond.push(condIfSkip)
			}

			switch {
			case c.constExpr(line[1:]):
				return cond.push(condIfOn)
			default:
				return cond.push(condIfOff)
			}
		case idIfdef:
			if !cond.on() {
				return cond.push(condIfSkip)
			}

			line = trimAllSpace(line[1:])
			if len(line) == 0 {
				c.err(t, "empty #ifdef not allowed")
				break
			}

			if len(line) > 1 {
				c.err(t, "extra tokens after #ifdef not allowed")
				break
			}

			if line[0].Rune != IDENTIFIER {
				c.err(line[0], "expected identifier")
				break
			}

			if _, ok := c.macros[line[0].Val]; ok {
				return cond.push(condIfOn)
			}

			return cond.push(condIfOff)
		case idIfndef:
			if !cond.on() {
				return cond.push(condIfSkip)
			}

			line = trimAllSpace(line[1:])
			if len(line) == 0 {
				c.err(t, "empty #ifndef not allowed")
				break
			}

			if len(line) > 1 {
				c.err(t, "extra tokens after #ifndef not allowed")
				break
			}

			if line[0].Rune != IDENTIFIER {
				c.err(line[0], "expected identifier")
				break
			}

			if _, ok := c.macros[line[0].Val]; ok {
				return cond.push(condIfOff)
			}

			return cond.push(condIfOn)
		case idInclude:
			if !cond.on() {
				break
			}

			line = c.expands(trimAllSpace(line[1:]))
			if c.tweaks.cppExpandTest {
				w.write(line...)
				break
			}

			if len(line) == 0 {
				c.err(t, "empty include not allowed")
				break
			}

			switch line[0].Rune {
			case '<':
				var nm string
				for _, v := range line[1:] {
					if v.Rune == '>' {
						c.include(t, nm, c.sysIncludePaths, w)
						return cond
					}

					nm += TokSrc(v)
				}
				c.err(t, "invalid include file name specification")
			case STRINGLITERAL:
				b := dict.S(line[0].Val)      // `"foo.h"`
				nm := string(b[1 : len(b)-1]) // `foo.h`
				c.include(t, nm, c.includePaths, w)
				return cond
			default:
				panic(PrettyString(line))
			}
		case idEndif:
			switch cond.tos() {
			case condIfOn, condIfOff, condIfSkip:
				return cond.pop()
			default:
				panic(fmt.Errorf("%v: %v", c.position(t), cond.tos()))
			}
		case idPragma:
			if !cond.on() {
				break
			}

			panic(fmt.Errorf("%v", c.position(t)))
		case idUndef:
			if !cond.on() {
				break
			}

			line = trimSpace(line[1:])
			if len(line) == 0 {
				panic("TODO")
			}

			if len(line) > 1 {
				panic("TODO")
			}

			if line[0].Rune != IDENTIFIER {
				panic("TODO")
			}

			delete(c.macros, line[0].Val)
		case idWarning:
			panic(fmt.Errorf("%v", c.position(t)))
		default:
			panic(fmt.Errorf("%v", c.position(t)))
		}
	default:
		panic(PrettyString(t))
	}
	return cond
}

func (c *cpp) include(n Node, nm string, paths []string, w tokenWriter) {
	if c.includeLevel == maxIncludeLevel {
		c.err(n, "too many include levels")
	}

	c.includeLevel++

	defer func() { c.includeLevel-- }()

	var path string
	for _, v := range paths {
		if v == "@" {
			v = filepath.Dir(c.position(n).Filename)
		}

		p := filepath.Join(v, nm)
		fi, err := os.Stat(p)
		if err != nil || fi.IsDir() {
			continue
		}

		path = p
		break
	}

	if path == "" {
		panic(c.position(n))
	}

	r, err := c.parse(newFileSource(path))
	if err != nil {
		c.err(n, "%s", err.Error())
	}

	c.expand(r, w, cond(nil).push(condZero))
}

func (c *cpp) constExpr(toks []xc.Token) (y bool) {
	toks = trimAllSpace(toks)
	for i, v := range toks {
		if v.Rune == IDENTIFIER && v.Val == idDefined {
			s := toks[i:]
			switch {
			case len(s) > 1 && s[1].Rune == IDENTIFIER:
				s[0].Rune = INTCONST
				s[0].Val = idZero
				if _, ok := c.macros[s[1].Val]; ok {
					s[0].Val = idOne
				}
				s[1].Rune = ' '
				continue
			case len(s) > 3 && s[1].Rune == '(' && s[2].Rune == IDENTIFIER && s[3].Rune == ')':
				s[0].Rune = INTCONST
				s[0].Val = idZero
				if _, ok := c.macros[s[2].Val]; ok {
					s[0].Val = idOne
				}
				s[1].Rune = ' '
				s[2].Rune = ' '
				s[3].Rune = ' '
				continue
			}
		}
	}
	toks = c.expands(trimAllSpace(toks))
	for i, v := range toks {
		if v.Rune == IDENTIFIER {
			toks[i].Rune = INTCONST
			toks[i].Val = idZero
		}
	}
	c.lx.ungetBuffer = c.lx.ungetBuffer[:0]
	c.lx.ungets(toks...)
	//defer func(n Node, in string) { dbg("%v: %q: %v", c.position(n), in, y) }(toks[0], toksDump(toks, ""))
	if !c.lx.parseExpr() {
		return false
	}

	e := c.lx.ast.(*ConstExpr)
	v := e.eval(c.context)
	if v.Type != Int {
		return false
	}

	switch x := v.Value.(type) {
	case nil:
		return false
	case *ir.Int64Value:
		return x.Value != 0
	default:
		panic(fmt.Errorf("%T", x))
	}
}

func (c *cpp) define(line []xc.Token) {
	switch line[0].Rune {
	case ' ':
		c.defineMacro(line[1:])
	default:
		panic(PrettyString(line))
	}
}

func (c *cpp) defineMacro(line []xc.Token) {
	if len(line) == 0 {
		panic("internal error")
	}

	if line[0].Rune == ' ' {
		line = line[1:]
	}

	switch t := line[0]; t.Rune {
	case IDENTIFIER:
		nm := t.Val
		line := line[1:]
		var repl []xc.Token
		if len(line) != 0 {
			switch line[0].Rune {
			case '\n', lex.RuneEOF:
				// nop
			case ' ':
				repl = line[1:]
			case '(':
				c.defineFnMacro(t, line[1:])
				return
			default:
				panic(fmt.Errorf(PrettyString(line[0])))
			}
		}

		if ex := c.macros[nm]; ex != nil {
			if c.identicalReplacementLists(repl, ex.repl) {
				return
			}

			c.err(t, "replacement lists differ")
			return
		}

		c.macros[nm] = newMacro(t, repl)
	default:
		panic(PrettyString(t))
	}
}

func (c *cpp) identicalReplacementLists(a, b []xc.Token) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		w := b[i]
		if v.Rune != w.Rune || v.Val != w.Val {
			return false
		}
	}

	return true
}

func (c *cpp) defineFnMacro(nmTok xc.Token, line []xc.Token) {
	ident := true
	var params []int
	variadic := false
	for i, v := range line {
		switch v.Rune {
		case IDENTIFIER:
			if !ident {
				panic("TODO")
			}

			params = append(params, v.Val)
			ident = false
		case ')':
			m := newMacro(nmTok, trimSpace(line[i+1:]))
			m.fnLike = true
			m.variadic = variadic
			m.fp = params
			if ex := c.macros[nmTok.Val]; ex != nil {
				if c.identicalParamLists(params, ex.fp) && c.identicalReplacementLists(m.repl, ex.repl) && m.variadic == ex.variadic {
					return
				}

				c.err(nmTok, "parameter and/or replacement lists differ")
				return
			}

			c.macros[nmTok.Val] = m
			return
		case ',':
			if ident {
				panic("TODO")
			}

			ident = true
		case ' ':
			// nop
		case DDD:
			variadic = true
		default:
			panic(PrettyString(v))
		}
	}
}

func (c *cpp) identicalParamLists(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func (c *cpp) line(r tokenReader) (toks []xc.Token) {
	for {
		switch t := r.read(); t.Rune {
		case '\n', lex.RuneEOF:
			for len(toks) != 0 && toks[0].Rune == ' ' {
				toks = toks[1:]
			}
			return toks
		default:
			toks = append(toks, t)
		}
	}
}
