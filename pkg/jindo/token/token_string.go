// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package token

var tokenString = map[Token]string{
	EOF: "fileOrEof",

	// names and literals
	Name:    "name",
	Literal: "Literal",

	// operators and operations
	// Operator is excluding '*' (Star)
	Op:       "op",
	AssignOp: "op=",
	IncOp:    "opop",
	Assign:   "=",
	Define:   ":=",
	Star:     "*",

	// delimiters
	Lparen:    "(",
	Lbrack:    "[",
	Lbrace:    "{",
	Rparen:    ")",
	Rbrack:    "]",
	Rbrace:    "}",
	Comma:     ",",
	Semi:      ";",
	Colon:     ":",
	Dot:       ".",
	DotDotDot: "...",

	Var:      "var",
	Const:    "const",
	Type:     "type",
	Import:   "import",
	If:       "if",
	Else:     "else",
	Space:    "space",
	Oper:     "oper",
	Func:     "func",
	Return:   "return",
	For:      "for",
	While:    "while",
	Break:    "break",
	Continue: "continue",
}

func (t Token) String() string { return tokenString[t] }
func KeywordOrName(lit string) Token {
	for tok, k := range tokenString {
		if k == lit {
			return tok
		}
	}
	return Name
}
