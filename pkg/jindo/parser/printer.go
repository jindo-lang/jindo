// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements printing of syntax trees in source format.

package parser

import (
	"fmt"
	"io"
	"jindo/pkg/jindo/ast"
	"jindo/pkg/jindo/token"
	"strings"
)

// Form controls print formatting.
type Form uint

const (
	_         Form = iota // default
	LineForm              // use spaces instead of linebreaks where possible
	ShortForm             // like LineForm but print "…" for non-empty function or composite literal bodies
)

// Fprint prints node x to w in the specified form.
// It returns the number of bytes written, and whether there was an error.
func Fprint(w io.Writer, x ast.Node, form Form) (n int, err error) {
	p := printer{
		output:     w,
		form:       form,
		linebreaks: form == 0,
	}

	defer func() {
		n = p.written
		if e := recover(); e != nil {
			err = e.(ast.WriteError).Err // re-panics if it's not a writeError
		}
	}()

	p.print(x)
	p.flush(token.EOF)

	return
}

// String is a convenience function that prints n in ShortForm
// and returns the printed string.
func String(n ast.Node) string {
	var buf strings.Builder
	_, err := Fprint(&buf, n, ShortForm)
	if err != nil {
		fmt.Fprintf(&buf, "<<< ERROR: %s", err)
	}
	return buf.String()
}

type ctrlSymbol int

const (
	none ctrlSymbol = iota
	semi
	blank
	newline
	indent
	outdent
	// comment
	// eolComment
)

type whitespace struct {
	last token.Token
	kind ctrlSymbol
	//text string // comment text (possibly ""); valid if kind == comment
}

type printer struct {
	output     io.Writer
	written    int // number of bytes written
	form       Form
	linebreaks bool // print linebreaks instead of semis

	indent  int // current indentation level
	nlcount int // number of consecutive newlines

	pending []whitespace // pending whitespace
	lastTok token.Token  // last token.Token (after any pending semi) processed by print
}

// write is a thin wrapper around p.output.Write
// that takes care of accounting and error handling.
func (p *printer) write(data []byte) {
	n, err := p.output.Write(data)
	p.written += n
	if err != nil {
		panic(ast.NewWriteError(err))
	}
}

var (
	tabBytes    = []byte("\t\t\t\t\t\t\t\t")
	newlineByte = []byte("\n")
	blankByte   = []byte(" ")
)

func (p *printer) writeBytes(data []byte) {
	if len(data) == 0 {
		panic("expected non-empty []byte")
	}
	if p.nlcount > 0 && p.indent > 0 {
		// write indentation
		n := p.indent
		for n > len(tabBytes) {
			p.write(tabBytes)
			n -= len(tabBytes)
		}
		p.write(tabBytes[:n])
	}
	p.write(data)
	p.nlcount = 0
}

func (p *printer) writeString(s string) {
	p.writeBytes([]byte(s))
}

// If impliesSemi returns true for a non-blank line's final token.Token tok,
// a semicolon is automatically inserted. Vice versa, a semicolon may
// be omitted in those cases.
func impliesSemi(tok token.Token) bool {
	switch tok {
	case token.Name,
		token.Break, token.Continue, token.Return,
		/*_Inc, _Dec,*/ token.Rparen, token.Rbrack, token.Rbrace: // TODO(gri) fix this
		return true
	}
	return false
}

// TODO(gri) provide table of []byte values for all token.Tokens to avoid repeated string conversion

func lineComment(text string) bool {
	return strings.HasPrefix(text, "//")
}

func (p *printer) addWhitespace(kind ctrlSymbol, text string) {
	p.pending = append(p.pending, whitespace{p.lastTok, kind /*text*/})
	switch kind {
	case semi:
		p.lastTok = token.Semi
	case newline:
		p.lastTok = 0
		// TODO(gri) do we need to handle /*-style comments containing newlines here?
	}
}

