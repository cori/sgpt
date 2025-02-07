// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
)

func IsPipedShell() (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return false, nil
	}
	return true, nil
}

func ExecuteCommandWithConfirmation(ctx context.Context, input io.Reader, output io.Writer, command string) error {
	// Require user confirmation
	ok, err := getUserConfirmation(input, output)
	if err != nil {
		return err
	}
	if ok {
		return executeShellCommand(ctx, output, command)
	}
	return nil
}

func getUserConfirmation(input io.Reader, output io.Writer) (bool, error) {
	for {
		if _, err := fmt.Fprint(output, "Do you want to execute this command? (Y/n) "); err != nil {
			return false, err
		}
		reader := bufio.NewReader(input)
		char, _, err := reader.ReadRune()
		if err != nil {
			return false, err
		}
		if char == '\n' || char == '\r' || char == 'Y' || char == 'y' {
			slog.Debug("User confirmed")
			return true, nil
		} else if char == 'N' || char == 'n' {
			slog.Debug("User denied")
			return false, nil
		} else {
			slog.Debug("User entered unrecognised input for confirmation: " + string(char))
		}
	}
}

func executeShellCommand(ctx context.Context, output io.Writer, command string) error {
	var executeCommand string
	var args []string
	switch runtime.GOOS {
	case "windows":
		slog.Debug("Running on Windows - using cmd")
		executeCommand = "cmd"
		args = []string{"/C", command}
	default:
		slog.Debug("Running on Linux like OS - using bash")
		executeCommand = "bash"
		args = []string{"-c", command}
	}

	cmd := exec.CommandContext(ctx, executeCommand, args...)
	cmd.Stdout = output
	err := cmd.Run()
	if err != nil {
		return err
	}
	slog.Debug("Command executed successfully")
	return nil
}
