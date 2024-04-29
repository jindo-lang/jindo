// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package scanner implements a scanner for Jindo source text.
// It takes a []byte as source which can then be tokenized
// through repeated calls to the Scan method.
//
package scanner

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
	"jindo/pkg/jindo/token"
)


type Scanner struct {
	source
	semi bool

	line, col uint // position
	token     token.Token
	lit       string   // valid if tok istoken.Name,token.Literal, ortoken.Semi ("semicolon", "newline", or "EOF"); may be malformed if bad is true
	bad       bool     // valid if tok istoken.Literal, true if a syntax error occurred, lit may be malformed
	kind      token.LitKind  // valid if tok istoken.Literal
	op        token.Operator // valid if tok istoken.Operator,token.AssignOp, ortoken.IncOp
	prec      int      // valid if tok istoken.Operator,token.AssignOp, ortoken.IncOp
}

const EOFCHAR = '\000'

func (l *Scanner) init(r io.Reader, errh func(line, col uint, msg string)) {
	l.source.init(r, errh)
	l.semi = false
}

func (l *Scanner) errorf(format string, args ...any) {
	l.error(fmt.Sprintf(format, args...))
}

// errorAtf reports an error at a byte column offset relative to the current Token start.
func (l *Scanner) errorAtf(offset uint, format string, args ...any) {
	l.errh(l.line, l.col+offset, fmt.Sprintf(format, args...))
}

func (l *Scanner) ident() {
	for isLetter(l.ch) || isDecimal(l.ch) {
		l.nextch()
	}

	lit := l.segment()
	tok := token.Keyword(string(lit))
	if tok.IsKeyword() {
		l.token = tok
		return
	}
	l.semi = true
	l.lit = string(lit)
	l.token =token.Name
}

func (l *Scanner) setLit(kind token.LitKind, ok bool) {
	l.semi = true
	l.token =token.Literal
	l.lit = string(l.segment())
	l.bad = !ok
	l.kind = kind
}

func (l *Scanner) next() {
	semi := l.semi
	l.semi = false

	//redo:
	//iLine, iCol := l.pos()

	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' && !semi || l.ch == '\r' {
		l.nextch()
	}
	l.source.start()

	l.line, l.col = l.pos()

	if isLetter(l.ch) || l.ch >= utf8.RuneSelf && l.atIdentChar(true) {
		l.ident()
		return
	}

	switch l.ch {
	case -1:
		if semi {
			l.lit = "EOF"
			l.token =token.Semi
			break
		}
		l.token =token.EOF

	case '\n':
		l.nextch()
		l.lit = "newline"
		l.token =token.Semi

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		l.number(false)
	case '"':
		l.stdString()
	case '`':
		l.rawString()
	case '\'':
		l.rune()
	case '(':
		l.nextch()
		l.token =token.Lparen

	case '[':
		l.nextch()
		l.token =token.Lbrack

	case '{':
		l.nextch()
		l.token =token.Lbrace

	case ',':
		l.nextch()
		l.token =token.Comma

	case ';':
		l.nextch()
		l.lit = "semicolon"
		l.token =token.Semi

	case ')':
		l.nextch()
		l.semi = true
		l.token =token.Rparen

	case ']':
		l.nextch()
		l.semi = true
		l.token =token.Rbrack

	case '}':
		l.nextch()
		l.semi = true
		l.token =token.Rbrace

	case ':':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.token =token.Define
			break
		}
		l.token =token.Colon

	case '.':
		l.nextch()
		if isDecimal(l.ch) {
			l.number(true)
			break
		}
		if l.ch == '.' {
			l.nextch()
			if l.ch == '.' {
				l.nextch()
				l.token =token.DotDotDot
				break
			}
			l.rewind() // now s.ch holds 1st '.'
			l.nextch() // consume 1st '.' again
		}
		l.token =token.Dot

	case '+':
		l.nextch()
		l.op, l.prec = token.Add,token.PrecAdd
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.token =token.IncOp
	case '-':
		l.nextch()
		l.op, l.prec = token.Sub,token.PrecAdd
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.token =token.IncOp
	case '*':
		l.nextch()
		l.op, l.prec = token.Mul,token.PrecMul
		if l.ch != '*' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.token =token.IncOp
	case '/':
		l.nextch()
		if l.ch == '/' {
			for l.ch != '\n' && l.ch != -1 {
				l.nextch()
			}
			l.next()
			return
		}
		l.op, l.prec = token.Div,token.PrecMul
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.token =token.IncOp
	case '%':
		l.nextch()
		l.op, l.prec = token.Rem,token.PrecMul
		goto assignoper
	case '<':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.op, l.prec = token.Leq,token.PrecCmp
			l.token =token.Op
			break
		}
		//if l.ch == '<' {
		//	l.nextch()
		//	l.op, l.prec = Shl,token.PrecMul
		//	goto assignoper
		//}
		l.op, l.prec = token.Lss,token.PrecCmp
		l.token =token.Op

	case '>':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.op, l.prec = token.Geq,token.PrecCmp
			l.token =token.Op
			break
		}
		//if l.ch == '>' {
		//	l.nextch()
		//	l.op, l.prec = Shr,token.PrecMul
		//	goto assignoper
		//}
		l.op, l.prec = token.Gtr,token.PrecCmp
		l.token =token.Op

	case '=':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.op, l.prec = token.Eql,token.PrecCmp
			l.token =token.Op
			break
		}
		l.token =token.Assign

	case '!':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.op, l.prec = token.Neq,token.PrecCmp
			l.token =token.Op
			break
		}
		l.op, l.prec = token.Not, 0
		l.token =token.Op
	}

	return

