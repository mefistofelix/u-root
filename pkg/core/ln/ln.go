// Copyright 2026 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ln implements the ln command.
package ln

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/u-root/u-root/pkg/core"
	"github.com/u-root/u-root/pkg/uroot/unixflag"
)

type command struct {
	core.Base
}

var errSkipLink = errors.New("skip link")

type flags struct {
	symlink  bool
	verbose  bool
	force    bool
	nondir   bool
	prompt   bool
	logical  bool
	physical bool
	relative bool
	dirtgt   string
}

// New creates a new ln command.
func New() core.Command {
	c := &command{}
	c.Init()
	return c
}

func (c *command) Run(args ...string) error {
	return c.RunContext(context.Background(), args...)
}

func (c *command) RunContext(ctx context.Context, args ...string) error {
	_ = ctx

	var f flags
	fs := flag.NewFlagSet("ln", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	fs.BoolVar(&f.symlink, "s", false, "make symbolic links instead of hard links")
	fs.BoolVar(&f.verbose, "v", false, "print name of each linked file")
	fs.BoolVar(&f.force, "f", false, "remove destination files")
	fs.BoolVar(&f.nondir, "T", false, "treat linkname operand as a non-dir always")
	fs.BoolVar(&f.prompt, "i", false, "prompt before overwrite")
	fs.BoolVar(&f.logical, "L", false, "dereference targets that are symbolic links")
	fs.BoolVar(&f.physical, "P", false, "make hard links directly to symbolic links")
	fs.BoolVar(&f.relative, "r", false, "create symlinks relative to link location")
	fs.StringVar(&f.dirtgt, "t", "", "specify the directory to put the links")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: ln [-svfTiLPr] TARGET LINK\n")
		fmt.Fprintf(fs.Output(), "Usage: ln [-svfTiLPr] TARGET... DIRECTORY\n")
		fmt.Fprintf(fs.Output(), "Usage: ln [-svfTiLPr] -t DIRECTORY TARGET...\n")
	}
	if err := fs.Parse(unixflag.ArgsToGoArgs(args)); err != nil {
		return err
	}
	if err := f.run(c, fs.Args()); err != nil {
		return fmt.Errorf("ln: %w", err)
	}
	return nil
}

func (f flags) run(c *command, args []string) error {
	if len(args) == 0 {
		return errors.New("missing file operand")
	}
	targets, linkName, err := f.eval_args(c, args)
	if err != nil {
		return err
	}
	linkFunc := os.Link
	if f.symlink {
		linkFunc = os.Symlink
	}
	for _, target := range targets {
		resolvedTarget, resolvedLinkName, resolvedLinkFunc, err := f.prepare_paths(c, target, linkName, linkFunc)
		if err != nil {
			return err
		}
		if err := f.maybe_remove_existing(c, resolvedLinkName); err != nil {
			if errors.Is(err, errSkipLink) {
				continue
			}
			return err
		}
		if f.relative {
			if !f.symlink {
				return errors.New("cannot do -r without -s")
			}
			relTarget, err := filepath.Rel(filepath.Dir(resolvedLinkName), resolvedTarget)
			if err != nil {
				return err
			}
			resolvedTarget = relTarget
		}
		if err := resolvedLinkFunc(resolvedTarget, resolvedLinkName); err != nil {
			return err
		}
		if f.verbose {
			fmt.Fprintf(c.Stdout, "%q -> %q\n", resolvedLinkName, resolvedTarget)
		}
	}
	return nil
}

func (f *flags) eval_args(c *command, args []string) ([]string, string, error) {
	if f.dirtgt != "" {
		return resolve_paths(c, args), "", nil
	}
	if len(args) == 1 {
		return resolve_paths(c, args), "", nil
	}
	targets := resolve_paths(c, args[:len(args)-1])
	lastArg := c.ResolvePath(args[len(args)-1])
	if info, err := os.Stat(lastArg); !f.nondir && err == nil && info.IsDir() {
		f.dirtgt = lastArg
		return targets, "", nil
	}
	return targets, lastArg, nil
}

func resolve_paths(c *command, args []string) []string {
	resolved := make([]string, 0, len(args))
	for _, arg := range args {
		resolved = append(resolved, c.ResolvePath(arg))
	}
	return resolved
}

func (f flags) prepare_paths(c *command, target string, linkName string, linkFunc func(string, string) error) (string, string, func(string, string) error, error) {
	resolvedTarget := target
	resolvedLinkName := linkName
	resolvedLinkFunc := linkFunc

	if f.logical || f.physical {
		evaluatedTarget, err := filepath.EvalSymlinks(resolvedTarget)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", "", nil, err
		}
		if err == nil {
			resolvedTarget = evaluatedTarget
			if f.physical {
				resolvedLinkFunc = os.Link
			}
		}
	}

	if resolvedLinkName == "" {
		resolvedLinkName = filepath.Base(resolvedTarget)
	}
	if f.dirtgt != "" {
		resolvedLinkName = filepath.Join(f.dirtgt, filepath.Base(resolvedLinkName))
	}

	return resolvedTarget, resolvedLinkName, resolvedLinkFunc, nil
}

func (f flags) maybe_remove_existing(c *command, linkName string) error {
	_, err := os.Lstat(linkName)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if f.prompt && !f.force && !prompt_overwrite(c.Stdin, c.Stdout, linkName) {
		return errSkipLink
	}
	if f.force || f.prompt {
		return os.Remove(linkName)
	}
	return fmt.Errorf("target exists: %s", linkName)
}

func prompt_overwrite(stdin io.Reader, stdout io.Writer, fileName string) bool {
	fmt.Fprintf(stdout, "ln: overwrite '%s'? ", fileName)
	answer, err := bufio.NewReader(stdin).ReadString('\n')
	if err != nil || answer == "" {
		return false
	}
	return strings.ToLower(strings.TrimSpace(answer[:1])) == "y"
}
