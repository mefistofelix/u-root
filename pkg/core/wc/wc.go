// Copyright 2013-2023 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package wc implements the wc core utility.
package wc

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/u-root/u-root/pkg/core"
)

type params struct {
	lines  bool
	words  bool
	runes  bool
	broken bool
	chars  bool
}

type cnt struct {
	lines, words, runes, badRunes, chars int64
}

type command struct {
	core.Base
	params
}

// New creates a new wc command.
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

	fs := flag.NewFlagSet("wc", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	fs.BoolVar(&c.lines, "l", false, "count lines")
	fs.BoolVar(&c.words, "w", false, "count words")
	fs.BoolVar(&c.runes, "r", false, "count runes")
	fs.BoolVar(&c.broken, "b", false, "count broken UTF")
	fs.BoolVar(&c.chars, "c", false, "count bytes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if !c.lines && !c.words && !c.runes && !c.broken && !c.chars {
		c.lines, c.words, c.chars = true, true, true
	}

	if len(fs.Args()) == 0 {
		res := c.count(c.Stdin, "")
		c.report(res, "")
		return nil
	}

	var totals cnt
	for _, arg := range fs.Args() {
		path := c.ResolvePath(arg)
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(c.Stderr, "wc: %s: %v\n", arg, err)
			continue
		}
		res := c.count(file, arg)
		_ = file.Close()
		totals.lines += res.lines
		totals.words += res.words
		totals.runes += res.runes
		totals.badRunes += res.badRunes
		totals.chars += res.chars
		c.report(res, arg)
	}
	if len(fs.Args()) > 1 {
		c.report(totals, "total")
	}
	return nil
}

func invalidCount(p []byte) int64 {
	var n int64
	i := 0
	for i < len(p) {
		if p[i] < utf8.RuneSelf {
			i++
			continue
		}
		_, size := utf8.DecodeRune(p[i:])
		if size == 1 {
			n++
		}
		i += size
	}
	return n
}

func (c *command) count(in io.Reader, fname string) cnt {
	b := bufio.NewReaderSize(in, 8192)
	counted := false
	count := cnt{}
	for !counted {
		line, err := b.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				counted = true
			} else {
				fmt.Fprintf(c.Stderr, "wc: %s: %v\n", fname, err)
				return cnt{}
			}
		}
		if !counted {
			count.lines++
		}
		count.words += int64(len(bytes.Fields(line)))
		count.runes += int64(utf8.RuneCount(line))
		count.chars += int64(len(line))
		count.badRunes += invalidCount(line)
	}
	return count
}

func (c *command) report(count cnt, fname string) {
	fields := []string{}
	if c.lines {
		fields = append(fields, fmt.Sprintf("%d", count.lines))
	}
	if c.words {
		fields = append(fields, fmt.Sprintf("%d", count.words))
	}
	if c.runes {
		fields = append(fields, fmt.Sprintf("%d", count.runes))
	}
	if c.broken {
		fields = append(fields, fmt.Sprintf("%d", count.badRunes))
	}
	if c.chars {
		fields = append(fields, fmt.Sprintf("%d", count.chars))
	}
	if fname != "" {
		fields = append(fields, fname)
	}
	fmt.Fprintln(c.Stdout, strings.Join(fields, " "))
}
