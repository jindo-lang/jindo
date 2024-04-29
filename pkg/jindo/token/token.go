package token

func (t Token) IsReversedOper() bool {
	return t > Reversed_oper && t < Operator_end
}

// Token is the set of lexical tokens of the Go programming language.
type Token int

const (
	_    Token = iota
	EOF       // EOF

	// names and literals
	Name    // name
	Literal // Literal

	// operators and operations
	// _Operator is excluding '*' (_Star)
	Op // op
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
	Comment   // //

	// keywords
	keyword_beg
	Import // import
	If     // if
	Else   // else
	Space  // space
	Var    // var
	Const  // const
	Type   // type
	Oper   // oper
	Func   // func
	Return // return
	For    // for
	While  // while
	Break  // break

	Operator_beg
	OperNot       // !
	OperAdd       // add
	OperSub       // sub
	OperMul       // mul
	OperDiv       // div
	OperRem       // rem
	OperEql       // eql
	OperGtr       // gtr
	Reversed_oper // rem
	OperRAdd      // radd
	OperRSub      // rsub
	OperRMul      // rmul
	OperRDiv      // rdiv
	OperRRem      // rrem
	OperREql      // reql
	OperRGtr      // rgtr
	Operator_end
	keyword_end
)

//	// keywords
//	_Case        // case
//	_Chan        // chan
//	_Const       // const
//	_Continue    // continue
//	_Default     // default
//	_Defer       // defer
//	_Fallthrough // fallthrough
//	_Go          // go
//	_Goto        // goto
//	_Interface   // interface
//	_Map         // map
//	_Range       // range
//	_Select      // select
//	_Struct      // struct
//	_Switch      // switch
//
//	// empty line comment to exclude it from .String
//	tokenCount //
//)

func (t Token) String() string {
	return tokenString[t]
}

var tokenString = map[Token]string{
	EOF: "EOF",

	// names and literals
	Name:    "name",
	Literal: "Literal",

	// operators and operations
	// _Operator is excluding '*' (_Star)
	Op: "op",
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
	Comment:   "//",

	Var:     "var",
	Const:   "const",
	Type:    "type",
	Import:  "import",
	If:      "if",
	Else:    "else",
	Space:   "space",
	Oper:    "oper",
	Func:    "func",
	Return:  "return",
	For:     "for",
	While:   "while",
	Break:   "break",
	OperNot:  "not",
	OperAdd:  "add",
	OperSub:  "sub",
	OperMul:  "mul",
	OperDiv:  "div",
	OperEql:  "eql",
	OperGtr:  "gtr",
	OperRem:  "rem",
	OperRAdd: "radd",
	OperRSub: "rsub",
	OperRMul: "rmul",
	OperRDiv: "rdiv",
	OperRRem: "rrem",
	OperREql: "reql",
	OperRGtr: "rgtr",
}

var KeywordToken map[Token]string

func (t Token) IsKeyword() bool {
	return t > keyword_beg && t < keyword_end
}

func (t Token) IsOperator() bool {
	return t > Operator_beg && t < Operator_end
}

func Keyword(word string) Token {
	for tok, str := range tokenString {
		if str == word {
			return tok
		}
	}
	return Name
}

type LitKind int

const (
	IntLit LitKind = iota
	FloatLit
	RuneLit
	StringLit
)

func (t Token) OperTokenToOperator() Operator {
	switch t {
	case OperEql:
		return Eql // ==
	case OperGtr:
		return Gtr // >
	case OperAdd:
		return Add // +
	case OperSub:
		return Sub // -
	case OperMul:
		return Mul // *
	case OperDiv:
		return Div // /
	case OperRem:
		return Rem // %
	}
	return 0
}

type Operator int

const (
	NoneOP Operator = iota

	// Def is the : in :=
	Def // :
	Not // !

	// precOrOr
	OrOr // ||

	// precAndAnd
	AndAnd // &&

	// precCmp
	Eql // ==
	Neq // !=
	Lss // <
	Leq // <=
	Gtr // >
	Geq // >=

	// precAdd
	Add // +
	Sub // -
	//Or  // |
	//Xor // ^

	// precMul
	Mul // *
	Div // /
	Rem // %
	//And    // &
	//AndNot // &^
	//Shl    // <<
	//Shr    // >>
)

var _op = [...]string{
	Def:    ":",
	Not:    "!",
	OrOr:   "||",
	AndAnd: "&&",
	Eql:    "==",
	Neq:    "!=",
	Lss:    "<",
	Leq:    "<=",
	Gtr:    ">",
	Geq:    ">=",
	Add:    "+",
	Sub:    "-",
	//Or:     "|",
	//Xor:    "^",
	//Mul:    "*",
	//Div:    "/",
	Rem: "%",
	//And:    "&",
	//AndNot: "&^",
	//Shl:    "<<",
	//Shr:    ">>",
}

func (o Operator) String() string {
	return _op[o]
}

// Operator precedences
const (
	_ = iota
	PrecOrOr
	PrecAndAnd
	PrecCmp
	PrecAdd
	PrecMul
)
