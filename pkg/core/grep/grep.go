// Copyright 2012-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package grep implements the grep core utility.
package grep

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/u-root/u-root/pkg/core"
	"github.com/u-root/u-root/pkg/uroot/unixflag"
)

type params struct {
	invert, ignoreCase, lineNum, filesOnly, countOnly, fixed, quiet, recursive bool
	patterns                                                                    []string
}

type command struct {
	core.Base
	params
}

// New creates a new grep command.
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
	fs_set := flag.NewFlagSet("grep", flag.ContinueOnError)
	fs_set.SetOutput(c.Stderr)
	var invert, ignoreCase, lineNum, filesOnly, countOnly, fixed, quiet, recursive bool
	var expr string
	fs_set.BoolVar(&invert, "v", false, "invert match")
	fs_set.BoolVar(&ignoreCase, "i", false, "case-insensitive")
	fs_set.BoolVar(&lineNum, "n", false, "print line numbers")
	fs_set.BoolVar(&filesOnly, "l", false, "print only filenames")
	fs_set.BoolVar(&countOnly, "c", false, "print match count")
	fs_set.BoolVar(&fixed, "F", false, "fixed strings")
	fs_set.BoolVar(&quiet, "q", false, "quiet")
	fs_set.BoolVar(&recursive, "r", false, "recursive")
	fs_set.StringVar(&expr, "e", "", "pattern")
	if err := fs_set.Parse(unixflag.ArgsToGoArgs(args)); err != nil {
		return err
	}
	positional := fs_set.Args()
	patterns := []string{}
	if expr != "" {
		patterns = append(patterns, expr)
	}
	if len(patterns) == 0 && len(positional) > 0 {
		patterns = append(patterns, positional[0])
		positional = positional[1:]
	}
	if len(patterns) == 0 {
		return fmt.Errorf("missing pattern")
	}

	combined := strings.Join(patterns, "|")
	if fixed {
		parts := make([]string, len(patterns))
		for i, p := range patterns {
			parts[i] = regexp.QuoteMeta(p)
		}
		combined = strings.Join(parts, "|")
	}
	if ignoreCase {
		combined = "(?i)" + combined
	}
	re, err := regexp.Compile(combined)
	if err != nil {
		return err
	}

	files := []string{}
	if len(positional) == 0 {
		files = append(files, "-")
	} else {
		for _, p := range positional {
			resolved := c.ResolvePath(p)
			if recursive {
				_ = filepath.WalkDir(resolved, func(path string, d fs.DirEntry, err error) error {
					if err == nil && !d.IsDir() {
						files = append(files, path)
					}
					return nil
				})
			} else {
				files = append(files, resolved)
			}
		}
	}
	show_name := len(files) > 1
	matched_any := false
	for _, fname := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var r io.Reader
		if fname == "-" {
			r = c.Stdin
		} else {
			f, err := os.Open(fname)
			if err != nil {
				fmt.Fprintln(c.Stderr, "grep:", err)
				continue
			}
			defer f.Close()
			r = f
		}
		scanner := bufio.NewScanner(r)
		lineNo := 0
		matchCount := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			hit := re.MatchString(line)
			if invert {
				hit = !hit
			}
			if !hit {
				continue
			}
			matched_any = true
			matchCount++
			if quiet || filesOnly || countOnly {
				continue
			}
			prefix := ""
			if show_name {
				prefix = fname + ":"
			}
			if lineNum {
				prefix += fmt.Sprintf("%d:", lineNo)
			}
			fmt.Fprintln(c.Stdout, prefix+line)
		}
		if filesOnly && matchCount > 0 {
			fmt.Fprintln(c.Stdout, fname)
		}
		if countOnly {
			if show_name {
				fmt.Fprintf(c.Stdout, "%s:%d\n", fname, matchCount)
			} else {
				fmt.Fprintln(c.Stdout, matchCount)
			}
		}
	}
	if !matched_any {
		return fmt.Errorf("no match")
	}
	return nil
}