func (p *printer) flush(next token.Token) {
	// eliminate semis and redundant whitespace
	sawNewline := next == token.EOF
	sawParen := next == token.Rparen || next == token.Rbrace
	for i := len(p.pending) - 1; i >= 0; i-- {
		switch p.pending[i].kind {
		case semi:
			k := semi
			if sawParen {
				sawParen = false
				k = none // eliminate semi
			} else if sawNewline && impliesSemi(p.pending[i].last) {
				sawNewline = false
				k = none // eliminate semi
			}
			p.pending[i].kind = k
		case newline:
			sawNewline = true
		case blank, indent, outdent:
			// nothing to do
		// case comment:
		// 	// A multi-line comment acts like a newline; and a ""
		// 	// comment implies by definition at least one newline.
		// 	if text := p.pending[i].text; strings.HasPrefix(text, "/*") && strings.ContainsRune(text, '\n') {
		// 		sawNewline = true
		// 	}
		// case eolComment:
		// 	// TODO(gri) act depending on sawNewline
		default:
			panic("unreachable")
		}
	}

	// print pending
	prev := none
	for i := range p.pending {
		switch p.pending[i].kind {
		case none:
			// nothing to do
		case semi:
			p.writeString(";")
			p.nlcount = 0
			prev = semi
		case blank:
			if prev != blank {
				// at most one blank
				p.writeBytes(blankByte)
				p.nlcount = 0
				prev = blank
			}
		case newline:
			const maxEmptyLines = 1
			if p.nlcount <= maxEmptyLines {
				p.write(newlineByte)
				p.nlcount++
				prev = newline
			}
		case indent:
			p.indent++
		case outdent:
			p.indent--
			if p.indent < 0 {
				panic("negative indentation")
			}
		// case comment:
		// 	if text := p.pending[i].text; text != "" {
		// 		p.writeString(text)
		// 		p.nlcount = 0
		// 		prev = comment
		// 	}
		// 	// TODO(gri) should check that line comments are always followed by newline
		default:
			panic("unreachable")
		}
	}

	p.pending = p.pending[:0] // re-use underlying array
}

func mayCombine(prev token.Token, next byte) (b bool) {
	return // for now
	// switch prev {
	// case lexical.Int:
	// 	b = next == '.' // 1.
	// case lexical.Add:
	// 	b = next == '+' // ++
	// case lexical.Sub:
	// 	b = next == '-' // --
	// case lexical.Quo:
	// 	b = next == '*' // /*
	// case lexical.Lss:
	// 	b = next == '-' || next == '<' // <- or <<
	// case lexical.And:
	// 	b = next == '&' || next == '^' // && or &^
	// }
	// return
}

func (p *printer) print(args ...interface{}) {
	for i := 0; i < len(args); i++ {
		switch x := args[i].(type) {
		case nil:
			// we should not reach here but don't crash

		case ast.Node:
			p.printNode(x)

		case token.Token:
			// token.Token.Name implies an immediately following string
			// argument which is the actual value to print.
			var s string
			if x == token.Name {
				i++
				if i >= len(args) {
					panic("missing string argument after token.Token.Name")
				}
				s = args[i].(string)
			} else {
				s = x.String()
			}

			// TODO(gri) This check seems at the wrong place since it doesn't
			//           take into account pending white space.
			if mayCombine(p.lastTok, s[0]) {
				panic("adjacent token.Tokens combine without whitespace")
			}

			if x == token.Semi {
				// delay printing of semi
				p.addWhitespace(semi, "")
			} else {
				p.flush(x)
				p.writeString(s)
				p.nlcount = 0
				p.lastTok = x
			}

		case token.Operator:
			if x != 0 {
				p.flush(token.Op)
				p.writeString(x.String())
			}

		case ctrlSymbol:
			switch x {
			case none, semi /*, comment*/ :
				panic("unreachable")
			case newline:
				// TODO(gri) need to handle mandatory newlines after a //-style comment
				if !p.linebreaks {
					x = blank
				}
			}
			p.addWhitespace(x, "")

		// case *Comment: // comments are not ast.Nodes
		// 	p.addWhitespace(comment, x.Text)

		default:
			panic(fmt.Sprintf("unexpected argument %v (%T)", x, x))
		}
	}
}

