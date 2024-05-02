// Copyright 2024 The Jindo Authors. All rights reserved.
// This file is part of jindo and is licensed under
// the GNU General Public License version 3, which is available at
// https://www.gnu.org/licenses/gpl-3.0.html or in the LICENSE file
// located in the root directory of this source tree.

package compile

import (
	"context"
	"fmt"
	"testing"
)

type field struct {
	name string
	args []string
}

const expath = "./testdata/"

var valids = []field{
	{"same-dir-same-space", []string{"main0.paw", "main1.paw"}},
}
var invalids = []field{
	{"same-dir-diff-space", []string{"test0.paw", "main1.paw"}},
	{"diff-dir-same-space", []string{"pkg0/main0.paw", "main1.paw"}},
	{"wrong-extension", []string{"wrong.file"}},
}

func exargs(args []string) {
	for i, arg := range args {
		args[i] = fmt.Sprintf(expath + arg)
	}
}

func Test_runCompile(t *testing.T) {
	for _, tt := range valids {
		func() {
			defer func() {
				e := recover()
				if e != nil {
					t.Errorf("got error: %v", e)
				}
			}()
			ctx := context.Background()
			FlagO = ""
			exargs(tt.args)
			runCompile(ctx, CmdCompile, tt.args)
		}()
	}

	for _, tt := range invalids {
		func() {
			err := true
			defer func() {
				e := recover()
				if !err {
					t.Error("no error")
				}
				fmt.Println("[good] got error: ", e)
			}()
			ctx := context.Background()
			FlagO = ""
			exargs(tt.args)
			runCompile(ctx, CmdCompile, tt.args)

			err = false
			panic(nil)
		}()
	}
}
