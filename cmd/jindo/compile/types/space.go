// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	// "cmd/internal/obj"
	// "cmd/internal/objabi"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// PathToPrefix converts raw string to the prefix that will be used in the
// symbol table. All control characters, space, '%' and '"', as well as
// non-7-bit clean bytes turn into %xx. The period needs escaping only in the
// last segment of the path, and it makes for happier users if we escape that as
// little as possible.
func PathToPrefix(s string) string {
	slash := strings.LastIndex(s, "/")
	// check for chars that need escaping
	n := 0
	for r := 0; r < len(s); r++ {
		if c := s[r]; c <= ' ' || (c == '.' && r > slash) || c == '%' || c == '"' || c >= 0x7F {
			n++
		}
	}

	// quick exit
	if n == 0 {
		return s
	}

	// escape
	const hex = "0123456789abcdef"
	p := make([]byte, 0, len(s)+2*n)
	for r := 0; r < len(s); r++ {
		if c := s[r]; c <= ' ' || (c == '.' && r > slash) || c == '%' || c == '"' || c >= 0x7F {
			p = append(p, '%', hex[c>>4], hex[c&0xF])
		} else {
			p = append(p, c)
		}
	}

	return string(p)
}

// spaceMap maps a space path to a space.
var spaceMap = make(map[string]*Space)

type Space struct {
	Path   string // string literal used in import statement, e.g. "internal/runtime/sys"
	Name   string // space name, e.g. "sys"
	Prefix string // escaped path for use in symbol table
	Syms   map[string]*Sym
	// Pathsym *obj.LSym

	Direct bool // imported directly
}

// NewSpace returns a new space for the given space path and name.
// Unless name is the empty string, if the space exists already,
// the existing space name and the provided name must match.
func NewSpace(path, name string) *Space {
	if p := spaceMap[path]; p != nil {
		if name != "" && p.Name != name {
			panic(fmt.Sprintf("conflicting space names %s and %s for path %q", p.Name, name, path))
		}
		return p
	}

	p := new(Space)
	p.Path = path
	p.Name = name
	if path == "go.shape" {
		// Don't escape "go.shape", since it's not needed (it's a builtin
		// space), and we don't want escape codes showing up in shape type
		// names, which also appear in names of function/method
		// instantiations.
		p.Prefix = path
	} else {
		p.Prefix = PathToPrefix(path)
	}
	p.Syms = make(map[string]*Sym)
	spaceMap[path] = p

	return p
}

func SpaceMap() map[string]*Space {
	return spaceMap
}

var nospace = &Space{
	Syms: make(map[string]*Sym),
}

func (space *Space) Lookup(name string) *Sym {
	s, _ := space.LookupOK(name)
	return s
}

// LookupOK looks up name in space and reports whether it previously existed.
func (space *Space) LookupOK(name string) (s *Sym, existed bool) {
	// TODO(gri) remove this check in favor of specialized lookup
	if space == nil {
		space = nospace
	}
	if s := space.Syms[name]; s != nil {
		return s, true
	}

	s = &Sym{
		Name: name,
		Space:  space,
	}
	space.Syms[name] = s
	return s, false
}

func (space *Space) LookupBytes(name []byte) *Sym {
	// TODO(gri) remove this check in favor of specialized lookup
	if space == nil {
		space = nospace
	}
	if s := space.Syms[string(name)]; s != nil {
		return s
	}
	str := InternString(name)
	return space.Lookup(str)
}

// LookupNum looks up the symbol starting with prefix and ending with
// the decimal n. If prefix is too long, LookupNum panics.
func (space *Space) LookupNum(prefix string, n int) *Sym {
	var buf [20]byte // plenty long enough for all current users
	copy(buf[:], prefix)
	b := strconv.AppendInt(buf[:len(prefix)], int64(n), 10)
	return space.LookupBytes(b)
}

// Selector looks up a selector identifier.
func (space *Space) Selector(name string) *Sym {
	if IsExported(name) {
		space = LocalSpace
	}
	return space.Lookup(name)
}

var (
	internedStringsmu sync.Mutex // protects internedStrings
	internedStrings   = map[string]string{}
)

func InternString(b []byte) string {
	internedStringsmu.Lock()
	s, ok := internedStrings[string(b)] // string(b) here doesn't allocate
	if !ok {
		s = string(b)
		internedStrings[s] = s
	}
	internedStringsmu.Unlock()
	return s
}
