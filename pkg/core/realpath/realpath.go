// Copyright 2012-2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package realpath implements the realpath core utility.
package realpath

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
	quiet bool
}

// New creates a new realpath command.
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

	fs := flag.NewFlagSet("realpath", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	fs.BoolVar(&c.quiet, "q", false, "quiet mode")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var errs error
	for _, arg := range fs.Args() {
		target := c.ResolvePath(arg)
		abs_path, err := filepath.Abs(target)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		real_path, err := filepath.EvalSymlinks(abs_path)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if _, err := fmt.Fprintln(c.Stdout, filepath.Clean(real_path)); err != nil {
			return err
		}
	}
	if errs != nil && c.quiet {
		return os.ErrInvalid
	}
	return errs
}
