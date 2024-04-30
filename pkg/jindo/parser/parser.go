// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package parser implements a parser for Jindo source files. Input may be
// provided in a variety of forms (see the various Parse* functions); the
// output is an abstract syntax tree (AST) representing the Jindo source. The
// parser is invoked through one of the Parse* functions.
package parser

import (
	"fmt"
	"io"
	"jindo/pkg/jindo/ast"
	"os"
	"strconv"
	"strings"

	"jindo/pkg/jindo/scanner"
	"jindo/pkg/jindo/token"
)

var Reset string = "\033[0m"
var Cyan string = "\033[36m"
var Green string = "\033[32m"
var Yellow string = "\033[33m"
var White string = "\033[97m"
var Blue string = "\033[34m"
var Red string = "\033[31m"
var Purple string = "\033[35m"
var _ error = Error{}

type Error struct {
	Pos scanner.Pos
	Msg string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s %s", e.Pos, e.Msg)
}

type ErrHandler func(err error)

func Parse(file *scanner.PosBase, src io.Reader, errh ErrHandler, verbose bool) *ast.File {
	var p parser
	p.verbose = verbose
	p.init(file, src, errh)
	p.Next()
	return p.EOF()
}

type parser struct {
	file *scanner.PosBase
	errh ErrHandler
	scanner.Scanner
	base    *scanner.PosBase
	indent  string
	first   error
	verbose bool
}

// nil means error has occured
func (p *parser) EOF() *ast.File {
	if p.verbose {
		defer p.trace("file")()
	}

	// SourceFile = Space ";" { TopLevelDecl ";" } .
	f := new(ast.File)
	f.SetPos(p.pos())
	if !p.got(token.Space) {
		fmt.Println("expected space, got '" + p.Token.String() + "'")
		os.Exit(-1)
		return nil
	}
	f.SpaceName = p.name()
	p.print("space: " + f.SpaceName.Value)
	p.want(token.Semi)

	// TopLevelDecl = Declaration | FuncDecl | OperDecl .
	for p.Token != token.EOF {
		switch p.Token {
		case token.Type:
			p.Next()
			f.DeclList = p.appendGroup(f.DeclList, p.typeDecl)

		case token.Var:
			p.Next()
			f.DeclList = p.appendGroup(f.DeclList, p.varDecl)

		case token.Func:
			p.Next()
			f.DeclList = p.appendGroup(f.DeclList, p.funcDeclOrNil)

		case token.Oper:
			p.Next()
			f.DeclList = p.appendGroup(f.DeclList, p.operDecl)

		case token.Semi:
			p.Next()

		default:
			str := p.Token.String()
			if p.Token == token.Name {
				str += "(" + string(p.Segment()) + ")"
			}
			p.errorAt(p.pos(), "ERROR: non-declaration statement outside function body: "+Red+str+Reset)
			p.Next()
		}
	}
	return f
}

func (p *parser) trace(msg string) func() {
	p.print(msg + ":")
	const tab = "    "
	p.indent += tab

	return func() {
		p.indent = p.indent[:len(p.indent)-len(tab)]
		if err := recover(); err != nil {
			panic(err)
		}
		//p.print("")
	}
}

var line = -1

func (p *parser) print(msg string) {
	if !p.verbose {
		return
	}
	if line != int(p.Line) {
		fmt.Printf("line %-4d%s%s\n", p.Line, p.indent, msg)
	} else {
		fmt.Printf("         %s%s\n", p.indent, msg)
	}
	line = int(p.Line)
}

// Testing Literal
func (p *parser) gotLiteral() error {
	print("Literal: " + p.Lit + "\n")
	p.Next()
	p.want(token.EOF)
	return nil
}

func (p *parser) want(tok token.Token) {
	if !p.got(tok) {
		p.syntaxError(fmt.Sprintf("expected %s, got %s", tok, p.Token))
	}
}

