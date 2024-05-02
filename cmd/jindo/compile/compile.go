// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package compile

import (
	"context"
	"jindo-tool/command"
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
	// init compiler
	// load source paths

	print(args)

	//compiler := NewCompiler()
}

type Compiler struct {
	WorkDir string
}

func NewCompiler() *Compiler {
	return nil
}
