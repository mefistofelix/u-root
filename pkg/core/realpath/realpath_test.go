// Copyright 2012-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package realpath

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealpath(t *testing.T) {
	dir := t.TempDir()
	resolved_dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.CreateTemp(resolved_dir, "")
	if err != nil {
		t.Fatal(err)
	}
	_ = file.Close()

	link_path := filepath.Join(resolved_dir, "symlink")
	err = os.Symlink(file.Name(), link_path)
	if err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{name: "real file", args: []string{file.Name()}, want: file.Name() + "\n"},
		{name: "symlink", args: []string{link_path}, want: file.Name() + "\n"},
		{name: "cleaned path", args: []string{resolved_dir + "/../" + filepath.Base(resolved_dir) + "/symlink"}, want: file.Name() + "\n"},
		{name: "missing file", args: []string{"filenotexists"}, wantErr: os.ErrNotExist},
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

func TestRealpathQuiet(t *testing.T) {
	cmd := New().(*command)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.RunContext(t.Context(), "-q", "filenotexists")
	if !errors.Is(err, os.ErrInvalid) {
		t.Fatalf("expected quiet error %v, got %v", os.ErrInvalid, err)
	}
	if strings.TrimSpace(stdout.String()) != "" {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
}
