// +build windows
// Copyright 2013, Örjan Persson. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logging

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"syscall"
)

var (
	kernel32DLL    = syscall.NewLazyDLL("kernel32.dll")
	setConsoleMode = kernel32DLL.NewProc("SetConsoleMode")
	kernelSetup    = false
)

const EnableVirtualTerminalProcessing uint32 = 0x4

// Character attributes
// Note:
// -- The attributes are combined to produce various colors (e.g., Blue + Green will create Cyan).
//    Clearing all foreground or background colors results in black; setting all creates white.
// See https://msdn.microsoft.com/en-us/library/windows/desktop/ms682088(v=vs.85).aspx#_win32_character_attributes.
type color int

const (
	ColorBlack color = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

var (
	colors = []string{
		CRITICAL: ColorSeq(ColorMagenta),
		ERROR:    ColorSeq(ColorRed),
		WARNING:  ColorSeq(ColorYellow),
		NOTICE:   ColorSeq(ColorGreen),
		DEBUG:    ColorSeq(ColorCyan),
	}
	boldcolors = []string{
		CRITICAL: ColorSeqBold(ColorMagenta),
		ERROR:    ColorSeqBold(ColorRed),
		WARNING:  ColorSeqBold(ColorYellow),
		NOTICE:   ColorSeqBold(ColorGreen),
		DEBUG:    ColorSeqBold(ColorCyan),
	}
)

// LogBackend utilizes the standard log module.
type LogBackend struct {
	Logger *log.Logger
	Color  bool

	ColorConfig []string
}

// NewLogBackend creates a new LogBackend.
func NewLogBackend(out io.Writer, prefix string, flag int) *LogBackend {
	if !kernelSetup {
		kernelSetup = true
		var mode uint32
		err := syscall.GetConsoleMode(syscall.Stdout, &mode)
		if err != nil {
			panic(err) // TODO: check if panic is wanted
		}
		mode |= EnableVirtualTerminalProcessing

		ret, _, err := setConsoleMode.Call(uintptr(syscall.Stdout), uintptr(mode))
		if ret == 0 {
			panic(err)
		}
	}

	return &LogBackend{Logger: log.New(out, prefix, flag)}
}

// Log implements the Backend interface.
func (b *LogBackend) Log(level Level, calldepth int, rec *Record) error {
	if b.Color {
		col := colors[level]
		if len(b.ColorConfig) > int(level) && b.ColorConfig[level] != "" {
			col = b.ColorConfig[level]
		}

		buf := &bytes.Buffer{}
		buf.Write([]byte(col))
		buf.Write([]byte(rec.Formatted(calldepth + 1)))
		buf.Write([]byte("\x1b[0m"))
		// For some reason, the Go logger arbitrarily decided "2" was the correct
		// call depth...
		return b.Logger.Output(calldepth+2, buf.String())
	}

	return b.Logger.Output(calldepth+2, rec.Formatted(calldepth+1))
}

// ConvertColors takes a list of ints representing colors for log levels and
// converts them into strings for ANSI color formatting
func ConvertColors(colors []int, bold bool) []string {
	converted := []string{}
	for _, i := range colors {
		if bold {
			converted = append(converted, ColorSeqBold(color(i)))
		} else {
			converted = append(converted, ColorSeq(color(i)))
		}
	}

	return converted
}

func ColorSeq(color color) string {
	return fmt.Sprintf("\x1b[%dm", int(color))
}

func ColorSeqBold(color color) string {
	return fmt.Sprintf("\x1b[1;%dm", int(color))
}

func doFmtVerbLevelColor(layout string, level Level, output io.Writer) {
	if layout == "bold" {
		output.Write([]byte(boldcolors[level]))
	} else if layout == "reset" {
		output.Write([]byte("\x1b[0m"))
	} else {
		output.Write([]byte(colors[level]))
	}
}
