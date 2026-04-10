// Copyright 2012-2022 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tail implements the tail core utility.
package tail

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"syscall"
	"time"

	"github.com/u-root/u-root/pkg/core"
)

type readAtSeeker interface {
	io.ReaderAt
	io.Seeker
}

type tailConfig struct {
	follow         bool
	numLines       int
	followDuration time.Duration
}

type command struct {
	core.Base
}

// New creates a new tail command.
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

	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	follow := fs.Bool("f", false, "follow the end of the file")
	numLines := fs.Int("n", 10, "number of lines to show")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return c.run(c.Stdin, c.Stdout, *follow, *numLines, 500*time.Millisecond, fs.Args())
}

func getBlockSize(numLines int) int64 {
	return 81 * int64(numLines)
}

func lastNLines(buf []byte, n int) []byte {
	slice := buf
	var data []byte
	if len(slice) != 0 {
		if slice[len(slice)-1] == '\n' {
			slice = slice[:len(slice)-1]
		}
		var foundLines int
		var idx int
		for foundLines < n {
			idx = bytes.LastIndexByte(slice, '\n')
			if idx == -1 {
				break
			}
			foundLines++
			if len(slice) > 1 && slice[idx-1] == '\n' {
				slice = slice[:idx]
			} else {
				slice = slice[:idx-1]
			}
		}
		if idx == -1 {
			data = buf
		} else {
			data = buf[idx+1:]
		}
	}
	return data
}

func readLastLinesBackwards(input readAtSeeker, writer io.Writer, numLines int) error {
	blkSize := getBlockSize(numLines)
	lastPos, err := input.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	readData := make([]byte, 0)
	buf := make([]byte, blkSize)
	pos := lastPos
	var foundLines int
	for pos != 0 {
		thisChunkSize := min(pos, blkSize)
		pos -= thisChunkSize
		n, err := input.ReadAt(buf, pos)
		if err != nil && err != io.EOF {
			return err
		}
		n = min(n, int(thisChunkSize))
		readData = slices.Concat(buf[:n], readData)
		foundLines += bytes.Count(buf[:n], []byte{'\n'})
		if foundLines >= numLines {
			break
		}
	}
	data := lastNLines(readData, numLines)
	if _, err = writer.Write(data); err != nil {
		return err
	}
	_, err = input.Seek(lastPos, io.SeekStart)
	return err
}

func readLastLinesFromBeginning(input io.ReadSeeker, writer io.Writer, numLines int) error {
	blkSize := getBlockSize(numLines)
	buf := make([]byte, blkSize)
	var slice []byte
	for {
		n, err := io.ReadFull(input, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			if err != io.ErrUnexpectedEOF {
				return err
			}
		}
		slice = append(slice, buf[:n]...)
		slice = lastNLines(slice, numLines)
	}
	_, err := writer.Write(slice)
	return err
}

func isTruncated(file *os.File) (bool, error) {
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fileInfo.Size(), nil
}

func tailFile(inFile *os.File, writer io.Writer, config tailConfig) error {
	retryFromBeginning := false
	err := readLastLinesBackwards(inFile, writer, config.numLines)
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok && pathErr.Err == syscall.ESPIPE {
			retryFromBeginning = true
		} else {
			return err
		}
	}
	if retryFromBeginning {
		if err = readLastLinesFromBeginning(inFile, writer, config.numLines); err != nil {
			return err
		}
	}
	if config.follow {
		blkSize := getBlockSize(1)
		buf := make([]byte, blkSize)
		for {
			n, err := inFile.Read(buf)
			if err == io.EOF {
				time.Sleep(config.followDuration)
				truncated, errTruncated := isTruncated(inFile)
				if errTruncated != nil {
					break
				}
				if truncated {
					if _, errSeekStart := inFile.Seek(0, io.SeekStart); errSeekStart != nil {
						break
					}
				}
				continue
			}
			if err == nil {
				if _, err := writer.Write(buf[:n]); err != nil {
					return err
				}
				continue
			}
			break
		}
	}
	return nil
}

func (c *command) run(reader io.Reader, writer io.Writer, follow bool, numLines int, followDuration time.Duration, args []string) error {
	if numLines < 0 {
		numLines = -1 * numLines
	}
	config := tailConfig{follow: follow, numLines: numLines, followDuration: followDuration}

	if len(args) == 0 {
		file, ok := reader.(*os.File)
		if !ok {
			return fmt.Errorf("tail: stdin is not a file")
		}
		return tailFile(file, writer, config)
	}
	if len(args) > 1 && follow {
		return fmt.Errorf("tail: can only read one file at a time if follow true")
	}

	for i, fileArg := range args {
		if len(args) > 1 {
			fmt.Fprintf(writer, "==> %s <==\n", fileArg)
		}
		inFile, err := os.Open(c.ResolvePath(fileArg))
		if err != nil {
			return err
		}
		err = tailFile(inFile, writer, config)
		closeErr := inFile.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if len(args) > 1 && i < len(args)-1 {
			fmt.Fprintln(writer)
		}
	}
	return nil
}
