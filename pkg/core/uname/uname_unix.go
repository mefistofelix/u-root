// Copyright 2023-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !plan9 && !windows

package uname

import "golang.org/x/sys/unix"

func (c *command) collect_info() (unameInfo, error) {
	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return unameInfo{}, err
	}
	return unameInfo{
		kernel:  unix.ByteSliceToString(u.Sysname[:]),
		node:    unix.ByteSliceToString(u.Nodename[:]),
		release: unix.ByteSliceToString(u.Release[:]),
		version: unix.ByteSliceToString(u.Version[:]),
		machine: unix.ByteSliceToString(u.Machine[:]),
	}, nil
}
