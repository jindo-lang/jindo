// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package command

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
)

// A Command is an implementation of a jindo command
// like jindo build or jindo fix.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(ctx context.Context, cmd *Command, args []string)

	// UsageLine is the one-line usage message.
	// The words between "jindo" and the first flag or argument in the line are taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'jindo help' output.
	Short string

	// Long is the long message shown in the 'jindo help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool

	// Commands lists the available commands and help topics.
	// The order here is the order in which they are printed by 'jindo help'.
	// Note that subcommands are in general best avoided.
	Commands []*Command
}

// Lookup returns the subcommand with the given name, if any.
// Otherwise it returns nil.
//
// Lookup ignores subcommands that have len(c.Commands) == 0 and c.Run == nil.
// Such subcommands are only for use as arguments to "help".
func (c *Command) Lookup(name string) *Command {
	for _, sub := range c.Commands {
		if sub.Name() == name && (len(c.Commands) > 0 || c.Runnable()) {
			return sub
		}
	}
	return nil
}

// hasFlag reports whether a command or any of its subcommands contain the given
// flag.
func hasFlag(c *Command, name string) bool {
	if f := c.Flag.Lookup(name); f != nil {
		return true
	}
	for _, sub := range c.Commands {
		if hasFlag(sub, name) {
			return true
		}
	}
	return false
}

// LongName returns the command's long name: all the words in the usage line between "jindo" and a flag or argument,
func (c *Command) LongName() string {
	name := c.UsageLine
	if i := strings.Index(name, " ["); i >= 0 {
		name = name[:i]
	}
	if name == "jindo" {
		return ""
	}
	return strings.TrimPrefix(name, "jindo ")
}

// Name returns the command's short name: the last word in the usage line before a flag or argument.
func (c *Command) Name() string {
	name := c.LongName()
	if i := strings.LastIndex(name, " "); i >= 0 {
		name = name[i+1:]
	}
	return name
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "Run 'jindo help %s' for details.\n", c.LongName())
	SetExitStatus(2)
	Exit()
}

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

var atExitFuncs []func()

func AtExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

func Exit() {
	for _, f := range atExitFuncs {
		f()
	}
	os.Exit(exitStatus)
}

func Fatalf(format string, args ...any) {
	Errorf(format, args...)
	Exit()
}

func Errorf(format string, args ...any) {
	log.Printf(format, args...)
	SetExitStatus(1)
}

func ExitIfErrors() {
	if exitStatus != 0 {
		Exit()
	}
}

func Error(err error) {
	// We use errors.Join to return multiple errors from various routines.
	// If we receive multiple errors joined with a basic errors.Join,
	// handle each one separately so that they all have the leading "jindo: " prefix.
	// A plain interface check is not good enough because there might be
	// other kinds of structured errors that are logically one unit and that
	// add other context: only handling the wrapped errors would lose
	// that context.
	if err != nil && reflect.TypeOf(err).String() == "*errors.joinError" {
		for _, e := range err.(interface{ Unwrap() []error }).Unwrap() {
			Error(e)
		}
		return
	}
	Errorf("jindo: %v", err)
}

func Fatal(err error) {
	Error(err)
	Exit()
}

var exitStatus = 0
var exitMu sync.Mutex

func SetExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

func GetExitStatus() int {
	return exitStatus
}

// Run runs the command, with stdout and stderr
// connected to the jindo command's own stdout and stderr.
// If the command fails, Run reports the error using Errorf.
func Run(cmdargs ...any) {
	//TODO Fix this
	//cmdline := str.StringList(cmdargs...)
	//if cfg.BuildN || cfg.BuildX {
	//	fmt.Printf("%s\n", strings.Join(cmdline, " "))
	//	if cfg.BuildN {
	//		return
	//	}
	//}
	//
	//cmd := exec.Command(cmdline[0], cmdline[1:]...)
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//if err := cmd.Run(); err != nil {
	//	Errorf("%v", err)
	//}
}

// RunStdin is like run but connects Stdin.
func RunStdin(cmdline []string) {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// TODO: Fix this
	//cmd.Env = cfg.OrigEnv
	//StartSigHandlers()
	if err := cmd.Run(); err != nil {
		Errorf("%v", err)
	}
}

// Usage is the usage-reporting function, filled in by package main
// but here for reference by other packages.
var Usage func()