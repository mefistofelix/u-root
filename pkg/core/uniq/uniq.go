// Copyright 2013-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package uniq implements the uniq core utility.
package uniq

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new uniq command.
func New() core.Command {
	c := &command{}
	c.Init()
	return c
}

// Run executes the command with a background context.
func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

// RunContext executes the command.
func (c *command) RunContext(ctx context.Context, args ...string) error {
	fs := flag.NewFlagSet("uniq", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	var countFlag, dupOnly, uniqueOnly, ignoreCase bool
	fs.BoolVar(&countFlag, "c", false, "prefix count")
	fs.BoolVar(&dupOnly, "d", false, "only duplicated lines")
	fs.BoolVar(&uniqueOnly, "u", false, "only unique lines")
	fs.BoolVar(&ignoreCase, "i", false, "case-insensitive")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	var r io.Reader = c.Stdin
	var w io.Writer = c.Stdout
	if fs.NArg() >= 1 {
		f, err := os.Open(c.ResolvePath(fs.Arg(0)))
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	if fs.NArg() >= 2 {
		f, err := os.Create(c.ResolvePath(fs.Arg(1)))
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	eq := func(a, b string) bool {
		if ignoreCase {
			return strings.EqualFold(a, b)
		}
		return a == b
	}
	flush := func(line string, n int) {
		if n == 0 {
			return
		}
		if dupOnly && n == 1 {
			return
		}
		if uniqueOnly && n > 1 {
			return
		}
		if countFlag {
			fmt.Fprintf(w, "%7d %s\n", n, line)
		} else {
			fmt.Fprintln(w, line)
		}
	}

	scanner := bufio.NewScanner(r)
	prev := ""
	run := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		line := scanner.Text()
		if run == 0 {
			prev = line
			run = 1
		} else if eq(line, prev) {
			run++
		} else {
			flush(prev, run)
			prev = line
			run = 1
		}
	}
	flush(prev, run)
	return nil
}
