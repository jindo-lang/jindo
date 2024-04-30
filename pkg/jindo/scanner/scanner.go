// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package scanner implements a scanner for Jindo source text.
// It takes a []byte as source which can then be tokenized
// through repeated calls to the Scan method.
package scanner

import (
	"fmt"
	"io"
	"jindo/pkg/jindo/token"
	"unicode"
	"unicode/utf8"
)

type Scanner struct {
	source
	semi bool

	Line, Col uint // position
	Token     token.Token
	Lit       string         // valid if tok istoken.Name,token.Literal, ortoken.Semi ("semicolon", "newline", or "EOF"); may be malformed if bad is true
	Bad       bool           // valid if tok istoken.Literal, true if a syntax error occurred, Lit may be malformed
	Kind      token.LitKind  // valid if tok istoken.Literal
	Op        token.Operator // valid if tok istoken.Operator,token.AssignOp, ortoken.IncOp
	Prec      int            // valid if tok istoken.Operator,token.AssignOp, ortoken.IncOp
}

const EOFCHAR = '\000'

func (l *Scanner) Init(r io.Reader, errh func(line, col uint, msg string)) {
	l.source.init(r, errh)
	l.semi = false
}

func (l *Scanner) errorf(format string, args ...any) {
	l.error(fmt.Sprintf(format, args...))
}

// errorAtf reports an error at a byte column offset relative to the current Token start.
func (l *Scanner) errorAtf(offset uint, format string, args ...any) {
	l.errh(l.Line, l.Col+offset, fmt.Sprintf(format, args...))
}

func (l *Scanner) ident() {
	for isLetter(l.ch) || isDecimal(l.ch) {
		l.nextch()
	}

	lit := l.Segment()
	tok := token.Keyword(string(lit))
	if tok.IsKeyword() {
		l.Token = tok
		return
	}
	l.semi = true
	l.Lit = string(lit)
	l.Token = token.Name
}

func (l *Scanner) setLit(kind token.LitKind, ok bool) {
	l.semi = true
	l.Token = token.Literal
	l.Lit = string(l.Segment())
	l.Bad = !ok
	l.Kind = kind
}

func (l *Scanner) Next() {
	semi := l.semi
	l.semi = false

	//redo:
	//iLine, iCol := l.pos()

	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' && !semi || l.ch == '\r' {
		l.nextch()
	}
	l.source.start()

	l.Line, l.Col = l.pos()

	if isLetter(l.ch) || l.ch >= utf8.RuneSelf && l.atIdentChar(true) {
		l.ident()
		return
	}

	switch l.ch {
	case -1:
		if semi {
			l.Lit = "\n"
			l.Token = token.Semi
			break
		}
		l.Token = token.EOF

	case '\n':
		l.nextch()
		l.Lit = "\n"
		l.Token = token.Semi

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
		l.Token = token.Lparen

	case '[':
		l.nextch()
		l.Token = token.Lbrack

	case '{':
		l.nextch()
		l.Token = token.Lbrace

	case ',':
		l.nextch()
		l.Token = token.Comma

	case ';':
		l.nextch()
		l.Lit = "semicolon"
		l.Token = token.Semi

	case ')':
		l.nextch()
		l.semi = true
		l.Token = token.Rparen

	case ']':
		l.nextch()
		l.semi = true
		l.Token = token.Rbrack

	case '}':
		l.nextch()
		l.semi = true
		l.Token = token.Rbrace

	case ':':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Token = token.Define
			break
		}
		l.Token = token.Colon

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
				l.Token = token.DotDotDot
				break
			}
			l.rewind() // now s.ch holds 1st '.'
			l.nextch() // consume 1st '.' again
		}
		l.Token = token.Dot

	case '+':
		l.nextch()
		l.Op, l.Prec = token.Add, token.PrecAdd
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.Token = token.IncOp
	case '-':
		l.nextch()
		l.Op, l.Prec = token.Sub, token.PrecAdd
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.Token = token.IncOp
	case '*':
		l.nextch()
		l.Op, l.Prec = token.Mul, token.PrecMul
		if l.ch != '*' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.Token = token.IncOp
	case '/':
		l.nextch()
		if l.ch == '/' {
			for l.ch != '\n' && l.ch != -1 {
				l.nextch()
			}
			l.Next()
			return
		}
		l.Op, l.Prec = token.Div, token.PrecMul
		if l.ch != '+' {
			goto assignoper
		}
		l.nextch()
		l.semi = true
		l.Token = token.IncOp
	case '%':
		l.nextch()
		l.Op, l.Prec = token.Rem, token.PrecMul
		goto assignoper
	case '<':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Op, l.Prec = token.Leq, token.PrecCmp
			l.Token = token.Op
			break
		}
		//if l.ch == '<' {
		//	l.nextch()
		//	l.op, l.prec = Shl,Token.PrecMul
		//	goto assignoper
		//}
		l.Op, l.Prec = token.Lss, token.PrecCmp
		l.Token = token.Op

	case '>':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Op, l.Prec = token.Geq, token.PrecCmp
			l.Token = token.Op
			break
		}
		//if l.ch == '>' {
		//	l.nextch()
		//	l.op, l.prec = Shr,Token.PrecMul
		//	goto assignoper
		//}
		l.Op, l.Prec = token.Gtr, token.PrecCmp
		l.Token = token.Op

	case '=':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Op, l.Prec = token.Eql, token.PrecCmp
			l.Token = token.Op
			break
		}
		l.Token = token.Assign

	case '!':
		l.nextch()
		if l.ch == '=' {
			l.nextch()
			l.Op, l.Prec = token.Neq, token.PrecCmp
			l.Token = token.Op
			break
		}
		l.Op, l.Prec = token.Not, 0
		l.Token = token.Op
	}

	return

assignoper:
	if l.ch == '=' {
		l.nextch()
		l.Token = token.AssignOp
		return
	}
	l.Token = token.Op

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
	seg := l.Lit
	seg = seg[1 : len(seg)-1]
	str := make([]byte, 0)
	for i := 0; i < len(seg); i++ {
		if seg[i] == '\\' {
			_len := len(seg)
			if _len < i+2 {
				panic("invalid string Lit")
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
				panic("invalid string Lit")
			}
			i++
			continue
		}
		str = append(str, seg[i])
	}
	l.Lit = string(str)
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
	l.Lit = l.Lit[1 : len(l.Lit)-1]
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isLetter(ch rune) bool  { return unicode.IsLetter(ch) || ch == '_' }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
