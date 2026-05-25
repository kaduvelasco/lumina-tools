package tui

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// funcCmd wraps a Go domain function as a tea.ExecCommand.
// Bubble Tea suspends the TUI, calls SetStdout/SetStderr with the real
// terminal writers, then calls Run — giving the function full terminal access.
type funcCmd struct {
	ctx    context.Context
	fn     func(ctx context.Context, exe *executor.Executor, stdout io.Writer) error
	stdout io.Writer
	stderr io.Writer
}

func newFuncCmd(
	ctx context.Context,
	fn func(context.Context, *executor.Executor, io.Writer) error,
) *funcCmd {
	return &funcCmd{ctx: ctx, fn: fn}
}

func (c *funcCmd) SetStdin(_ io.Reader) {}
func (c *funcCmd) SetStdout(w io.Writer) { c.stdout = w }
func (c *funcCmd) SetStderr(w io.Writer) { c.stderr = w }
func (c *funcCmd) Run() error {
	exe := executor.New(c.stdout, c.stderr)
	return c.fn(c.ctx, exe, c.stdout)
}

// interactiveFuncCmd is like funcCmd but also provides stdin to the function,
// enabling interactive sub-programs (e.g., multi-select forms) inside tea.Exec.
type interactiveFuncCmd struct {
	ctx    context.Context
	fn     func(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func newInteractiveFuncCmd(
	ctx context.Context,
	fn func(context.Context, *executor.Executor, io.Reader, io.Writer) error,
) *interactiveFuncCmd {
	return &interactiveFuncCmd{ctx: ctx, fn: fn}
}

func (c *interactiveFuncCmd) SetStdin(r io.Reader)  { c.stdin = r }
func (c *interactiveFuncCmd) SetStdout(w io.Writer) { c.stdout = w }
func (c *interactiveFuncCmd) SetStderr(w io.Writer) { c.stderr = w }
func (c *interactiveFuncCmd) Run() error {
	exe := executor.New(c.stdout, c.stderr)
	return c.fn(c.ctx, exe, c.stdin, c.stdout)
}