func (p *parser) got(tok token.Token) bool {
	if p.Token == tok {
		p.Next()
		return true
	}
	return false
}

func commentText(s string) string {
	if s[:2] == "/*" {
		return s[2 : len(s)-2] // lop off /* and */
	}

	// line comment (does not include newline)
	// (on Windows, the line comment may end in \r\n)
	i := len(s)
	if s[i-1] == '\r' {
		i--
	}
	return s[2:i] // lop off //, and \r at end, if any
}

func (p *parser) init(file *scanner.PosBase, r io.Reader, errh ErrHandler) {
	p.errh = errh
	p.file = file
	p.Scanner.Init(r,
		func(line, col uint, msg string) {
			if msg[0] != '/' {
				p.errorAt(p.posAt(line, col), msg)
				return
			}

			// otherwise it must be a comment containing a line or go: directive.
			// //line directives must be at the start of the line (column colbase).
			// /*line*/ directives can be anywhere in the line.
			text := commentText(msg)
			if (col == scanner.Colbase || msg[1] == '*') && strings.HasPrefix(text, "line ") {
				var pos scanner.Pos // position immediately following the comment
				if msg[1] == '/' {
					// line comment (newline is part of the comment)
					pos = scanner.MakePos(p.file, line+1, scanner.Colbase)
				} else {
					// regular comment
					// (if the comment spans multiple lines it's not
					// a valid line directive and will be discarded
					// by updateBase)
					pos = scanner.MakePos(p.file, line, col+uint(len(msg)))
				}
				p.updateBase(pos, line, col+2+5, text[5:]) // +2 to skip over // or /*
				return
			}

			//// go: directive (but be conservative and test)
			//if pragh != nil && strings.HasPrefix(text, "go:") {
			//	p.pragma = pragh(p.posAt(line, col+2), p.scanner.blank, text, p.pragma) // +2 to skip over // or /*
			//}
		},
		//func(line, col uint, msg string) {
		//	p.errorAt(p.posAt(line, col), msg)
		//
		//},
	)
	p.base = file
}

func tokstring(tok token.Token) string {
	switch tok {
	case token.Comma:
		return "comma"
	case token.Semi:
		return "semicolon or newline"
	}
	return tok.String()
}

// ----------------------------------------------------------------------------
// Error handling
func (p *parser) pos() scanner.Pos                 { return p.posAt(p.Line, p.Col) }
func (p *parser) posAt(line, col uint) scanner.Pos { return scanner.MakePos(p.base, line, col) }
func (p *parser) error(msg string)                 { p.errorAt(p.pos(), msg) }
func (p *parser) errorAt(pos scanner.Pos, msg string) {
	err := Error{pos, msg}
	if p.errh == nil {
		println(Yellow + err.Msg + Reset)
		return
	}
	p.errh(err)
}
func (p *parser) syntaxError(msg string) { p.syntaxErrorAt(p.pos(), msg) }

func (p *parser) syntaxErrorAt(pos scanner.Pos, msg string) {
	if p.verbose {
		p.print(Yellow + "syntax error: " + msg + Reset)
	}

	//if p.Token == token.EOF && p.first != nil {
	//	return // avoid meaningless follow-up errors
	//}

	// add punctuation etc. as needed to msg
	switch {
	case msg == "":
		// nothing to do
	case strings.HasPrefix(msg, "in "), strings.HasPrefix(msg, "at "), strings.HasPrefix(msg, "after "):
		msg = " " + msg
	case strings.HasPrefix(msg, "expecting "):
		msg = ", " + msg
	default:
		// plain error - we don't care about current Token
		p.errorAt(pos, "syntax error: "+msg)
		return
	}

	// determine Token string
	var tok string
	switch p.Token {
	case token.Name, token.Semi:
		tok = p.Lit
	case token.Literal:
		tok = "gotLiteral " + p.Lit
	case token.Op:
		tok = p.Op.String()
	case token.AssignOp:
		tok = p.Op.String() + "="
	case token.IncOp:
		tok = p.Op.String()
		tok += tok
	default:
		tok = tokstring(p.Token)
	}

	p.errorAt(pos, "syntax error: unexpected "+tok+msg)
}

