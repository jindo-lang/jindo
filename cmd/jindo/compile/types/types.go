// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modified by Jindo Project, 2024.

package types

type Kind uint8

type bitset8 uint8

func (f *bitset8) set(mask uint8, b bool) {
	if b {
		*(*uint8)(f) |= mask
	} else {
		*(*uint8)(f) &^= mask
	}
}

// BuiltinSpace is a fake package that declares the universe block.
var BuiltinSpace *Space

// LocalSpace is the package being compiled.
var LocalSpace *Space

// UnsafeSpace is package unsafe.
var UnsafeSpace *Space

// BlankSym is the blank (_) symbol.
var BlankSym *Sym

const (
	Txxx Kind = iota

	TINT8
	TUINT8
	TINT16
	TUINT16
	TINT32
	TUINT32
	TINT64
	TUINT64
	TINT
	TUINT
	TUINTPTR

	// End of currently implemented types
	TAVAIL

	TCOMPLEX64
	TCOMPLEX128

	TFLOAT32
	TFLOAT64

	TBOOL

	TPTR
	TFUNC
	TSLICE
	TARRAY
	TSTRUCT
	TCHAN
	TMAP
	TINTER
	TFORW
	TANY
	TSTRING
	TUNSAFEPTR

	// pseudo-types for literals
	TIDEAL // untyped numeric constants
	TNIL
	TBLANK

	// pseudo-types used temporarily only during frame layout (CalcSize())
	TFUNCARGS
	TCHANARGS

	// SSA backend types
	TSSA     // internal types used by SSA backend (flags, memory, etc.)
	TTUPLE   // a pair of types, used by SSA backend
	TRESULTS // multiple types; the result of calling a function or method, with a memory at the end.

	NTYPE
)
