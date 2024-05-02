// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package main

import (
	"flag"
	"jindo-tool/command"
)

var Jindo = command.Command{
	UsageLine: "jindo",
	Long:      `Jindo is a tool for managing Jindo source code`,
}

func init() {
	Jindo.Commands = []*command.Command{}
}

func mainUsage() {
	_ = "call help usage"
	var i int64 = -1
	_ = 1 << i
}

func main() {
	args := flag.Args()
	if len(args) < 1 {
		mainUsage()
	}
}
