package db

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
)

// errCancelled is returned when the user deliberately leaves a required field
// empty, signalling that the current operation should be aborted gracefully.
var errCancelled = errors.New("operacao cancelada")

func requireContainer(ctx context.Context, exe *executor.Executor, name string) error {
	out, err := exe.Output(ctx, executor.Options{},
		"docker", "ps", "--filter", "name="+name, "--format", "{{.Names}}")
	if err != nil {
		return fmt.Errorf("container '%s' nao encontrado. Inicie a stack primeiro.", name)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == name {
			return nil
		}
	}
	return fmt.Errorf("container '%s' nao encontrado. Inicie a stack primeiro.", name)
}

func promptCredentials(stdout io.Writer) (user, pass string, err error) {
	fmt.Fprint(stdout, "   Usuário MariaDB (Enter para cancelar): ")
	user = strings.TrimSpace(prompt.ReadLine())
	if user == "" {
		return "", "", errCancelled
	}
	pass, err = prompt.ReadPassword(stdout, "   Senha MariaDB: ")
	if err != nil {
		return "", "", err
	}
	return user, pass, nil
}

func ensureDirExists(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// shellQuote returns a shell single-quoted string safe for bash -c invocations.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// writeTempSecret writes content to a new temporary file with 0600 permissions.
// Returns the file path and a cleanup function that removes the file.
func writeTempSecret(content, pattern string) (string, func(), error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", nil, err
	}
	_ = os.Chmod(f.Name(), 0o600)
	if _, err := fmt.Fprint(f, content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("credencial: %w", err)
	}
	f.Close()
	name := f.Name()
	return name, func() { os.Remove(name) }, nil
}

