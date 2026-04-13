// Copyright 2023-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package uname

import (
	"os"
	"runtime"
)

func (c *command) collect_info() (unameInfo, error) {
	node, err := os.Hostname()
	if err != nil {
		return unameInfo{}, err
	}
	return unameInfo{
		kernel:  "Windows",
		node:    node,
		release: runtime.GOOS,
		version: runtime.Version(),
		machine: runtime.GOARCH,
	}, nil
}
