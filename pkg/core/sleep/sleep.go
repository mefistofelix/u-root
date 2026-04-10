// Copyright 2017-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package sleep implements the sleep core utility.
package sleep

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new sleep command.
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
	if len(args) < 1 {
		return fmt.Errorf("missing operand")
	}
	s := args[0]
	d, err := time.ParseDuration(s)
	if err != nil {
		secs, nerr := strconv.ParseFloat(s, 64)
		if nerr != nil {
			return fmt.Errorf("invalid duration: %s", s)
		}
		d = time.Duration(secs * float64(time.Second))
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
	}
	return nil
}
