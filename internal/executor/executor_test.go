package executor_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

func TestDryRun(t *testing.T) {
	var buf bytes.Buffer
	exe := executor.New(&buf, &buf)
	exe.DryRun = true

	if err := exe.Run(context.Background(), executor.Options{}, "echo", "hello"); err != nil {
		t.Fatalf("DryRun must not error: %v", err)
	}
	if !strings.Contains(buf.String(), "dry-run") {
		t.Errorf("expected dry-run marker in output, got: %q", buf.String())
	}
}

func TestDryRunSudoPrepended(t *testing.T) {
	var buf bytes.Buffer
	exe := executor.New(&buf, &buf)
	exe.DryRun = true

	if err := exe.Run(context.Background(), executor.Options{RequiresSudo: true}, "apt-get", "update"); err != nil {
		t.Fatalf("DryRun must not error: %v", err)
	}
	if !strings.Contains(buf.String(), "sudo") {
		t.Errorf("expected 'sudo' in dry-run output, got: %q", buf.String())
	}
}

func TestRunEcho(t *testing.T) {
	var buf bytes.Buffer
	exe := executor.New(nil, nil)

	err := exe.Run(context.Background(), executor.Options{Stdout: &buf}, "echo", "lumina")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "lumina") {
		t.Errorf("expected 'lumina' in output, got: %q", buf.String())
	}
}

func TestRunFailure(t *testing.T) {
	exe := executor.New(nil, nil)
	err := exe.Run(context.Background(), executor.Options{}, "false")
	if err == nil {
		t.Fatal("expected error from 'false' command")
	}
}

func TestOutput(t *testing.T) {
	exe := executor.New(nil, nil)
	out, err := exe.Output(context.Background(), executor.Options{}, "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected 'hello' in output, got: %q", out)
	}
}

func TestOutputDryRun(t *testing.T) {
	exe := executor.New(nil, nil)
	exe.DryRun = true
	out, err := exe.Output(context.Background(), executor.Options{}, "whoami")
	if err != nil {
		t.Fatalf("DryRun Output must not error: %v", err)
	}
	if !strings.Contains(out, "dry-run") {
		t.Errorf("expected dry-run marker in output, got: %q", out)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	exe := executor.New(nil, nil)
	err := exe.Run(ctx, executor.Options{}, "sleep", "10")
	if err == nil {
		t.Fatal("expected error with cancelled context")
	}
}
