// Copyright 2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cut implements the cut core utility.
package cut

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new cut command.
func New() core.Command {
	c := &command{}
	c.Init()
	return c
}

// Run executes the command with a background context.
func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

// parseRanges parses a comma-separated range spec like "1,3-5,7-".
func parseRanges(spec string) [][2]int {
	var ranges [][2]int
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "-"); idx >= 0 {
			lo, _ := strconv.Atoi(part[:idx])
			hiStr := part[idx+1:]
			if hiStr == "" {
				ranges = append(ranges, [2]int{lo, 1<<31 - 1})
			} else {
				hi, _ := strconv.Atoi(hiStr)
				ranges = append(ranges, [2]int{lo, hi})
			}
		} else {
			n, _ := strconv.Atoi(part)
			ranges = append(ranges, [2]int{n, n})
		}
	}
	return ranges
}

func inRange(n int, ranges [][2]int) bool {
	for _, r := range ranges {
		if n >= r[0] && n <= r[1] {
			return true
		}
	}
	return false
}

// RunContext executes the command.
func (c *command) RunContext(ctx context.Context, args ...string) error {
	fs := flag.NewFlagSet("cut", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	var fieldsSpec, charsSpec, delim string
	delim = "\t"
	fs.StringVar(&fieldsSpec, "f", "", "fields")
	fs.StringVar(&charsSpec, "c", "", "characters")
	fs.StringVar(&delim, "d", "\t", "delimiter")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	process := func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		if fieldsSpec != "" {
			ranges := parseRanges(fieldsSpec)
			for scanner.Scan() {
				parts := strings.Split(scanner.Text(), delim)
				out := []string{}
				for i, p := range parts {
					if inRange(i+1, ranges) {
						out = append(out, p)
					}
				}
				fmt.Fprintln(c.Stdout, strings.Join(out, delim))
			}
		} else if charsSpec != "" {
			ranges := parseRanges(charsSpec)
			for scanner.Scan() {
				runes := []rune(scanner.Text())
				out := []rune{}
				for i, r := range runes {
					if inRange(i+1, ranges) {
						out = append(out, r)
					}
				}
				fmt.Fprintln(c.Stdout, string(out))
			}
		}
	}

	if fs.NArg() == 0 {
		process(c.Stdin)
	} else {
		for _, fname := range fs.Args() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			f, err := os.Open(c.ResolvePath(fname))
			if err != nil {
				return err
			}
			process(f)
			f.Close()
		}
	}
	return nil
}