func (p *printer) printNode(n ast.Node) {
	// ncom := *n.Comments()
	// if ncom != nil {
	// 	// TODO(gri) in general we cannot make assumptions about whether
	// 	// a comment is a /*- or a //-style comment since the syntax
	// 	// tree may have been manipulated. Need to make sure the correct
	// 	// whitespace is emitted.
	// 	for _, c := range ncom.Alone {
	// 		p.print(c, newline)
	// 	}
	// 	for _, c := range ncom.Before {
	// 		if c.Text == "" || lineComment(c.Text) {
	// 			panic("unexpected empty line or //-style 'before' comment")
	// 		}
	// 		p.print(c, blank)
	// 	}
	// }

	p.printRawNode(n)

	// if ncom != nil && len(ncom.After) > 0 {
	// 	for i, c := range ncom.After {
	// 		if i+1 < len(ncom.After) {
	// 			if c.Text == "" || lineComment(c.Text) {
	// 				panic("unexpected empty line or //-style non-final 'after' comment")
	// 			}
	// 		}
	// 		p.print(blank, c)
	// 	}
	// 	//p.print(newline)
	// }
}

func (p *printer) printRawNode(n ast.Node) {
	switch n := n.(type) {
	case nil:
		// we should not reach here but don't crash

	// expressions and types
	case *ast.BadExpr:
		p.print(token.Name, "<bad expr>")

	case *ast.Name:
		p.print(token.Name, n.Value) // token.Token.Name requires actual value following immediately

	case *ast.BasicLit:
		p.print(token.Name, n.Value) // token.Token.Name requires actual value following immediately

	case *ast.ParenExpr:
		p.print(token.Lparen, n.X, token.Rparen)

	case *ast.SelectorExpr:
		p.print(n.X, token.Dot, n.Sel)

	case *ast.IndexExpr:
		p.print(n.X, token.Lbrack, n.Index, token.Rbrack)

	case *ast.CallExpr:
		p.print(n.Func, token.Lparen)
		p.printExprList(n.ArgList)
		p.print(token.Rparen)

	case *ast.Operation:
		if n.Y == nil {
			// unary expr
			p.print(n.Op)
			// if n.Op == lexical.Range {
			// 	p.print(blank)
			// }
			p.print(n.X)
		} else {
			// binary expr
			// TODO(gri) eventually take precedence into account
			// to control possibly missing parentheses
			p.print(n.X, blank, n.Op, blank, n.Y)
		}

	case *ast.SliceType:
		p.print(token.Lbrack, token.Rbrack, n.Elem)

	// statements
	case *ast.DeclStmt:
		p.printDecl(n.DeclList)

	case *ast.EmptyStmt:
		// nothing to print

	case *ast.ExprStmt:
		p.print(n.X)

	case *ast.AssignStmt:
		p.print(n.Lhs)
		if n.Rhs == nil {
			// TODO(gri) This is going to break the mayCombine
			//           check once we enable that again.
			p.print(n.Op, n.Op) // ++ or --
		} else {
			p.print(blank, n.Op, token.Assign, blank)
			p.print(n.Rhs)
		}

	case *ast.ReturnStmt:
		p.print(token.Return)
		if n.Result != nil {
			p.print(blank, n.Result)
		}

	case *ast.BlockStmt:
		p.print(token.Lbrace)
		if len(n.StmtList) > 0 {
			p.print(newline, indent)
			p.printStmtList(n.StmtList, true)
			p.print(outdent, newline)
		}
		p.print(token.Rbrace)

	case *ast.IfStmt:
		p.print(token.If, blank)
		p.print(n.Cond, blank, n.Block)
		if n.Else != nil {
			p.print(blank, token.Else, blank, n.Else)
		}

	case *ast.ForStmt:
		p.print(token.For, blank)
		if n.Init == nil && n.Post == nil {
			if n.Cond != nil {
				p.print(n.Cond, blank)
			}
		} else {
			if n.Init != nil {
				p.print(n.Init)
			}
			p.print(token.Semi, blank)
			if n.Cond != nil {
				p.print(n.Cond)
			}
			p.print(token.Semi, blank)
			if n.Post != nil {
				p.print(n.Post, blank)
			}
		}
		p.print(n.Body)

	case *ast.ImportDecl:
		if n.Group == nil {
			p.print(token.Import, blank)
		}
		p.print(n.Path)

	case *ast.TypeDecl:
		if n.Group == nil {
			p.print(token.Type, blank)
		}
		p.print(n.Name)
		p.print(blank)
		if n.Alias {
			p.print(token.Assign, blank)
		}
		p.print(n.Type)

	case *ast.VarDecl:
		if n.Group == nil {
			p.print(token.Var, blank)
		}
		p.printNameList([]*ast.Name{n.NameList})
		if n.Type != nil {
			p.print(blank, n.Type)
		}
		if n.Values != nil {
			p.print(blank, token.Assign, blank, n.Values)
		}

	case *ast.FuncDecl:
		p.print(token.Func, blank)

		// receiver not implemented
		//if r := n.Recv; r != nil {
		//	p.print(token.Lparen)
		//	if r.Name != nil {
		//		p.print(r.Name, blank)
		//	}
		//	p.printNode(r.Type)
		//	p.print(token.Rparen, blank)
		//}
		p.print(n.Name)
		p.printSignature(n)
		if n.Body != nil {
			p.print(blank, n.Body)
		}

	case *printGroup:
		p.print(n.Tok, blank, token.Lparen)
		if len(n.Decls) > 0 {
			p.print(newline, indent)
			for _, d := range n.Decls {
				p.printNode(d)
				p.print(token.Semi, newline)
			}
			p.print(outdent)
		}
		p.print(token.Rparen)

	// files
	case *ast.File:
		p.print(token.Space, blank, n.SpaceName)
		if len(n.DeclList) > 0 {
			p.print(token.Semi, newline, newline)
			p.printDeclList(n.DeclList)
		}

	default:
		panic(fmt.Sprintf("syntax.Iterate: unexpected node type %T", n))
	}
}

