// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package ast declares the types used to represent syntax trees for Jindo
// packages.

package ast

import (
	"jindo/pkg/jindo/scanner"
	"jindo/pkg/jindo/token"
)

// ----------------------------------------------------------------------------
// Interfaces
//
// There are 3 main classes of nodes: Expressions and type nodes,
// statement nodes, and declaration nodes. The node names usually
// match the corresponding Jindo spec production names to which they
// correspond. The node fields correspond to the individual parts
// of the respective productions.
//
// All nodes contain position information marking the beginning of
// the corresponding source text segment; it is accessible via the
// Pos accessor method. Nodes may contain additional position info
// for language constructs where comments may be found between parts
// of the construct (typically any larger, parenthesized subpart).
// That position information is needed to properly position comments
// when printing the construct.

//// All node types implement the Node interface.
//type Node interface {
//	Pos() token.Pos // position of first character belonging to the node
//	End() token.Pos // position of first character immediately after the node
//}
//
//// All expression nodes implement the Expr interface.
//type Expr interface {
//	Node
//	exprNode()
//}
//
//// All statement nodes implement the Stmt interface.
//type Stmt interface {
//	Node
//	stmtNode()
//}
//
//// All declaration nodes implement the Decl interface.
//type Decl interface {
//	Node
//	declNode()
//}

type Node interface {
	GetPos() scanner.Pos
	aNode()
	SetPos(pos scanner.Pos)
}

type node struct {
	pos scanner.Pos
}

func (n *node) GetPos() scanner.Pos { return n.pos }
func (*node) aNode()                {}
func (n *node) SetPos(pos scanner.Pos) {
	n.pos = pos
}

type File struct {
	SpaceName *Name
	DeclList  []Decl
	EOF       scanner.Pos
	node
}

// Top Level Declarations
type (
	Decl interface {
		Node
		aDecl()
	}

	OperDecl struct {
		Group        *Group
		TypeL, TypeR *Field
		Oper         token.Token
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

func NewName(pos scanner.Pos, value string) *Name {
	n := new(Name)
	n.pos = pos
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
		Return Expr
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
		Rbrace   scanner.Pos
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

	BinaryExpr interface {
		Node
		aBinExpr()
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
		binExpr
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

type binExpr struct{ node }

func (*binExpr) aBinExpr() {}

func (*binExpr) aExpr() {}

type Group struct {
	_ int // not empty so we are guaranteed different Group instances
}
