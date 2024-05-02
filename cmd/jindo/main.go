// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package main

import (
	"context"
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

// lookupCmd interprets the initial elements of args
// to find a command to run (cmd.Runnable() == true)
// or else a command group that ran out of arguments
// or had an unknown subcommand (len(cmd.Commands) > 0).
// It returns that command and the number of elements of args
// that it took to arrive at that command.
func lookupCmd(args []string) (cmd *command.Command, used int) {
	cmd = Jindo
	for used < len(args) {
		c := cmd.Lookup(args[used])
		if c == nil {
			break
		}
		if c.Runnable() {
			cmd = c
			used++
			break
		}
		if len(c.Commands) > 0 {
			cmd = c
			used++
			if used >= len(args) || args[0] == "help" {
				break
			}
			continue
		}
		// len(c.Commands) == 0 && !c.Runnable() => help text; stop at "help"
		break
	}
	return cmd, used
}

func invoke(cmd *command.Command, args []string) {
	cmd.Flag.Usage = func() { cmd.Usage() }
	if cmd.CustomFlags {
		args = args[1:]
	} else {
		cmd.Flag.Parse(args[1:])
		args = cmd.Flag.Args()
	}
	// TODO: add DebugRuntimeTrace support
	//if cfg.DebugRuntimeTrace != "" {
	//	f, err := os.Create(cfg.DebugRuntimeTrace)
	//	if err != nil {
	//		base.Fatalf("creating trace file: %v", err)
	//	}
	//	if err := rtrace.Start(f); err != nil {
	//		base.Fatalf("starting event trace: %v", err)
	//	}
	//	defer func() {
	//		rtrace.Stop()
	//	}()
	//}

	ctx := context.Background()
	cmd.Run(ctx, cmd, args)
}
