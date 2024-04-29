// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package parser implements a parser for Jindo source files. Input may be
// provided in a variety of forms (see the various Parse* functions); the
// output is an abstract syntax tree (AST) representing the Jindo source. The
// parser is invoked through one of the Parse* functions.
//
package parser

import (
	_ "jindo/pkg/jindo/scanner"
	_ "jindo/pkg/jindo/token"
)

