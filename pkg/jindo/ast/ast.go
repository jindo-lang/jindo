// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package ast

import (
	"jindo/pkg/jindo/position"
	"jindo/pkg/jindo/token"
)

type Node interface {
	GetPos() position.Pos
	aNode()
	SetPos(pos position.Pos)
}

type node struct {
	Pos position.Pos
}

func (n *node) GetPos() position.Pos { return n.Pos }
func (*node) aNode()                 {}
func (n *node) SetPos(pos position.Pos) {
	n.Pos = pos
}

type File struct {
	SpaceName *Name
	DeclList  []Decl
	EOF       position.Pos
	node
}

// Top Level Declarations
type (
	Decl interface {
		Node
		aDecl()
	}

	//              Path
	ImportDecl struct {
		Group *Group    // nil means not part of a group
		Path  *BasicLit // Path.Bad || Path.Kind == StringLit; nil means no path
		decl
	}

	OperDecl struct {
		Group        *Group
		TypeL, TypeR *Field
		Oper         token.Operator
		Return       Expr
		Body         *BlockStmt
		decl
	}

	TypeDecl struct {
		Group *Group
		Name  *Name
		Alias bool
		Type  Expr
		decl
	}

	VarDecl struct {
		Group    *Group // nil means not part of a group
		NameList *Name
		Type     Expr // nil means no type
		Values   Expr // nil means no values
		decl
	}

	FuncDecl struct {
		Group  *Group // nil means not part of a group
		Param  []*Field
		Name   *Name // identifier
		Return Expr  // nil means no return type
		Body   *BlockStmt
		decl
	}
)

type decl struct{ node }

func (*decl) aDecl() {}

func NewName(pos position.Pos, value string) *Name {
	n := new(Name)
	n.Pos = pos
	n.Value = value
	return n
}

type StmtType uint8

const (
	ExprSt StmtType = iota
	EmptySt
	IncDecSt
	ContinueSt
	BreakSt
	ReturnSt
	DeclSt
	DefineSt
	AssignSt
	IfSt
	ForSt
	WhileSt
	BlockSt
)

type (
	Stmt interface {
		Node
		aStmt()
		StmtType() StmtType
	}

	SimpleStmt interface {
		Stmt
		aSimpleStmt()
	}

	ExprStmt struct {
		X Expr
		simpleStmt
	}

	EmptyStmt struct {
		simpleStmt
	}

	IncDecStmt struct {
		X   Expr
		Tok token.Token
		simpleStmt
	}

	ContinueStmt struct {
		simpleStmt
	}

	BreakStmt struct {
		simpleStmt
	}

	ReturnStmt struct {
		Result Expr
		stmt
	}

	DeclStmt struct {
		DeclList []Decl
		stmt
	}

	DefineStmt struct {
		Lhs Expr
		Rhs Expr
		simpleStmt
	}

	AssignStmt struct {
		Lhs Expr
		Op  token.Operator
		Rhs Expr
		simpleStmt
	}

	IfStmt struct {
		Cond  Expr
		Block *BlockStmt
		Else  Stmt
		stmt
	}

	ForStmt struct {
		Init SimpleStmt
		Cond Expr
		Post SimpleStmt
		Body *BlockStmt
		stmt
	}

	WhileStmt struct {
		Cond Expr
		Body *BlockStmt
		stmt
	}

	simpleStmt struct {
		stmt
	}

	BlockStmt struct {
		StmtList []Stmt
		Rbrace   position.Pos
		stmt
	}
)

func (s *stmt) StmtType() StmtType {
	//TODO implement me
	panic("implement me")
}

type stmt struct {
	node
	_type StmtType
}

func (*stmt) aStmt() {}

type (
	Expr interface {
		Node
		aExpr()
	}

	// Placeholder for an expression that failed to parse
	// correctly and where we can't provide a better node.
	BadExpr struct {
		reason string
		expr
	}

	// Value
	Name struct {
		Value string
		expr
	}

	// Value
	BasicLit struct {
		Value string
		Kind  token.LitKind
		Bad   bool // true means the gotLiteral Value has syntax errors
		expr
	}

	SliceLit struct {
		ElemType Expr
		Elems    []Expr
		expr
	}

	Operation struct {
		Op   token.Operator
		X, Y Expr // Y == nil means unary expression
		expr
	}

	ParenExpr struct {
		X Expr
		expr
	}
	SliceType struct {
		Elem Expr
		expr
	}

	// X.Sel
	SelectorExpr struct {
		X   Expr
		Sel *Name
		expr
	}

	IndexExpr struct {
		X     Expr
		Index Expr
		expr
	}

	// Func(ArgList[0], ArgList[1], ...)
	CallExpr struct {
		Func    Expr
		ArgList []Expr // nil means no arguments
		expr
	}

	Field struct {
		Name *Name // nil means anonymous field/parameter (structs/parameters), or embedded element (interfaces)
		Type Expr  // field names declared in a list share the same Type (identical pointers)
		expr
	}
)

func (simpleStmt) aSimpleStmt() {}

type expr struct{ node }

func (*expr) aExpr() {}

type Group struct {
	_ int // not empty so we are guaranteed different Group instances
}
