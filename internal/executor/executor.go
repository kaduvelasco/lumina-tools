package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Options configures a single command invocation.
type Options struct {
	RequiresSudo bool
	Stdin        io.Reader
	Stdout       io.Writer
	Stderr       io.Writer
	Env          []string // extra KEY=VALUE pairs appended to os.Environ()
}

// Executor is the single point through which all external commands run.
// Set DryRun to true to print commands without executing them.
type Executor struct {
	DryRun bool
	Stdout io.Writer
	Stderr io.Writer
}

// New returns an Executor that writes to the provided writers by default.
func New(stdout, stderr io.Writer) *Executor {
	return &Executor{Stdout: stdout, Stderr: stderr}
}

// Run executes name with args, escalating to sudo when opts.RequiresSudo is true.
// Output writers in opts take precedence over the Executor defaults.
func (e *Executor) Run(ctx context.Context, opts Options, name string, args ...string) error {
	stdout := firstWriter(opts.Stdout, e.Stdout, io.Discard)
	stderr := firstWriter(opts.Stderr, e.Stderr, io.Discard)

	cmdName, cmdArgs := buildCmd(opts.RequiresSudo, name, args)

	if e.DryRun {
		fmt.Fprintf(stdout, "[dry-run] %s %v\n", cmdName, cmdArgs)
		return nil
	}

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stdin = opts.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmdName, err)
	}
	return nil
}

// Output runs name with args and returns the combined stdout as a string.
// Stderr is forwarded to the Executor's Stderr writer.
func (e *Executor) Output(ctx context.Context, opts Options, name string, args ...string) (string, error) {
	stderr := firstWriter(opts.Stderr, e.Stderr, io.Discard)

	cmdName, cmdArgs := buildCmd(opts.RequiresSudo, name, args)

	if e.DryRun {
		return fmt.Sprintf("[dry-run] %s %v", cmdName, cmdArgs), nil
	}

	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
	cmd.Stdin = opts.Stdin
	cmd.Stderr = stderr
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", cmdName, err)
	}
	return string(out), nil
}

func buildCmd(sudo bool, name string, args []string) (string, []string) {
	if sudo {
		return "sudo", append([]string{name}, args...)
	}
	return name, args
}

func firstWriter(writers ...io.Writer) io.Writer {
	for _, w := range writers {
		if w != nil {
			return w
		}
	}
	return io.Discard
}
