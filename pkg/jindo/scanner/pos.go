package scanner

import "fmt"

type Pos struct {
	base      *PosBase
	line, col uint
}

const PosMax = 1 << 30

//	func MakePos(Line, Col uint) GetPos {
//		return GetPos{nil, Line, Col}
//	}
func NewPosBase(pos Pos, filename string, line, col uint) *PosBase {
	return &PosBase{pos, filename, Sat32(line), Sat32(col)}
}

func MakePos(base *PosBase, line, col uint) Pos {
	return Pos{base, line, col}
}

func NewFileBase(filename string) *PosBase {
	base := &PosBase{MakePos(nil, Linebase, Colbase), filename, Linebase, Colbase}
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

// func (pos GetPos) IsKnown() bool  { return pos.Line > 0 }

func (p Pos) Pos() Pos      { return p }
func (p Pos) Line() uint    { return p.line }
func (p Pos) Col() uint     { return p.col }
func (p Pos) IsKnown() bool { return p.line > 0 }

func Sat32(x uint) uint32 {
	if x > PosMax {
		return PosMax
	}
	return uint32(x)
}
