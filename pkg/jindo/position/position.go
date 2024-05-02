// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package position

import "fmt"

type Pos struct {
	base      *PosBase
	line, col uint
}

const PosMax = 1 << 30

// starting points for line and column numbers
const linebase = 1
const Colbase = 1

func MakePos(base *PosBase, line, col uint) Pos {
	return Pos{base, line, col}
}
func NewLineBase(pos Pos, filename string, line, col uint) *PosBase {
	return &PosBase{pos, filename, sat32(line), sat32(col)}
}

func NewFileBase(filename string) *PosBase {
	base := &PosBase{MakePos(nil, linebase, Colbase), filename, linebase, Colbase}
	base.pos.base = base
	return base
}

func (p Pos) String() string {
	return fmt.Sprintf("%s:%d:%d", p.base.Filename(), p.line, p.col)
}

type PosBase struct {
	pos       Pos
	filename  string
	line, col uint32
}

func (b PosBase) Filename() string {
	return b.filename
}

// func (pos pos) IsKnown() bool  { return pos.line > 0 }

func (p Pos) Pos() Pos      { return p }
func (p Pos) Line() uint    { return p.line }
func (p Pos) Col() uint     { return p.col }
func (p Pos) IsKnown() bool { return p.line > 0 }

func sat32(x uint) uint32 {
	if x > PosMax {
		return PosMax
	}
	return uint32(x)
}
