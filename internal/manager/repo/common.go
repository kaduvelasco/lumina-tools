package repo

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
)

// errCancelled is returned when the user leaves a required field empty,
// signalling that the current operation should be aborted gracefully.
var errCancelled = errors.New("operacao cancelada")

// libsecret binary locations across distros.
var libsecretPaths = []string{
	"/usr/share/doc/git/contrib/credential/libsecret/git-credential-libsecret",
	"/usr/lib/git-core/git-credential-libsecret",
	"/usr/libexec/git-core/git-credential-libsecret",
	"/usr/lib/git/git-credential-libsecret",
}

func resolveCredHelper(ctx context.Context, exe *executor.Executor) string {
	for _, p := range libsecretPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if _, err := exe.Output(ctx, executor.Options{}, "which", "git-credential-libsecret"); err == nil {
		return "libsecret"
	}
	return "cache"
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

func cwd() string {
	d, _ := os.Getwd()
	return d
}
