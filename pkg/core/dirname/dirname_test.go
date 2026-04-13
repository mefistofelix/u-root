// Copyright 2016-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dirname

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
)

func TestDirname(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{name: "missing operand", wantErr: err_no_arg},
		{name: "single path", args: []string{"a/b/c.txt"}, want: filepath.Dir("a/b/c.txt") + "\n"},
		{name: "multiple paths", args: []string{"a/b", "c"}, want: filepath.Dir("a/b") + "\n" + filepath.Dir("c") + "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := New().(*command)
			stdout := &bytes.Buffer{}
			cmd.Stdout = stdout
			err := cmd.RunContext(t.Context(), tt.args...)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if stdout.String() != tt.want {
				t.Fatalf("expected output %q, got %q", tt.want, stdout.String())
			}
		})
	}
}
