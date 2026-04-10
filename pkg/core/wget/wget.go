// Copyright 2012-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package wget implements the wget core utility.
package wget

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/u-root/u-root/pkg/core"
	"github.com/u-root/u-root/pkg/curl"

	_ "github.com/u-root/cpuid"
)

var errEmptyURL = errors.New("empty url")

type command struct {
	core.Base
	url        string
	outputPath string
}

// New creates a new wget command.
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
	outPath, rawURL, err := parseFlags(args...)
	if err != nil {
		return err
	}
	c.outputPath = outPath
	c.url = rawURL

	if c.url == "" {
		return errEmptyURL
	}

	parsedURL, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	if c.outputPath == "" {
		c.outputPath = defaultOutputPath(parsedURL.Path)
	}

	schemes := curl.Schemes{
		"tftp":  curl.DefaultTFTPClient,
		"http":  curl.DefaultHTTPClient,
		"https": curl.DefaultHTTPClient,
		"file":  &curl.LocalFileClient{},
	}

	reader, err := schemes.FetchWithoutCache(ctx, parsedURL)
	if err != nil {
		return fmt.Errorf("failed to download %v: %w", c.url, err)
	}

	if c.outputPath == "-" {
		_, err = io.Copy(c.Stdout, reader)
		return err
	}

	file, err := os.Create(c.ResolvePath(c.outputPath))
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, reader)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func parseFlags(args ...string) (string, string, error) {
	if len(args) == 0 {
		return "", "", errEmptyURL
	}

	fs := flag.NewFlagSet("wget", flag.ContinueOnError)
	outPath := fs.String("O", "", "output file")
	if err := fs.Parse(args); err != nil {
		return "", "", err
	}
	if len(fs.Args()) == 0 {
		return "", "", errEmptyURL
	}
	urlArg := fs.Args()[0]
	if err := fs.Parse(fs.Args()[1:]); err != nil {
		return "", "", err
	}
	return *outPath, urlArg, nil
}

func defaultOutputPath(urlPath string) string {
	if urlPath == "" || strings.HasSuffix(urlPath, "/") {
		return "index.html"
	}
	return path.Base(urlPath)
}