const stopset uint64 = 1<<token.If |
	1<<token.Var

func (p *parser) gotAssign() bool {
	switch p.Token {
	case token.Define:
		p.error("expecting =")
		fallthrough
	case token.Assign:
		p.Next()
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Declarations
func (p *parser) appendGroup(list []ast.Decl, f func(group *ast.Group) ast.Decl) []ast.Decl {
	if x := f(nil); x != nil {
		list = append(list, x)
	}
	return list
}

// TypeSpec = identifier [ TypeParams ] [ "=" ] Type .
func (p *parser) typeDecl(group *ast.Group) ast.Decl {
	if p.verbose {
		defer p.trace("typeDecl")()
	}

	d := new(ast.TypeDecl)
	d.SetPos(p.pos())
	d.Group = group

	d.Name = p.name()
	d.Alias = p.gotAssign()
	d.Type = p.typeOrNil()

	if d.Type == nil {
		d.Type = p.badExpr()
		p.syntaxError("in type declaration")
	} else if p.verbose {
		p.print("id: " + d.Name.Value)
		p.print("type: " + d.Type.(*ast.Name).Value)
	}
	return d
}

// VarDecl = "var" identifier ( Type [ "=" ast.Expr ] | "=" ast.Expr ) .
func (p *parser) varDecl(group *ast.Group) ast.Decl {
	if p.verbose {
		defer p.trace("varDecl")()
	}

	d := new(ast.VarDecl)
	d.SetPos(p.pos())
	d.Group = group

	d.NameList = p.name()
	p.print("id: " + d.NameList.Value)
	if p.gotAssign() {
		d.Values = p.expr()
	} else {
		if p.Token != token.Name {
			p.syntaxError("expecting name")
			p.Next()
			return nil
		}

		d.Type = p.name()
		p.print("type: " + d.Type.(*ast.Name).Value)
	}

	return d
}

// TypeDecl =

// FuncDecl = "func" FuncName Signature FuncBody .
// FuncName = identifier .
func (p *parser) funcDeclOrNil(group *ast.Group) ast.Decl {
	if p.verbose {
		defer p.trace("funcDecl")()
	}

	// func name(name type) type {Body}
	d := new(ast.FuncDecl)
	d.SetPos(p.pos())
	d.Group = group

	if p.Token != token.Name {
		p.errorAt(p.pos(), "expecting name")
		return nil
	}

	//function name
	d.Name = p.name()
	p.print("id: " + d.Name.Value)

	// Signature
	d.Param, d.Return = p.funcType()

	// FuncBody
	if p.Token == token.Lbrace {
		d.Body = p.funcBody()
	}
	return d
}

// OperDecl = "oper" Receiver OperName OperOperand ReturnType OperBody .
// Receiver = "(" Param ")" .
// OperName =
//
//	"add" | "sub" | "mul" | "div" | "mod" |
//	"radd" | "rsub" | "rmul" | "rdiv" | "rmod" .
//
// OperOperand = "(" Param ")" .
// ReturnType = Type .
// OperBody = FuncBody .
func (p *parser) operDecl(group *ast.Group) ast.Decl {
	if p.verbose {
		defer p.trace("operDecl")()
	}

	d := new(ast.OperDecl)
	d.SetPos(p.pos())
	d.Group = group
	d.TypeL = p.singleParam()

	if !p.Token.IsOperator() {
		p.errorAt(p.pos(), "Unexpected Operator name")
		return nil
	}
	d.Oper = p.Token
	p.Next()
	p.print("oper type: " + d.Oper.String())
	d.TypeR = p.singleParam()
	p.print("operands: " + d.TypeL.Name.Value + " " + d.TypeR.Name.Value)
	if p.Token != token.Name {
		p.errorAt(p.pos(), "expecting type")
		return nil
	}
	d.Return = p.name()
	p.print("return type: " + d.Return.(*ast.Name).Value)
	d.Body = p.funcBody()

	return d
}

// FuncBody = Block .
func (p *parser) funcBody() *ast.BlockStmt {
	body := p.blockStmt("")
	return body
}

func (p *parser) funcType() ([]*ast.Field, ast.Expr) {
	params := make([]*ast.Field, 0)
	p.want(token.Lparen)
	params = p.paramlist()
	ftype := p.typeOrNil()
	switch ftype.(type) {
	case *ast.Name:
		p.print("return type: " + ftype.(*ast.Name).Value)
	case *ast.SliceType:
		slice := ftype.(*ast.SliceType)

		p.print("return type: slice of " + slice.Elem.(*ast.Name).Value)
	}
	return params, ftype
}

// ----------------------------------------------------------------------------
// Statements

// SimpleStmt = EmptyStmt | ast.ExpressionStmt | IncDecStmt | Assignment | ShortVarDecl .
func (p *parser) simpleStmt(ls ast.Expr, keyword token.Token) ast.SimpleStmt {
	if p.verbose {
		defer p.trace("simpleStmt")()
	}

	if ls == nil {
		ls = p.expr()
	}

	pos := p.pos()
	switch p.Token {
	case token.AssignOp, token.Assign:
		if p.verbose {
			defer p.trace("assignment")()
		}
		op := p.Op
		if p.Token == token.Assign {
			op = token.NoneOP
		}
		p.Next()
		return p.assignStmt(pos, op, ls, p.expr())
	case token.Define:
		if p.verbose {
			defer p.trace("shortVarDecl")()
		}
		p.Next()
		return p.defineStmt(pos, ls, p.expr())
	default:
		if p.verbose {
			defer p.trace("exprStmt")()
		}
		s := new(ast.ExprStmt)
		s.SetPos(ls.GetPos())
		s.X = ls
		return s
	}

}

func (p *parser) declStmt(f func(*ast.Group) ast.Decl) *ast.DeclStmt {
	if p.verbose {
		defer p.trace("declStmt")()
	}

	s := new(ast.DeclStmt)
	s.SetPos(p.pos())

	p.Next() // token.Const, token.Type, or token.Var
	s.DeclList = p.appendGroup(nil, f)

	return s
}

// Assignment = ast.Expr assigntoken.op ast.Expr .
// assigntoken.op = [ asstoken.op | multoken.op ] "=" .
func (p *parser) assignStmt(pos scanner.Pos, op token.Operator, lhs, rhs ast.Expr) *ast.AssignStmt {
	a := new(ast.AssignStmt)
	a.SetPos(pos)
	a.Op = op
	a.Lhs = lhs
	a.Rhs = rhs
	return a
}

func (p *parser) defineStmt(pos scanner.Pos, lhs, rhs ast.Expr) *ast.DefineStmt {
	s := new(ast.DefineStmt)
	s.SetPos(pos)
	s.Lhs = lhs
	s.Rhs = rhs
	return s
}

// Block = "{" StatementList "}" .
func (p *parser) blockStmt(context string) *ast.BlockStmt {
	if p.verbose {
		defer p.trace("blockStmt")()
	}
	s := new(ast.BlockStmt)
	s.SetPos(p.pos())
	// people coming from C may forget that braces are mandatory in Go
	if !p.got(token.Lbrace) {
		p.syntaxError("expecting '{'")
		return nil
	}
	s.StmtList = p.stmtList()

	s.Rbrace = p.pos()
	p.want(token.Rbrace)

	return s
}

// StatementList = { Statement ";" } .
func (p *parser) stmtList() (l []ast.Stmt) {
	if p.verbose {
		defer p.trace("stmtList")()
	}

	for p.Token != token.EOF && p.Token != token.Rbrace {
		s := p.stmtOrNil()
		if s == nil {
			break
		}
		l = append(l, s)
		// ";" is optional before "}"
		if !p.got(token.Semi) && p.Token != token.Rbrace {
			p.syntaxError("at end of statement")
			p.got(token.Semi) // avoid spurious empty statement
		}
	}
	return
}

// Statement =
//
//	Declaration | SimpleStmt | ReturnStmt | BreakStmt | ContinueStmt |
//	Block | IfStmt | ForStmt .
func (p *parser) stmtOrNil() ast.Stmt {
	if p.verbose {
		defer p.trace("stmt")()
	}

	if p.Token == token.Name {
		p.print("lhs:")
		lhs := p.expr()
		return p.simpleStmt(lhs, 0)
	}
	switch p.Token {
	case token.Var:
		return p.declStmt(p.varDecl)
	case token.Lbrace:
		return p.blockStmt("")
	case token.Literal, token.Name:
		return p.simpleStmt(nil, 0)
	case token.For:
		return p.forStmt()
	case token.While:
		p.Next()
		return p.whileStmt()
	case token.If:
		return p.ifStmt()
	case token.Return:
		s := new(ast.ReturnStmt)
		s.SetPos(p.pos())
		p.Next()
		if p.Token != token.Semi && p.Token != token.Rbrace {
			s.Return = p.expr()
		}
		return s
	case token.Break:
		s := new(ast.BreakStmt)
		s.SetPos(p.pos())
		p.Next()
		return s
	case token.Semi:
		func() { defer p.trace("empty stmt")() }()
		s := new(ast.EmptyStmt)
		s.SetPos(p.pos())
		return s
	}
	return nil
}

// ----------------------------------------------------------------------------
// ast.Expressions

func (p *parser) expr() ast.Expr {
	if p.verbose {
		defer p.trace("expr")()
	}

	return p.binaryExpr(0)
}

// ast.Expr = UnaryExpr | ast.Expr binarytoken.op ast.Expr .//a+b*x
func (p *parser) binaryExpr(prec int) ast.Expr {
	// don't p.verbose binaryExpr - only leads to overly nested p.verbose output

	x := p.unaryExpr()

	for (p.Token == token.Op) && p.Prec > prec {
		t := new(ast.Operation)
		t.SetPos(p.pos())
		t.Op = p.Op
		tprec := p.Prec
		p.print("operator(" + t.Op.String() + ")")
		p.Next()
		t.X = x
		t.Y = p.binaryExpr(tprec)

		switch t.Op {
		case token.Lss:
			t.Op = token.Gtr
			t.X, t.Y = t.Y, t.X
		}

		x = t
	}
	return x
}

// UnaryExpr = PrimaryExpr | unarytoken.op UnaryExpr .
func (p *parser) unaryExpr() ast.Expr {
	if p.verbose {
		defer p.trace("unaryExpr")()
	}
	switch p.Token {
	case token.Op:
		switch p.Op {
		case token.Mul, token.Add, token.Sub, token.Not: //, Xor:
			x := new(ast.Operation)
			x.SetPos(p.pos())
			x.Op = p.Op
			p.Next()
			x.X = p.unaryExpr()
			switch x.X.(type) {
			case *ast.BadExpr:
				p.error("bad expression")
			}
			return x

			//case And:
			//	x := new(Operation)
			//	x.SetPos(p.pos()
			//	x.Op = And
			//	p.Next()
			//	// unaryExpr may have returned a parenthesized composite gotLiteral
			//	// (see comment in operand) - remove parentheses if any
			//	x.X = unparen(p.unaryExpr())
			//	return x
		}
	}
	return p.pexpr()
}

func (p *parser) operand() (rtn ast.Expr) {
	if p.verbose {
		defer p.trace("operand")()
	}

	rtn = &ast.BadExpr{}
	tok := p.Token.String()
	switch p.Token {
	case token.Name:
		rtn = p.name()
		p.print(tok + "(" + rtn.(*ast.Name).Value + ")")
	case token.Lbrack:
		rtn = p.sliceLit()
		p.print(tok + "(" + ")")

	case token.Literal:
		lit := p.literal()
		rtn = lit
		p.print(tok + "(" + lit.Value + ")")
	}
	return
}

// PrimaryExpr =
//
//	Operand |
//	PrimaryExpr Selector |
//	PrimaryExpr Call .
//
// Selector       = "." identifier .
// Call			  = "(" [ ast.ExprList ] ")" .
func (p *parser) pexpr() ast.Expr {
	if p.verbose {
		defer p.trace("pexpr")()
	}
	x := p.operand()

loop:
	for {
		pos := p.pos()
		switch p.Token {
		case token.Dot:
			p.Next()
			switch p.Token {
			case token.Name:
				// pexpr '.' sym
				t := new(ast.SelectorExpr)
				t.SetPos(pos)
				t.X = x
				t.Sel = p.name()
				x = t

			default:
				p.syntaxError("expecting name or (")
			}
		case token.Lbrack:
			// pexpr '[' expr ']'
			t := new(ast.IndexExpr)
			t.SetPos(pos)
			t.X = x
			p.Next()
			t.Index = p.expr()
			p.want(token.Rbrack)
			x = t
		case token.Lparen:

			t := new(ast.CallExpr)
			t.SetPos(pos)
			t.Func = x
			t.ArgList = p.argList()
			x = t

		default:
			break loop
		}
	}

	return x
}

// ----------------------------------------------------------------------------
// Types
func (p *parser) typeOrNil() ast.Expr {
	switch p.Token {
	case token.Name:
		return p.name()
	case token.Lbrack:
		return p.sliceType()
	}
	return nil
}

func (p *parser) literal() *ast.BasicLit {
	if p.Token == token.Literal {
		b := new(ast.BasicLit)
		b.SetPos(p.pos())
		b.Value = p.Lit
		b.Kind = p.Kind
		b.Bad = p.Bad
		p.Next()
		return b
	}
	return nil
}

func (p *parser) singleParam() *ast.Field {
	param := new(ast.Field)
	if !p.got(token.Lparen) {
		p.syntaxError("expecting '('")
		return nil
	}
	first := true
recv:
	if p.Token != token.Name {
		str := "type"
		if first {
			str = "receiver"
		}
		p.syntaxError("expecting " + str)
		return nil
	}
	name := p.name()
	if first {
		param.Name = name
		first = false
		goto recv
	}
	param.Type = name
	p.want(token.Rparen)
	return param
}

func (p *parser) paramlist() []*ast.Field {
	list := make([]*ast.Field, 0)
	none := "none"
	str := " "
redo:
	param := new(ast.Field)
	switch p.Token {
	case token.Name:
		none = ""
		param.Name = p.name()
		if p.Token == token.Name {
			ptype := p.typeOrNil()
			str += none + param.Name.Value + "(" + ptype.(*ast.Name).Value + ") "
			param.Type = ptype
			list = append(list, param)
			switch p.Token {
			case token.Comma:
				p.Next()
				goto redo
			case token.Rparen:
				p.Next()
				p.print("params:" + str)
				return list
			default:
				p.syntaxError("expecting comma or ')'")
				p.Next()
				return nil
			}
		} else {
			p.syntaxError("expecting type")
			p.Next()
			return nil
		}
	case token.Rparen:
		p.Next()
		return nil
	default:
		p.syntaxError("expecting parameter or ')'")
		p.Next()
		return nil
	}
}

func (p *parser) argList() []ast.Expr {
	if p.verbose {
		defer p.trace("argList")()
	}
	list := make([]ast.Expr, 0)
	p.want(token.Lparen)
	if !p.got(token.Rparen) {
		list = append(list, p.expr())
		for !p.got(token.Rparen) {
			p.want(token.Comma)
			list = append(list, p.expr())
		}
	}

	return list
}

// ----------------------------------------------------------------------------
// Common
func (p *parser) name() *ast.Name {
	// no tracing to avoid overly p.verbose output

	if p.Token == token.Name {
		n := ast.NewName(p.pos(), p.Lit)
		p.Next()
		return n
	}

	n := ast.NewName(p.pos(), "token.")
	p.error("expecting name")
	return n
}

func (p *parser) nameList(first *ast.Name) []*ast.Name {
	if p.verbose {
		defer p.trace("nameList")()
	}

	l := []*ast.Name{first}
	for p.got(token.Comma) {
		l = append(l, p.name())
	}

	return l
}

func (p *parser) forStmt() ast.Stmt {
	if p.verbose {
		defer p.trace("forStmt")()
	}

	s := new(ast.ForStmt)
	s.SetPos(p.pos())

	s.Init, s.Cond, s.Post = p.header(token.For)
	s.Body = p.blockStmt("for clause")

	return s
}

func (p *parser) header(keyword token.Token) (init ast.SimpleStmt, cond ast.Expr, post ast.SimpleStmt) {
	p.want(keyword)
	if p.Token == token.Lbrace {
		if keyword == token.If {
			p.syntaxError("missing condition in if statement")
			cond = p.badExpr()
		}
		return
	}

	if p.Token != token.Semi {
		// accept potential varDecl but complain
		if p.got(token.Var) {
			p.syntaxError(fmt.Sprintf("var declaration not allowed in %s initializer", tokstring(keyword)))
		}
		init = p.simpleStmt(nil, keyword)
	}
	var condStmt ast.SimpleStmt
	var semi struct {
		pos scanner.Pos
		lit string // valid if pos.IsKnown()
	}
	if p.Token != token.Lbrace {
		if p.Token == token.Semi {
			semi.pos = p.pos()
			semi.lit = p.Lit
			p.Next()
		} else {
			// asking for a '{' rather than a ';' here leads to a better error message
			p.want(token.Lbrace)
		}
		if keyword == token.For {
			if p.Token != token.Semi {
				if p.Token == token.Lbrace {
					p.syntaxError("expecting for loop condition")
					goto done
				}
				condStmt = p.simpleStmt(nil, 0 /* range not permitted */)
			}
			p.want(token.Semi)
			if p.Token != token.Lbrace {
				post = p.simpleStmt(nil, 0 /* range not permitted */)
				if a, _ := post.(*ast.AssignStmt); a != nil && a.Op == token.Def {
					p.syntaxErrorAt(a.GetPos(), "cannot declare in post statement of for loop")
				}
			}
		} else if p.Token != token.Lbrace {
			condStmt = p.simpleStmt(nil, keyword)
		}
	} else {
		condStmt = init
		init = nil
	}
done:
	// unpack condStmt
	switch s := condStmt.(type) {
	case nil:
		if keyword == token.If && semi.pos.IsKnown() {
			if semi.lit != "semicolon" {
				p.syntaxErrorAt(semi.pos, fmt.Sprintf("unexpected %s, expecting { after if clause", semi.lit))
			} else {
				p.syntaxErrorAt(semi.pos, "missing condition in if statement")
			}
			b := new(ast.BadExpr)
			b.SetPos(semi.pos)
			cond = b
		}
	case *ast.ExprStmt:
		cond = s.X
	default:
		p.syntaxErrorAt(s.GetPos(), fmt.Sprintf("cannot use %s as value", s))
	}
	return
}

func (p *parser) badExpr() *ast.BadExpr {
	b := new(ast.BadExpr)
	b.SetPos(p.pos())
	return b
}

func (p *parser) ifStmt() *ast.IfStmt {
	if p.verbose {
		defer p.trace("ifStmt")()
	}
	s := new(ast.IfStmt)
	s.SetPos(p.pos())
	_, s.Cond, _ = p.header(token.If)
	s.Block = p.blockStmt("If clause")
	if p.got(token.Else) {
		switch p.Token {
		case token.If:
			s.Else = p.ifStmt()
		case token.Lbrace:
			s.Else = p.blockStmt("")
		default:
			p.syntaxError("else must be followed by if or statement block")
		}
	}
	return s
}

func (p *parser) whileStmt() ast.Stmt {
	if p.verbose {
		defer p.trace("whileStmt")()
	}
	s := new(ast.WhileStmt)
	s.SetPos(p.pos())
	s.Cond = p.expr()
	s.Body = p.blockStmt("While clause")
	return s
}

func (p *parser) sliceType() ast.Expr {
	if p.verbose {
		defer p.trace("sliceType")()
	}
	t := new(ast.SliceType)
	t.SetPos(p.pos())
	p.Next()
	p.want(token.Rbrack)
	t.Elem = p.typeOrNil()
	if t.Elem == nil {
		//elem = p.badExpr()
		p.syntaxError("invalid element type in slice")
	}
	//p.want(token.Rbrack)
	return t
}

func (p *parser) sliceLit() ast.Expr {
	if p.verbose {
		defer p.trace("sliceLit")()
	}
	l := new(ast.SliceLit)
	l.SetPos(p.pos())
	p.Next()
	p.want(token.Rbrack)
	l.ElemType = p.typeOrNil()
	if l.ElemType == nil {
		//elem = p.badExpr()
		p.syntaxError("invalid element type in slice")
	}
	p.want(token.Lbrace)
	l.Elems = make([]ast.Expr, 0)
	if !p.got(token.Rbrace) {
		l.Elems = append(l.Elems, p.expr())
		for !p.got(token.Rbrace) {
			p.want(token.Comma)
			l.Elems = append(l.Elems, p.expr())
		}
	}
	return l
}

func (p *parser) updateBase(pos scanner.Pos, tline, tcol uint, text string) {
	i, n, ok := trailingDigits(text)
	if i == 0 {
		return // ignore (not a line directive)
	}
	// i > 0

	if !ok {
		// text has a suffix :xxx but xxx is not a number
		p.errorAt(p.posAt(tline, tcol+i), "invalid line number: "+text[i:])
		return
	}

	var line, col uint
	i2, n2, ok2 := trailingDigits(text[:i-1])
	if ok2 {
		//line filename:line:col
		i, i2 = i2, i
		line, col = n2, n
		if col == 0 || col > scanner.PosMax {
			p.errorAt(p.posAt(tline, tcol+i2), "invalid column number: "+text[i2:])
			return
		}
		text = text[:i2-1] // lop off ":col"
	} else {
		//line filename:line
		line = n
	}

	if line == 0 || line > scanner.PosMax {
		p.errorAt(p.posAt(tline, tcol+i), "invalid line number: "+text[i:])
		return
	}

	// If we have a column (//line filename:line:col form),
	// an empty filename means to use the previous filename.
	filename := text[:i-1] // lop off ":line"
	if filename == "" && ok2 {
		filename = p.base.Filename()
	}

	p.base = NewLineBase(pos, filename, line, col)
}

func NewLineBase(pos scanner.Pos, filename string, line, col uint) *scanner.PosBase {
	return scanner.NewPosBase(pos, filename, line, col)
}

func trailingDigits(text string) (uint, uint, bool) {
	// Want to use LastIndexByte below but it's not defined in Go1.4 and bootstrap fails.
	i := strings.LastIndex(text, ":") // look from right (Windows filenames may contain ':')
	if i < 0 {
		return 0, 0, false // no ":"
	}
	// i >= 0
	n, err := strconv.ParseUint(text[i+1:], 10, 0)
	return uint(i + 1), uint(n), err == nil
}

func unparen(x ast.Expr) ast.Expr {
	for {
		p, ok := x.(*ast.ParenExpr)
		if !ok {
			break
		}
		x = p.X
	}
	return x
}
