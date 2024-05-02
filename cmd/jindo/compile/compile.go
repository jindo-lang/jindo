// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package compile

import (
	"context"
	"errors"
	"fmt"
	"jindo-tool/command"
	"jindo/pkg/jindo/ast"
	"jindo/pkg/jindo/parser"
	"os"
	"path/filepath"
)

var CmdCompile = &command.Command{
	UsageLine: "jindo compile [-o output] [build flags] [file or directory]",
	Short:     "compile single space from input",
	Long: `
Compile compiles the single space specified by the input,
handling the files as independent compilation units without compiling or linking dependencies.

The default output is a bytecode object. However, if the specified output name ends with '.ir',
the command outputs an intermediate representation (IR) file instead.

If the arguments to compile are a list of .paw files from a single directory,
compile treats them as a list of source files specifying a single space.

When the input is a directory, CompileCmd processes all .paw files within the directory
as a list of source files specifying a single space.

Compile does not include source files sharing the same space unless
they are specified in the arguments, or the input is a directory.

All source files compiled together should share the same space name and reside in the same directory.

When compiling a space from a list of .paw files, the object file is named after the first source file.
For example, 'jindo compile ed.paw rx.paw' writes 'ed.obj'. If the output file name is specified as ending with '.ir',
such as 'jindo compile -o output.ir ed.paw rx.paw', it writes 'output.ir'.

The -o flag allows specifying an output file or directory for the resulting IR file or object,
overriding the default naming convention. If the named output is an existing directory or
ends with a slash or backslash, the resulting outputs are written to that directory.
`,
}

var (
	FlagO string
)

func init() {
	CmdCompile.Run = runCompile
	CmdCompile.Flag.StringVar(&FlagO, "o", "", "output file or directory")
}

func runCompile(ctx context.Context, cmd *command.Command, args []string) {
	name, format, err := validateOutputName(FlagO)
	if err != nil {
		panic(err)
	}

	fmt.Printf("source(s): %v\noutput name: %v\nformat: %v\n", args, name, format)
	space, err := loadSpace(ctx, args)
	if err != nil {
		panic(err)
	}
	for _, f := range space.FileSet {
		e := ast.Fdump(os.Stdout, f)
		if e != nil {
			panic(e)
		}
		fmt.Println()
	}
	//args == sources
	// load(sources) => space
	// 		space{ spaceName; files; imports }
	// load(sources):
	//		check(file extensions) => ext
	//			if any(args...ext).not(".paw") => abort
	//
	//		check(file directions) => dir
	//			if any(args...dir).different() => abort
	//
	// 		check(space names)
	//			if any(args...pkgName).different() => abort
	//		check(imports)
	//	return space{spaceName, files}

}

type Compiler struct {
	WorkDir string
}

func NewCompiler() *Compiler {
	return nil
}

type Space struct {
	Name    string
	FileSet []*ast.File
}

func loadSpace(ctx context.Context, sources []string) (s *Space, e error) {
	if len(sources) == 0 {
		return nil, errors.New("no source files provided")
	}

	s = new(Space)

	// Check for file extensions and directory uniformity
	var dir string
	space := ""
	for _, file := range sources {
		if filepath.Ext(file) != ".paw" {
			return nil, fmt.Errorf("invalid file extension for %s, expected .paw", file)
		}

		currentDir := filepath.Dir(file)
		if dir != "" && currentDir != dir {
			return nil, fmt.Errorf("files must be in the same directory: %s is not in %s", file, dir)
		}
		dir = currentDir

		parsed, err := parser.ParseFile(file, nil)
		if err != nil {
			return nil, err
		}
		curSpace := parsed.SpaceName.Value
		if space != "" && curSpace != space {
			return nil, fmt.Errorf("space name mismatch: %s does not match %s", curSpace, space)
		}
		space = curSpace
		s.FileSet = append(s.FileSet, parsed)
	}
	s.Name = space

	return s, nil
}

func validateOutputName(outputName string) (name string, format string, err error) {
	// Default to output name and "obj" format
	name, format = outputName, "obj"
	if name == "" {
		return
	}

	ext := filepath.Ext(name)
	// Use a switch statement to check the file extension
	switch ext {
	case ".ir":
		// If the output name ends with .ir, set format to IR
		format = "ir"
	case ".o", ".obj", ".out":
		// If the output name ends with .obj, maintain the default format
	default:
		// If no valid extension is provided and name is not empty, return an error
		err = fmt.Errorf("output name must end with either .ir or .obj")
	}
	return
}
