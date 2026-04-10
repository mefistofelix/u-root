// Copyright 2012-2024 the u-root Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package head implements the head core utility.
package head

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"github.com/u-root/u-root/pkg/core"
)

type command struct {
	core.Base
}

// New creates a new head command.
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

	fs := flag.NewFlagSet("head", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	numLines := fs.Int("n", 10, "number of lines to print")
	numBytes := fs.Int("c", 0, "number of bytes to print")

	// Support legacy -N shorthand (e.g. head -5 file).
	remaining := args
	if len(remaining) > 0 && len(remaining[0]) > 1 && remaining[0][0] == '-' {
		n, err := strconv.Atoi(remaining[0][1:])
		if err == nil {
			*numLines = n
			remaining = remaining[1:]
		}
	}

	if err := fs.Parse(remaining); err != nil {
		return err
	}

	printHead := func(r io.Reader) error {
		if *numBytes > 0 {
			buf := make([]byte, *numBytes)
			n, err := io.ReadFull(r, buf)
			if n > 0 {
				if _, werr := c.Stdout.Write(buf[:n]); werr != nil {
					return werr
				}
			}
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}
		scanner := bufio.NewScanner(r)
		for i := 0; i < *numLines && scanner.Scan(); i++ {
			if _, err := fmt.Fprintln(c.Stdout, scanner.Text()); err != nil {
				return err
			}
		}
		return scanner.Err()
	}

	files := fs.Args()
	if len(files) == 0 {
		return printHead(c.Stdin)
	}
	for i, name := range files {
		if len(files) > 1 {
			header := fmt.Sprintf("==> %s <==\n", name)
			if _, err := io.WriteString(c.Stdout, header); err != nil {
				return err
			}
		}
		f, err := os.Open(c.ResolvePath(name))
		if err != nil {
			return err
		}
		err = printHead(f)
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if len(files) > 1 && i < len(files)-1 {
			if _, err := io.WriteString(c.Stdout, "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