func (p *printer) printNameList(list []*ast.Name) {
	for i, x := range list {
		if i > 0 {
			p.print(token.Comma, blank)
		}
		p.printNode(x)
	}
}

func (p *printer) printExprList(list []ast.Expr) {
	for i, x := range list {
		if i > 0 {
			p.print(token.Comma, blank)
		}
		p.printNode(x)
	}
}

func (p *printer) printExprLines(list []ast.Expr) {
	if len(list) > 0 {
		p.print(newline, indent)
		for _, x := range list {
			p.print(x, token.Comma, newline)
		}
		p.print(outdent)
	}
}

func groupFor(d ast.Decl) (token.Token, *ast.Group) {
	switch d := d.(type) {
	case *ast.ImportDecl:
		return token.Import, d.Group
	case *ast.TypeDecl:
		return token.Type, d.Group
	case *ast.VarDecl:
		return token.Var, d.Group
	case *ast.FuncDecl:
		return token.Func, nil
	default:
		panic("unreachable")
	}
}

type printGroup struct {
	ast.BadExpr
	Tok   token.Token
	Decls []ast.Decl
}

func (p *printer) printDecl(list []ast.Decl) {
	tok, group := groupFor(list[0])

	if group == nil {
		if len(list) != 1 {
			panic("unreachable")
		}
		p.printNode(list[0])
		return
	}

	// if _, ok := list[0].(*EmptyDecl); ok {
	// 	if len(list) != 1 {
	// 		panic("unreachable")
	// 	}
	// 	// TODO(gri) if there are comments inside the empty
	// 	// group, we may need to keep the list non-nil
	// 	list = nil
	// }

	// printGroup is here for consistent comment handling
	// (this is not yet used)
	var pg printGroup
	// *pg.Comments() = *group.Comments()
	pg.Tok = tok
	pg.Decls = list
	p.printNode(&pg)
}

