package distro

import (
	"context"
	"io"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// InstallPkgs installs pkgs using the distro-appropriate package manager.
// For Debian families it runs apt-get update before installing.
func InstallPkgs(ctx context.Context, exe *executor.Executor, stdout io.Writer, family string, pkgs ...string) error {
	opts := executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout}
	switch family {
	case Debian:
		if err := exe.Run(ctx, opts, "apt-get", "update", "-q"); err != nil {
			return err
		}
		args := append([]string{"install", "-y", "--"}, pkgs...)
		return exe.Run(ctx, opts, "apt-get", args...)
	case Fedora:
		args := append([]string{"install", "-y", "--"}, pkgs...)
		return exe.Run(ctx, opts, "dnf", args...)
	default:
		args := append([]string{"-S", "--noconfirm", "--"}, pkgs...)
		return exe.Run(ctx, opts, "pacman", args...)
	}
}
