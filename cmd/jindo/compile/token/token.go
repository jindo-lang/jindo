// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package token

type Token uint8

type token = Token

const (
	_   token = iota
	EOF       // EOF

	// names and literals
	Name    // name
	Literal // literal

	// operators and operations
	// Operator is excluding '*' (Star)
	Op       // op
	AssignOp // op=
	IncOp    // opop
	Assign   // =
	Define   // :=
	Star     // *

	// delimiters
	Lparen    // (
	Lbrack    // [
	Lbrace    // {
	Rparen    // )
	Rbrack    // ]
	Rbrace    // }
	Comma     // ,
	Semi      // ;
	Colon     // :
	Dot       // .
	DotDotDot // ...

	// keywords
	keyword_beg
	Break    // break
	Const    // const
	Continue // continue
	While
	Else   // else
	For    // for
	Func   // func
	If     // if
	Import // import
	Space  // space
	Return // return
	Type   // type
	Var    // var
	Oper   // oper
	keyword_end

	tokenCount
)

func (t token) IsKeyword() bool { return t > keyword_beg && t < keyword_end }

// Make sure we have at most 64 tokens so we can use them in a set.
const _ uint64 = 1 << (tokenCount - 1)

// Contains reports whether tok is in tokset.
func Contains(tokset uint64, tok token) bool {
	return tokset&(1<<tok) != 0
}

type LitKind uint8

// TODO(gri) With the 'i' (imaginary) suffix now permitted on integer
// and floating-point numbers, having a single ImagLit does
// not represent the literal kind well anymore. Remove it?
const (
	IntLit LitKind = iota
	FloatLit
	ImagLit
	RuneLit
	StringLit
)
