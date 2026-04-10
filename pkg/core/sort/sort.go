// Copyright 2017-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package sort implements the sort core utility.
package sort

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/u-root/u-root/pkg/core"
	"github.com/u-root/u-root/pkg/uroot/unixflag"
)

type command struct {
	core.Base
}

// New creates a new sort command.
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
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	var reverse, unique, numeric, fold, ignoreBlanks bool
	var outputFile string
	fs.BoolVar(&reverse, "r", false, "reverse")
	fs.BoolVar(&unique, "u", false, "unique")
	fs.BoolVar(&numeric, "n", false, "numeric sort")
	fs.BoolVar(&fold, "f", false, "fold case")
	fs.BoolVar(&ignoreBlanks, "b", false, "ignore leading blanks")
	fs.StringVar(&outputFile, "o", "", "output file")
	if err := fs.Parse(unixflag.ArgsToGoArgs(args)); err != nil {
		return err
	}

	var lines []string
	readLines := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
	}
	if fs.NArg() == 0 {
		readLines(c.Stdin)
	} else {
		for _, fname := range fs.Args() {
			f, err := os.Open(c.ResolvePath(fname))
			if err != nil {
				return err
			}
			readLines(f)
			f.Close()
		}
	}

	keyOf := func(s string) string {
		if ignoreBlanks {
			s = strings.TrimLeft(s, " \t")
		}
		if fold {
			s = strings.ToLower(s)
		}
		return s
	}
	sort.SliceStable(lines, func(i, j int) bool {
		a, b := keyOf(lines[i]), keyOf(lines[j])
		var less bool
		if numeric {
			af, aerr := strconv.ParseFloat(strings.Fields(a+" 0")[0], 64)
			bf, berr := strconv.ParseFloat(strings.Fields(b+" 0")[0], 64)
			if aerr == nil && berr == nil {
				less = af < bf
			} else {
				less = a < b
			}
		} else {
			less = a < b
		}
		if reverse {
			return !less
		}
		return less
	})
	if unique {
		deduped := lines[:0]
		for i, l := range lines {
			if i == 0 || keyOf(l) != keyOf(lines[i-1]) {
				deduped = append(deduped, l)
			}
		}
		lines = deduped
	}

	out := c.Stdout
	if outputFile != "" {
		f, err := os.Create(c.ResolvePath(outputFile))
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	for _, l := range lines {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		fmt.Fprintln(out, l)
	}
	return nil
}