func (p *printer) printDeclList(list []ast.Decl) {
	i0 := 0
	var tok token.Token
	var group *ast.Group
	for i, x := range list {
		if s, g := groupFor(x); g == nil || g != group {
			if i0 < i {
				p.printDecl(list[i0:i])
				p.print(token.Semi, newline)
				// print empty line between different declaration groups,
				// different kinds of declarations, or between functions
				if g != group || s != tok || s == token.Func {
					p.print(newline)
				}
				i0 = i
			}
			tok, group = s, g
		}
	}
	p.printDecl(list[i0:])
}

func (p *printer) printSignature(fn *ast.FuncDecl) {
	p.printParameterList(fn.Param, 0)
	p.printNode(fn.Return)
}

// If tok != 0 print a type parameter list: tok == token.Type means
// a type parameter list for a type, tok == _Func means a type
// parameter list for a func.
func (p *printer) printParameterList(list []*ast.Field, tok token.Token) {
	open, close := token.Lparen, token.Rparen

	//if tok != 0 {
	//	open, close = token.Lbrack, token.Rbrack
	//}
	// no generic support

	p.print(open)
	for i, f := range list {
		if i > 0 {
			p.print(token.Comma, blank)
		}
		if f.Name != nil {
			p.printNode(f.Name)
			if i+1 < len(list) {
				f1 := list[i+1]
				if f1.Name != nil && f1.Type == f.Type {
					continue // no need to print type
				}
			}
			p.print(blank)
		}
		p.printNode(Unparen(f.Type)) // no need for (extra) parentheses around parameter types
	}
	// A type parameter list [P T] where the name P and the type expression T syntactically
	// combine to another valid (value) expression requires a trailing comma, as in [P *T,]
	// (or an enclosing interface as in [P interface(*T)]), so that the type parameter list
	// is not parsed as an array length [P*T].
	if tok == token.Type && len(list) == 1 && combinesWithName(list[0].Type) {
		p.print(token.Comma)
	}
	p.print(close)
}

// combinesWithName reports whether a name followed by the expression x
// syntactically combines to another valid (value) expression. For instance
// using *T for x, "name *T" syntactically appears as the expression x*T.
// On the other hand, using  P|Q or *P|~Q for x, "name P|Q" or name *P|~Q"
// cannot be combined into a valid (value) expression.
func combinesWithName(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.Operation:
		if x.Y == nil {
			// name *x.X combines to name*x.X if x.X is not a type element
			return x.Op == token.Mul && !isTypeElem(x.X)
		}
		// binary expressions
		return combinesWithName(x.X) && !isTypeElem(x.Y)
	case *ast.ParenExpr:
		// name(x) combines but we are making sure at
		// the call site that x is never parenthesized.
		panic("unexpected parenthesized expression")
	}
	return false
}

func (p *printer) printStmtList(list []ast.Stmt, braces bool) {
	for i, x := range list {
		p.print(x, token.Semi)
		if i+1 < len(list) {
			p.print(newline)
		} else if braces {
			// Print an extra semicolon if the last statement is
			// an empty statement and we are in a braced block
			// because one semicolon is automatically removed.
			if _, ok := x.(*ast.EmptyStmt); ok {
				p.print(x, token.Semi)
			}
		}
	}
}

func (p *printer) printFields(fields []*ast.Field, tags []*ast.BasicLit, i, j int) {
	if i+1 == j && fields[i].Name == nil {
		// anonymous field
		p.printNode(fields[i].Type)
	} else {
		for k, f := range fields[i:j] {
			if k > 0 {
				p.print(token.Comma, blank)
			}
			p.printNode(f.Name)
		}
		p.print(blank)
		p.printNode(fields[i].Type)
	}
	if i < len(tags) && tags[i] != nil {
		p.print(blank)
		p.printNode(tags[i])
	}
}

func (p *printer) printFieldList(fields []*ast.Field, tags []*ast.BasicLit, sep token.Token) {
	i0 := 0
	var typ ast.Expr
	for i, f := range fields {
		if f.Name == nil || f.Type != typ {
			if i0 < i {
				p.printFields(fields, tags, i0, i)
				p.print(sep, newline)
				i0 = i
			}
			typ = f.Type
		}
	}
	p.printFields(fields, tags, i0, len(fields))
}
