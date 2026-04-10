// Copyright 2015-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package date implements the date core utility.
package date

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new date command.
func New() core.Command {
	c := &command{}
	c.Init()
	return c
}

// Run executes the command with a background context.
func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

// strftimeToGo converts a strftime-style format string to Go's time layout.
var strftimeReplacer = strings.NewReplacer(
	"%Y", "2006",
	"%m", "01",
	"%d", "02",
	"%H", "15",
	"%M", "04",
	"%S", "05",
	"%Z", "MST",
	"%A", "Monday",
	"%a", "Mon",
	"%B", "January",
	"%b", "Jan",
	"%j", "002",
	"%e", "2",
	"%T", "15:04:05",
	"%D", "01/02/06",
	"%F", "2006-01-02",
	"%R", "15:04",
	"%n", "\n",
	"%t", "\t",
)

// RunContext executes the command.
func (c *command) RunContext(_ context.Context, args ...string) error {
	now := time.Now()
	layout := "Mon Jan  2 15:04:05 MST 2006"
	if len(args) >= 2 && strings.HasPrefix(args[1], "+") {
		layout = strftimeReplacer.Replace(args[1][1:])
	}
	fmt.Fprintln(c.Stdout, now.Format(layout))
	return nil
}
