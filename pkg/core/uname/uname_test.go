// Copyright 2023-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uname

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name      string
		all       bool
		kernel    bool
		node      bool
		release   bool
		version   bool
		machine   bool
		processor bool
		expected  params
	}{
		{
			name: "flag a",
			all:  true,
			expected: params{
				kernel:  true,
				node:    true,
				release: true,
				version: true,
				machine: true,
			},
		},
		{
			name: "no flags",
			expected: params{
				kernel: true,
			},
		},
		{
			name:      "flag p",
			processor: true,
			expected: params{
				machine: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parse_params(tt.all, tt.kernel, tt.node, tt.release, tt.version, tt.machine, tt.processor)
			if got != tt.expected {
				t.Fatalf("expected %+v, got %+v", tt.expected, got)
			}
		})
	}
}

func TestHandleFlags(t *testing.T) {
	info := unameInfo{
		kernel:  "kernel",
		node:    "node",
		release: "release",
		version: "version",
		machine: "machine",
	}
	got := handle_flags(info, params{kernel: true, release: true, machine: true})
	if got != "kernel release machine" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestRunContext(t *testing.T) {
	stdout := &bytes.Buffer{}
	cmd := &command{}
	cmd.Init()
	cmd.Stdout = stdout
	cmd.info_provider = func() (unameInfo, error) {
		return unameInfo{
			kernel:  "kernel",
			node:    "node",
			release: "release",
			version: "version",
			machine: "machine",
		}, nil
	}

	if err := cmd.RunContext(t.Context(), "-a"); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	got := strings.TrimSpace(stdout.String())
	if got != "kernel node release version machine" {
		t.Fatalf("unexpected run output: %q", got)
	}
}
