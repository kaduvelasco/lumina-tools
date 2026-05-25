package repo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// ApplyIdent applies local git identity to the current repository.
func ApplyIdent(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Repositórios :: Aplicar Identidade")

	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		ui.Err(stdout, "Este diretório não é um repositório Git.")
		ui.WaitEnter(stdout)
		return fmt.Errorf("este diretorio nao e um repositorio Git")
	}

	if err := applyLocalIdentity(ctx, exe, stdout); err != nil {
		if errors.Is(err, errCancelled) {
			ui.Info(stdout, "Operação cancelada.")
		} else {
			ui.Err(stdout, err.Error())
		}
		ui.WaitEnter(stdout)
		return nil
	}
	ui.WaitEnter(stdout)
	return nil
}

// applyLocalIdentity applies local git identity in the current working directory.
func applyLocalIdentity(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	return applyLocalIdentityAt(ctx, exe, stdout, "")
}

// applyLocalIdentityAt applies local git identity in dir.
// When dir is empty, git operates on the current working directory.
func applyLocalIdentityAt(ctx context.Context, exe *executor.Executor, stdout io.Writer, dir string) error {
	gitCmd := func(args ...string) []string {
		if dir == "" {
			return args
		}
		out := make([]string, 0, 2+len(args))
		out = append(out, "-C", dir)
		out = append(out, args...)
		return out
	}

	curName, _ := exe.Output(ctx, executor.Options{}, "git", gitCmd("config", "--local", "user.name")...)
	curEmail, _ := exe.Output(ctx, executor.Options{}, "git", gitCmd("config", "--local", "user.email")...)
	if trim(curName) != "" {
		ui.Info(stdout, "Identidade atual: "+trim(curName)+" <"+trim(curEmail)+">")
	}

	defName, _ := exe.Output(ctx, executor.Options{}, "git", "config", "--global", "user.name")
	defEmail, _ := exe.Output(ctx, executor.Options{}, "git", "config", "--global", "user.email")

	fmt.Fprintf(stdout, "Nome [%s]: ", trim(defName))
	name := strings.TrimSpace(prompt.ReadLine())
	if name == "" {
		name = trim(defName)
	}

	fmt.Fprintf(stdout, "E-mail [%s]: ", trim(defEmail))
	email := strings.TrimSpace(prompt.ReadLine())
	if email == "" {
		email = trim(defEmail)
	}

	if name == "" || email == "" {
		return errCancelled
	}

	cred := resolveCredHelper(ctx, exe)

	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	for _, args := range [][]string{
		{"config", "--local", "user.name", name},
		{"config", "--local", "user.email", email},
		{"config", "--local", "credential.helper", cred},
	} {
		if err := exe.Run(ctx, opts, "git", gitCmd(args...)...); err != nil {
			return fmt.Errorf("git %v: %w", args, err)
		}
	}

	ui.Success(stdout, "Identidade aplicada: "+name+" <"+email+">")
	return nil
}
