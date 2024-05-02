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
	"io"
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

	// resolve_import_path: space.files... => f
	// 		# check path string error
	// 		if any( (f...).importPaths...).is(wrong_string)
	// 			abort
	//
	// 		# try generating import metadata
	// 		import(f...) => res[ importData ]
	//  	if any(res...).has_error()
	//  		abort
	//
	// ===> space.imports += f

	comp := NewCompiler(false, nil)
	err = comp.compile(ctx, format, space)
	if err != nil {
		panic(err)
	}
	comp.dump(name)
	os.Exit(0)
}

type Compiler struct {
	cwd           string
	space         *Space
	writer        io.Writer
	resolved      bool
	compileResult []byte
}

func NewCompiler() *Compiler {
	return nil
}

func (c *Compiler) compile(ctx context.Context, format string, space *Space) error {
	//panic("compile")
	return nil
}

func (c *Compiler) dump(oname string) {
	if !c.resolved {
		panic("cannot dump. compiler not resolved")
	}

	// means file writing
	if c.writer == nil {

		// means format == obj
		if oname == "" {
			// TODO: naming based on first file input
			oname = c.space.Name + ".obj"
		}

		outFile, err := os.OpenFile(oname, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic("Failed to open or create file: " + err.Error())
		}
		defer outFile.Close() // Ensure that the file is closed when all operations are done

		c.writer = outFile
	}

	_, err := c.writer.Write(c.compileResult)
	if err != nil {
		panic(err)
	}
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
