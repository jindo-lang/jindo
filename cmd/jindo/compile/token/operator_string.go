// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package token

// operators
var opString = [...]string{
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
	Mul:    "*",
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

func (op Operator) String() string { return opString[op] }

// operator overload
var opOverMap = map[string]Operator{
	"not": Not,
	"add": Add,
	"sub": Sub,
	"mul": Mul,
	"div": Div,
	"eql": Eql,
	"gtr": Gtr,
	"rem": Rem,

	"rnot": Not + Reverse,
	"radd": Add + Reverse,
	"rsub": Sub + Reverse,
	"rmul": Mul + Reverse,
	"rdiv": Div + Reverse,
	"reql": Eql + Reverse,
	"rgtr": Gtr + Reverse,
	"rrem": Rem + Reverse,
}

const operOverload = 1<<Not |
	1<<Add |
	1<<Sub |
	1<<Mul |
	1<<Div |
	1<<Eql |
	1<<Gtr |
	1<<Rem |
	1<<Not + Reverse |
	1<<Add + Reverse |
	1<<Sub + Reverse |
	1<<Mul + Reverse |
	1<<Div + Reverse |
	1<<Eql + Reverse |
	1<<Gtr + Reverse |
	Rem + Reverse

func OperOrNil(name string) Operator {
	for s, t := range opOverMap {
		if name == s {
			return t
		}
	}
	return NoneOp
}

func (op Operator) IsOperOverload() bool { return operOverload&op != 0 }
func (op Operator) IsReversed() bool     { return op > Reverse }
