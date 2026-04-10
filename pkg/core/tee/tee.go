// Copyright 2013-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package tee implements the tee core utility.
package tee

import (
	"context"
	"flag"
	"io"
	"os"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new tee command.
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
func (c *command) RunContext(_ context.Context, args ...string) error {
	fs := flag.NewFlagSet("tee", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	var appendMode bool
	fs.BoolVar(&appendMode, "a", false, "append")
	if err := fs.Parse(args); err != nil {
		return err
	}

	oflags := os.O_WRONLY | os.O_CREATE
	if appendMode {
		oflags |= os.O_APPEND
	} else {
		oflags |= os.O_TRUNC
	}
	writers := []io.Writer{c.Stdout}
	var closers []io.Closer
	for _, fname := range fs.Args() {
		f, err := os.OpenFile(c.ResolvePath(fname), oflags, 0o644)
		if err != nil {
			return err
		}
		writers = append(writers, f)
		closers = append(closers, f)
	}
	_, err := io.Copy(io.MultiWriter(writers...), c.Stdin)
	for _, cl := range closers {
		cl.Close()
	}
	return err
}
