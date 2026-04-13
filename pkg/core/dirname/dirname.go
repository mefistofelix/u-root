// Copyright 2016-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dirname implements the dirname core utility.
package dirname

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/u-root/u-root/pkg/core"
)

var err_no_arg = errors.New("missing operand")

type command struct {
	core.Base
}

// New creates a new dirname command.
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
	_ = ctx

	fs := flag.NewFlagSet("dirname", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(fs.Args()) < 1 {
		return err_no_arg
	}
	for _, name := range fs.Args() {
		if _, err := fmt.Fprintln(c.Stdout, filepath.Dir(name)); err != nil {
			return err
		}
	}
	return nil
}
