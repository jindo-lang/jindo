// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package main

import (
	"jindo-tool/command"
	"jindo-tool/help"
	"os"
)

var Jindo = &command.Command{
	UsageLine: "jindo",

	Long: `Jindo is a tool for managing Jindo source code`,
}

func init() {
	Jindo.Commands = []*command.Command{
		Jindo,
	}
}

func mainUsage() {
	help.PrintUsage(os.Stderr, Jindo)
	os.Exit(2)
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		mainUsage()
	}

	if args[0] == "help" {
		help.Help(os.Stdout, Jindo, args[1:])
		return
	}
}
