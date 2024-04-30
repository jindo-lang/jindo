package parser

import (
	"bytes"
	"errors"
	"io"
	"jindo/pkg/jindo/ast"
	"jindo/pkg/jindo/scanner"
	"jindo/pkg/jindo/token"
	"os"
	"strings"
	"testing"
)

// If src != nil, readSource converts src to a []byte if possible;
// otherwise it returns an error. If src == nil, readSource returns
// the result of reading the file specified by filename.
func readSource(filename string, src any) (io.Reader, error) {
	if src != nil {
		switch s := src.(type) {
		case string:
			return strings.NewReader(s), nil
		case []byte:
			return bytes.NewReader(s), nil
		case *bytes.Buffer:
			// is io.Reader, but src is already available in []byte form
			if s != nil {
				return s, nil
			}
		case io.Reader:
			return s, nil
		}
		return nil, errors.New("invalid source")
	}
	f, ferr := os.Open(filename)
	if ferr != nil {
		println(ferr.Error())
		os.Exit(-1)
	}
	return f, nil
}

var errReport error

func pickError() (e error) {
	e = errReport
	errReport = nil
	return
}

var test_errh = func(err error) { errReport = err }

// ParseExprFrom is a convenience function for parsing an expression.
// The arguments have the same meaning as for ParseFile, but the source must
// be a valid Go (type or value) expression. Specifically, fset must not
// be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by source position.
func ParseExprFrom(fset *scanner.PosBase, filename string, src any) (expr ast.Expr, err error) {
	if fset == nil {
		panic("parser.ParseExprFrom: no token.FileSet provided (fset == nil)")
	}

	// get source
	reader, err := readSource(filename, src)
	if err != nil {
		return nil, err
	}

	var p parser
	// parse expr
	p.init(fset, reader, test_errh)
	p.Next()
	expr = p.expr()

	// If a semicolon was inserted, consume it;
	// report an error if there's more tokens.
	if p.Token == token.Semi && p.Lit == "\n" {
		p.Next()
	}
	p.want(token.EOF)

	err = pickError()
	return
}

// ParseExpr is a convenience function for obtaining the AST of an expression x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
//
// If syntax errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors are
// returned via a scanner.ErrorList which is sorted by source position.
func ParseExpr(x string) (ast.Expr, error) {
	return ParseExprFrom(scanner.NewFileBase(""), "", x)
}

func TestParseExpr(t *testing.T) {
	// just kicking the tires:
	// a valid arithmetic expression
	src := "a + b"
	x, err := ParseExpr(src)
	if err != nil {
		t.Errorf("ParseExpr(%q): %v", src, err)
	}
	// sanity check
	if _, ok := x.(ast.BinaryExpr); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BinaryExpr", src, x)
	}

	// a valid type expression
	//src = "struct{x *int}"
	//x, err = ParseExpr(src)
	//if err != nil {
	//	t.Errorf("ParseExpr(%q): %v", src, err)
	//}
	//// sanity check
	//if _, ok := x.(ast.StructType); !ok {
	//	t.Errorf("ParseExpr(%q): got %T, want *ast.StructType", src, x)
	//}

	//an invalid expression
	src = "a + *"
	x, err = ParseExpr(src)
	if err == nil {
		t.Errorf("ParseExpr(%q): got no error", src)
	}
	if x == nil {
		t.Errorf("ParseExpr(%q): got no (partial) result", src)
	}
	if _, ok := x.(ast.BinaryExpr); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BinaryExpr", src, x)
	}

	// a valid expression followed by extra tokens is invalid
	//src = "a[i] := x"
	//if _, err := ParseExpr(src); err == nil {
	//	t.Errorf("ParseExpr(%q): got no error", src)
	//}

	// a semicolon is not permitted unless automatically inserted
	src = "a + b\n"
	if _, err := ParseExpr(src); err != nil {
		t.Errorf("ParseExpr(%q): got error %s", src, err)
	}
	src = "a + b;"
	if _, err := ParseExpr(src); err == nil {
		t.Errorf("ParseExpr(%q): got no error", src)
	}

	// various other stuff following a valid expression
	const validExpr = "a + b"
	const anything = "dh3*#D)#_"
	for _, c := range "!)]};," {
		src := validExpr + string(c) + anything
		if _, err := ParseExpr(src); err == nil {
			t.Errorf("ParseExpr(%q): got no error", src)
		}
	}

	//// ParseExpr must not crash
	//for _, src := range valids {
	//	ParseExpr(src)
	//}
}