assignoper:
	if l.ch == '=' {
		l.nextch()
		l.token =token.AssignOp
		return
	}
	l.token =token.Op

}
func (l *Scanner) rune() {
	ok := true
	l.nextch()

	n := 0
	for ; ; n++ {
		if l.ch == '\'' {
			if ok {
				if n == 0 {
					l.errorf("empty rune literal or unescaped '")
					ok = false
				} else if n != 1 {
					l.errorAtf(0, "more than one character in rune literal")
					ok = false
				}
			}
			l.nextch()
			break
		}
		if l.ch == '\\' {
			l.nextch()
			if !l.escape('\'') {
				ok = false
			}
			continue
		}
		if l.ch == '\n' {
			if ok {
				l.errorf("newline in rune literal")
				ok = false
			}
			break
		}
		if l.ch < 0 {
			if ok {
				l.errorAtf(0, "rune literal not terminated")
				ok = false
			}
			break
		}
		l.nextch()
	}

	l.setLit(token.RuneLit, ok)
}

func (l *Scanner) stdString() {
	ok := true
	l.nextch()

	for {
		if l.ch == '"' {
			l.nextch()
			break
		}
		if l.ch == '\\' {
			l.nextch()
			if !l.escape('"') {
				ok = false
			}
			continue
		}
		if l.ch == '\n' {
			l.errorf("newline in string")
			ok = false
			break
		}
		if l.ch < 0 {
			l.errorAtf(0, "string not terminated")
			ok = false
			break
		}
		l.nextch()
	}

	l.setLit(token.StringLit, ok)
	seg := l.lit
	seg = seg[1 : len(seg)-1]
	str := make([]byte, 0)
	for i := 0; i < len(seg); i++ {
		if seg[i] == '\\' {
			_len := len(seg)
			if _len < i+2 {
				panic("invalid string lit")
			}
			switch seg[i+1] {
			case '\\':
				str = append(str, []byte("\\")...)
			case 'a':
				str = append(str, []byte("\a")...)
			case 'b':
				str = append(str, []byte("\b")...)
			case 'f':
				str = append(str, []byte("\f")...)
			case 'n':
				str = append(str, []byte("\n")...)
			case 'r':
				str = append(str, []byte("\r")...)
			case 't':
				str = append(str, []byte("\t")...)
			case 'v':
				str = append(str, []byte("\v")...)
			default:
				panic("invalid string lit")
			}
			i++
			continue
		}
		str = append(str, seg[i])
	}
	l.lit = string(str)
}

func (l *Scanner) escape(quote rune) bool {
	switch l.ch {
	case quote, '\\', 'a', 'b', 'f', 'n', 'r', 't', 'v':
		return true
	}
	return false
}

func (l *Scanner) number(afterDot bool) {
	kind := token.IntLit
	ok := true
	if !afterDot {
		for isDecimal(l.ch) {
			l.nextch()
			if l.ch == '.' {
				l.nextch()
				afterDot = true
				break
			}
		}
	}

	if afterDot {
		kind = token.FloatLit
		digitExist := false
		for isDecimal(l.ch) {
			l.nextch()
			digitExist = true
		}
		if !digitExist {
			ok = false
			l.errorf("No digit after '.'")
		}
	}
	l.setLit(kind, ok)
}

func (l *Scanner) atIdentChar(first bool) bool {
	switch {
	case unicode.IsLetter(l.ch) || l.ch == '_':
		// ok
	case unicode.IsDigit(l.ch):
		if first {
			l.errorf("identifier cannot begin with digit %#U", l.ch)
		}
	case l.ch >= utf8.RuneSelf:
		l.errorf("invalid character %#U in identifier", l.ch)
	default:
		return false
	}
	return true
}

func (l *Scanner) rawString() {
	ok := true
	l.nextch()
	for {
		if l.ch == '`' {
			l.nextch()
			break
		}
		if l.ch < 0 {
			l.errorAtf(0, "string not terminated")
			ok = false
			break
		}
		l.nextch()
	}
	l.setLit(token.StringLit, ok)
	l.lit = l.lit[1 : len(l.lit)-1]
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isLetter(ch rune) bool  { return unicode.IsLetter(ch) || ch == '_' }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
