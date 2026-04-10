// Copyright 2018-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package tr implements the tr core utility.
package tr

import (
	"bufio"
	"context"
	"flag"
	"fmt"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new tr command.
func New() core.Command {
	c := &command{}
	c.Init()
	return c
}

// Run executes the command with a background context.
func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

// expandSet expands a set string supporting character ranges like a-z.
func expandSet(s string) []rune {
	runes := []rune(s)
	out := []rune{}
	for i := 0; i < len(runes); i++ {
		if i+2 < len(runes) && runes[i+1] == '-' {
			lo, hi := runes[i], runes[i+2]
			for ch := lo; ch <= hi; ch++ {
				out = append(out, ch)
			}
			i += 2
		} else {
			out = append(out, runes[i])
		}
	}
	return out
}

// RunContext executes the command.
func (c *command) RunContext(_ context.Context, args ...string) error {
	fs := flag.NewFlagSet("tr", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	var deleteMode, squeeze bool
	fs.BoolVar(&deleteMode, "d", false, "delete characters in set1")
	fs.BoolVar(&squeeze, "s", false, "squeeze repeated characters")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	reader := bufio.NewReader(c.Stdin)

	if deleteMode {
		if fs.NArg() < 1 {
			return fmt.Errorf("missing set1")
		}
		del := map[rune]bool{}
		for _, r := range expandSet(fs.Arg(0)) {
			del[r] = true
		}
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				break
			}
			if !del[r] {
				fmt.Fprintf(c.Stdout, "%c", r)
			}
		}
		return nil
	}

	if fs.NArg() < 2 {
		return fmt.Errorf("missing set1 or set2")
	}
	set1 := expandSet(fs.Arg(0))
	set2 := expandSet(fs.Arg(1))
	table := map[rune]rune{}
	for i, r := range set1 {
		if i < len(set2) {
			table[r] = set2[i]
		} else {
			table[r] = set2[len(set2)-1]
		}
	}
	prev := rune(-1)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}
		if mapped, ok := table[r]; ok {
			r = mapped
		}
		if squeeze && r == prev {
			continue
		}
		fmt.Fprintf(c.Stdout, "%c", r)
		prev = r
	}
	return nil
}
