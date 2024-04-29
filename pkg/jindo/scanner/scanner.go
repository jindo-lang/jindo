// Copyright 2024 The Jindo Authors. All rights reserved.
// Use of this source code is governed by a GPL-3 style
// license that can be found in the LICENSE file.

// Package scanner implements a scanner for Jindo source text.
// It takes a []byte as source which can then be tokenized
// through repeated calls to the Scan method.
//
package scanner

import (
	_ "jindo/pkg/jindo/token"
)

