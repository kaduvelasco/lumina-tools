package stackconfig

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// Workspace creates the dev workspace directory structure.
func Workspace(ctx context.Context, exe *executor.Executor, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Criar Workspace")

	fmt.Fprint(stdout, "Local do workspace [/srv/workspace]: ")
	path := strings.TrimSpace(prompt.ReadLine())
	if path == "" {
		path = "/srv/workspace"
	}

	expanded, err := config.ExpandPath(path)
	if err != nil {
		ui.Err(stdout, err.Error())
		ui.WaitEnter(stdout)
		return err
	}
	path = expanded

	if _, err := os.Stat(filepath.Join(path, "www")); err == nil {
		ui.Warning(stdout, "Workspace já existe em: "+path+"\nContinuar irá reinstalar os arquivos de interface e reajustar permissões.")
		fmt.Fprint(stdout, "Continuar? (s/N): ")
		confirm := strings.TrimSpace(prompt.ReadLine())
		if confirm != "s" && confirm != "S" {
			ui.Info(stdout, "Operação cancelada.")
			ui.WaitEnter(stdout)
			return nil
		}
	}

	ui.Info(stdout, "Criando estrutura em: "+path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := exe.Run(ctx,
			executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
			"mkdir", "-p", path,
		); err != nil {
			ui.Err(stdout, "Falha ao criar diretório base: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("criar diretorio base: %w", err)
		}
		user := executor.CurrentUser()
		if user != "" {
			_ = exe.Run(ctx,
				executor.Options{RequiresSudo: true, Stdout: stdout, Stderr: stdout},
				"chown", "-R", user+":"+user, path,
			)
		}
	}

	dirs := []string{
		filepath.Join(path, "www", "html"),
		filepath.Join(path, "www", "data"),
		filepath.Join(path, "databases", "mariadb"),
		filepath.Join(path, "logs", "nginx"),
		filepath.Join(path, "docker", "nginx"),
		filepath.Join(path, "docker", "php"),
		filepath.Join(path, "docker", "php-config"),
		filepath.Join(path, "docker", "mariadb", "conf.d"),
		filepath.Join(path, "docker", "mariadb", "init"),
		filepath.Join(path, "backups"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			ui.Err(stdout, "Falha ao criar "+d+": "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("criar %s: %w", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(path, "www", "html", "index.php"), []byte(indexPHP), 0o644); err != nil {
		ui.Err(stdout, "Falha ao escrever index.php: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("escrever index.php: %w", err)
	}
	if err := os.WriteFile(filepath.Join(path, "www", "html", "info.php"), []byte(infoPHP), 0o644); err != nil {
		ui.Err(stdout, "Falha ao escrever info.php: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("escrever info.php: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}
	cfg.WorkspacePath = path
	cfg.DockerComposeDir = filepath.Join(path, "docker")
	if err := config.Save(cfg); err != nil {
		ui.Err(stdout, "Falha ao salvar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("salvar config: %w", err)
	}

	ui.Success(stdout, "Workspace criado com sucesso em: "+path)
	ui.Info(stdout, "Projetos PHP : "+filepath.Join(path, "www", "html")+
		"\nBackups SQL  : "+filepath.Join(path, "backups")+
		"\nConfig salva : ~/.lumina/config.yaml")
	ui.WaitEnter(stdout)
	return nil
}

var indexPHP = `<?php
// LuminaStack - Dashboard
$containers = ['nginx', 'mariadb'];
?><!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"><title>LuminaStack</title></head>
<body>
<h1>LuminaStack</h1>
<p>PHP <?= PHP_VERSION ?> | <?= date('d/m/Y H:i') ?></p>
</body>
</html>
`

var infoPHP = `<?php phpinfo();
`
