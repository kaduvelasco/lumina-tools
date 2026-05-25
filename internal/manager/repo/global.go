package repo

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// ConfigureGlobal sets global git user.name, user.email and credential.helper.
func ConfigureGlobal(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Repositórios :: Identidade Global")

	curName, _ := exe.Output(ctx, executor.Options{}, "git", "config", "--global", "user.name")
	curEmail, _ := exe.Output(ctx, executor.Options{}, "git", "config", "--global", "user.email")
	ui.Info(stdout, "Atual: "+trim(curName)+" <"+trim(curEmail)+">")

	fmt.Fprint(stdout, "\nNome global (Enter para cancelar): ")
	name := strings.TrimSpace(prompt.ReadLine())
	if name == "" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	fmt.Fprint(stdout, "E-mail global (Enter para cancelar): ")
	email := strings.TrimSpace(prompt.ReadLine())
	if email == "" {
		ui.Info(stdout, "Operação cancelada.")
		ui.WaitEnter(stdout)
		return nil
	}

	cred := resolveCredHelper(ctx, exe)

	opts := executor.Options{Stdout: stdout, Stderr: stdout}
	for _, args := range [][]string{
		{"config", "--global", "user.name", name},
		{"config", "--global", "user.email", email},
		{"config", "--global", "credential.helper", cred},
	} {
		if err := exe.Run(ctx, opts, "git", args...); err != nil {
			ui.Err(stdout, "Falha: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("git %v: %w", args, err)
		}
	}

	ui.Success(stdout, "Identidade global configurada: "+name+" <"+email+">")
	ui.Info(stdout, "Dica: use seu Token (PAT) como senha no primeiro push.")
	ui.WaitEnter(stdout)
	return nil
}
