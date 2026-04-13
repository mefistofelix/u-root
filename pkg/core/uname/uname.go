// Copyright 2023-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package uname implements the uname core utility.
package uname

import (
	"context"
	"flag"
	"fmt"

	"github.com/u-root/u-root/pkg/core"
)

type params struct {
	kernel  bool
	node    bool
	release bool
	version bool
	machine bool
}

type unameInfo struct {
	kernel  string
	node    string
	release string
	version string
	machine string
}

type command struct {
	core.Base
	info_provider func() (unameInfo, error)
}

// New creates a new uname command.
func New() core.Command {
	c := &command{}
	c.Init()
	c.info_provider = c.collect_info
	return c
}

// Run executes the command with a background context.
func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

// RunContext executes the command.
func (c *command) RunContext(ctx context.Context, args ...string) error {
	_ = ctx

	fs := flag.NewFlagSet("uname", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	all := fs.Bool("a", false, "print everything")
	kernel := fs.Bool("s", false, "print the kernel name")
	node := fs.Bool("n", false, "print the network node name")
	release := fs.Bool("r", false, "print the kernel release")
	version := fs.Bool("v", false, "print the kernel version")
	machine := fs.Bool("m", false, "print the machine hardware name")
	processor := fs.Bool("p", false, "print the machine hardware name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	info, err := c.info_provider()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.Stdout, handle_flags(info, parse_params(*all, *kernel, *node, *release, *version, *machine, *processor)))
	return err
}

func parse_params(all, kernel, node, release, version, machine, processor bool) params {
	p := params{
		kernel:  kernel || all,
		node:    node || all,
		release: release || all,
		version: version || all,
		machine: machine || processor || all,
	}
	if !p.kernel && !p.node && !p.release && !p.version && !p.machine {
		p.kernel = true
	}
	return p
}

func handle_flags(info unameInfo, p params) string {
	fields := make([]string, 0, 5)
	if p.kernel {
		fields = append(fields, info.kernel)
	}
	if p.node {
		fields = append(fields, info.node)
	}
	if p.release {
		fields = append(fields, info.release)
	}
	if p.version {
		fields = append(fields, info.version)
	}
	if p.machine {
		fields = append(fields, info.machine)
	}
	return join_fields(fields)
}

func join_fields(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	out := fields[0]
	for i := 1; i < len(fields); i++ {
		out += " " + fields[i]
	}
	return out
}
