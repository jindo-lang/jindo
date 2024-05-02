// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package parser

import (
	"bytes"
	"fmt"
	"io"
	"jindo/pkg/jindo/ast"
	"jindo/pkg/jindo/position"
	"os"
	"testing"
)

const (
	src_ = "/Users/seungyeoplee/Workspace/jindo/example/test.paw"
)

func testOut() io.Writer {
	if testing.Verbose() {
		return os.Stdout
	}
	return io.Discard
}

func TestDump(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	parsed, _ := ParseFile(src_, func(err error) { t.Error(err) })

	if parsed != nil {
		ast.Fdump(testOut(), parsed)
	}
}

func TestParse(t *testing.T) {
	ParseFile(src_, func(err error) { t.Error(err) })
}

func TestVerify(t *testing.T) {
	ast, err := ParseFile(src_, func(err error) { t.Error(err) })
	if err != nil {
		return // error already reported
	}
	verifyPrint(t, src_, ast)
}

func verifyPrint(t *testing.T, filename string, ast1 *ast.File) {
	var buf1 bytes.Buffer
	_, err := Fprint(&buf1, ast1, LineForm)
	if err != nil {
		panic(err)
	}
	bytes1 := buf1.Bytes()

	ast2, err := Parse(position.NewFileBase(filename), &buf1, nil)
	if err != nil {
		panic(err)
	}

	var buf2 bytes.Buffer
	_, err = Fprint(&buf2, ast2, LineForm)
	if err != nil {
		panic(err)
	}
	bytes2 := buf2.Bytes()

	if !bytes.Equal(bytes1, bytes2) {
		fmt.Printf("--- %s ---\n", filename)
		fmt.Printf("%s\n", bytes1)
		fmt.Println()

		fmt.Printf("--- %s ---\n", filename)
		fmt.Printf("%s\n", bytes2)
		fmt.Println()

		t.Error("printed syntax trees do not match")
	}
}
