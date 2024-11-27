// This Package is Deprecated
package codegen

import (
	"jindo-tool/compile/ast"
	"jindo-tool/compile/base"

	"tinygo.org/x/go-llvm"
)

type (
	File ast.File

	// Declarations
	ImportDecl ast.ImportDecl
	TypeDecl   ast.TypeDecl
	VarDecl    ast.VarDecl
	FuncDecl   ast.FuncDecl

	// Statements
	BlockStmt ast.BlockStmt
	Stmt      ast.Stmt
)

var nest int = 0

func AddGlobalConstant(typ llvm.Type, name string) llvm.Value {
	v := llvm.AddGlobal(base.Module, typ, name)
	v.SetGlobalConstant(true)
	return v
}

func AddGlobalVariable(typ llvm.Type, name string) llvm.Value {
	v := llvm.AddGlobal(base.Module, typ, name)
	return v
}

func InitializeSpaceModule(name string) {
	base.Module = llvm.NewContext().NewModule(name)
	base.Builder = base.Module.Context().NewBuilder()

	v := AddGlobalConstant(base.Module.Context().Int1Type(), "true")
	zero := llvm.ConstInt(base.Module.Context().Int1Type(), 0, false)
	true_ := llvm.ConstICmp(llvm.IntPredicate(llvm.IntEQ), zero, zero)
	v.SetInitializer(true_)

	// init_fn := llvm.AddFunction(base.Module, "global_init", base.Module.Context().Int32Type())
	// block := llvm.AddBasicBlock(init_fn, "")

	// base.Builder.SetInsertPointAtEnd(block)
	// zero := llvm.ConstInt(base.Module.Context().Int1Type(), 0, false)

	// true_ := llvm.ConstICmp(llvm.IntPredicate(llvm.IntEQ), zero, zero)

}

func (file *File) Codegen() {
	for _, decl := range file.DeclList {
		switch d := decl.(type) {
		// case *ImportDecl:
		// 	d.codegen()
		// case *TypeDecl:
		// 	d.codegen()
		case *ast.VarDecl:
			(*VarDecl)(d).codegen()
		case *ast.FuncDecl:

			(*FuncDecl)(d).codegen()
		}
	}

	base.Module.Dump()
}

func (decl *FuncDecl) codegen() {
	nest++
	defer func() { nest-- }()

	fn_type := llvm.FunctionType(decl.parseFuncType(), nil, false)
	fn := llvm.AddFunction(base.Module, decl.Name.Value, fn_type)
	fn.SetFunctionCallConv(llvm.CallConv(llvm.CCallConv))

	fn_entry := llvm.AddBasicBlock(fn, "entry")
	base.Builder.SetInsertPointAtEnd(fn_entry)
	base.Builder.CreateAlloca(base.Module.Context().Int32Type(), "local_var")

	(*BlockStmt)(decl.Body).codegen(fn)
}

func (fn *FuncDecl) parseFuncType() llvm.Type {
	return llvm.FunctionType(base.Module.Context().VoidType(), nil, false)
}

func (decl *VarDecl) codegen() {
	llvm.AddGlobal(base.Module, llvm.GlobalContext().FloatType(), decl.NameList.Value)
}

func (decl *BlockStmt) codegen(fn llvm.Value) {
	nest++
	defer func() { nest-- }()

	ir_block := llvm.AddBasicBlock(fn, "")
	base.Builder.SetInsertPointAtEnd(ir_block)

	for _, stmt := range decl.StmtList {
		base.Builder.CreateAlloca(base.Module.Context().Int32Type(), "local_var")
		codegen(stmt, fn)
	}
}

func codegen(stmt Stmt, fn llvm.Value) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		// codegenExpr(s.X, fn)
	// case *EmptyStmt:
	// case *IncDecStmt:
	// case *ContinueStmt:
	// case *BreakStmt:
	// case *ReturnStmt:
	// case *DeclStmt:
	// case *DefineStmt:
	// case *AssignStmt:
	// case *IfStmt:
	// case *ForStmt:
	// case *WhileStmt:
	// case *simpleStmt:
	case *BlockStmt:
		(*BlockStmt)(s).codegen(fn)
	default:
		// panic("unreachable")
	}
}

// func codegenExpr(expr ast.Expr, fn llvm.Value) llvm.Value {
// 	switch e := expr.(type) {
// 	case *ast.Name:
// 		//TODO implement me
// 		panic("implement me")
// 	case *ast.BasicLit:
// 		//TODO implement me
// 		panic("implement me")
// 	case *ast.SliceLit:
// 		//TODO implement me
// 		panic("implement me")
// 	case *ast.Operation:
// 		//TODO implement me
// 		panic("implement me")
// 	default:
// 		panic("unreachable")
// 	}
// }
