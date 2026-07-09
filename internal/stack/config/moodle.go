package stackconfig

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaduvelasco/lumina-tools/internal/config"
	"github.com/kaduvelasco/lumina-tools/internal/executor"
	"github.com/kaduvelasco/lumina-tools/internal/prompt"
	"github.com/kaduvelasco/lumina-tools/internal/ui"
)

// hasPHPAtLeast reports whether any version in versions is >= minSuffix once
// formatted like phpSuffix (e.g. "8.2" -> 82).
func hasPHPAtLeast(versions []string, minSuffix int) bool {
	for _, v := range versions {
		n, err := strconv.Atoi(phpSuffix(v))
		if err == nil && n >= minSuffix {
			return true
		}
	}
	return false
}

// moodleURLPrefix returns the URL path (relative to the nginx webroot) that
// corresponds to moodleDir, and false if moodleDir is not inside
// <workspace>/www/html — the only tree nginx actually serves.
func moodleURLPrefix(workspace, moodleDir string) (string, bool) {
	htmlRoot := filepath.Join(workspace, "www", "html")
	rel, err := filepath.Rel(htmlRoot, moodleDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if rel == "." {
		return "", true
	}
	return filepath.ToSlash(rel), true
}

// moodleURLPath builds the nginx location path for a Moodle install folder,
// e.g. urlPrefix="mdle", folder="dev-501" -> "/mdle/dev-501/".
func moodleURLPath(urlPrefix, folder string) string {
	prefix := strings.Trim(urlPrefix, "/")
	if prefix == "" {
		return "/" + folder + "/"
	}
	return "/" + prefix + "/" + folder + "/"
}

// buildMoodleLocations generates one ^~ location block per marked Moodle
// install (Moodle 5.1+ router). A plain prefix location would silently lose
// to the generic `location ~ \.php$` regex below it — nginx always prefers a
// matching regex location over a normal prefix one — so ^~ is required to
// force nginx to stop at this block instead of falling through. phpRegex and
// dispatch let the caller plug in either the fixed-container dispatch or the
// subdomain-based one used by the two nginx server blocks.
//
// The rewrite maps every request under the install's URL prefix into its
// public/ subtree *before* try_files ever touches disk. Without it, $uri
// would resolve relative to the shared global root (the install's legacy
// root, not public/) — meaning any file that happens to exist outside
// public/ (config.php, composer.json, version.php, ...) would be served
// as a static file straight off disk, bypassing the router entirely. That's
// both a real information-disclosure bug (config.php source, credentials
// included) and exactly what Moodle 5.1+'s own installer/upgrade check
// ("rootdirpublic") detects and refuses to proceed past. Routing everything
// through public/ first closes both problems at once.
//
// try_files intentionally omits the `$uri/` directory-match alternative
// (unlike the plain PHP location below it): the install's public/ directory
// always exists on disk, so `$uri/` would match it and hand the request to
// nginx's `index` directive, which resolves to public/index.php directly
// instead of going through the router. Dropping `$uri/` forces every request
// that isn't a real file under public/ straight into r.php, matching
// Moodle's own documented nginx recipe.
func buildMoodleLocations(installs []string, urlPrefix, phpRegex, dispatch string) string {
	if len(installs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, name := range installs {
		path := moodleURLPath(urlPrefix, name)
		publicPath := path + "public/"
		fmt.Fprintf(&b, `
    location ^~ %s {
        rewrite ^%s(.*)$ %s$1 break;
        try_files $uri %sr.php$is_args$args;

        location ~ %s {
            %s
        }
    }
`, path, regexp.QuoteMeta(path), publicPath, publicPath, phpRegex, dispatch)
	}
	return b.String()
}

const moodleDefaultDispatch = `include fastcgi_params;
            fastcgi_pass {{MOODLE_PHP}}:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;`

const moodleVersionedDispatch = `fastcgi_split_path_info ^(.+\.php)(/.+)$;
            include fastcgi_params;
            if ($p_ver = "") { set $p_ver {{MOODLE_PHP_VER}}; }
            set $php_upstream php$p_ver:9000;
            fastcgi_pass $php_upstream;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            fastcgi_param PATH_INFO $fastcgi_path_info;`

// moodleDefaultPHP returns the lowest selected PHP version that satisfies the
// Moodle 5.1+ router (>= 8.2), as both the container name ("php82") and the
// bare suffix ("82") — used as the fallback for Moodle location blocks when
// $p_ver is empty, decoupled from the legacy default PHP version so an older
// version picked as versions[0] never breaks the Moodle router. Falls back to
// versions[0] if none qualifies, which should not happen in practice since
// callers only offer the Moodle prompt when hasPHPAtLeast(versions, 82) is
// already true.
func moodleDefaultPHP(versions []string) (container, suffix string) {
	best := 0
	for _, v := range versions {
		n, err := strconv.Atoi(phpSuffix(v))
		if err != nil || n < 82 {
			continue
		}
		if best == 0 || n < best {
			best = n
		}
	}
	if best == 0 {
		return "php" + phpSuffix(versions[0]), phpSuffix(versions[0])
	}
	suffix = strconv.Itoa(best)
	return "php" + suffix, suffix
}

// buildNginxConf renders the full nginx/default.conf content for the given
// PHP versions and marked Moodle installs. Pure — no I/O.
func buildNginxConf(versions []string, moodleInstalls []string, urlPrefix string) string {
	defaultPHP := "php" + phpSuffix(versions[0])
	defaultVer := phpSuffix(versions[0])

	moodleDefaultBlocks := buildMoodleLocations(moodleInstalls, urlPrefix, `\.php$`, moodleDefaultDispatch)
	moodleVersionedBlocks := buildMoodleLocations(moodleInstalls, urlPrefix, `[^/]\.php(/|$)`, moodleVersionedDispatch)

	nginxConf := nginxConfTpl
	nginxConf = strings.ReplaceAll(nginxConf, "{{MOODLE_LOCATIONS_DEFAULT}}", moodleDefaultBlocks)
	nginxConf = strings.ReplaceAll(nginxConf, "{{MOODLE_LOCATIONS_VERSIONED}}", moodleVersionedBlocks)
	nginxConf = strings.ReplaceAll(nginxConf, "{{DEFAULT_PHP}}", defaultPHP)
	nginxConf = strings.ReplaceAll(nginxConf, "{{DEFAULT_PHP_VER}}", defaultVer)
	if len(moodleInstalls) > 0 {
		moodlePHP, moodleVer := moodleDefaultPHP(versions)
		nginxConf = strings.ReplaceAll(nginxConf, "{{MOODLE_PHP}}", moodlePHP)
		nginxConf = strings.ReplaceAll(nginxConf, "{{MOODLE_PHP_VER}}", moodleVer)
	}
	return nginxConf
}

// promptMoodleInstalls asks for the Moodle install base directory (defaulting
// to previousDir, or <workspace>/www/html/mdle when empty), lists its
// subfolders and lets the user mark which ones are Moodle 5.1+ installs —
// folders present in previousInstalls come pre-selected. Shared by Compose
// (initial stack generation) and MoodleRouter (reconfiguring an existing
// stack without regenerating it).
func promptMoodleInstalls(ctx context.Context, stdin io.Reader, stdout io.Writer, workspace, previousDir, previousInstalls string) (dir, urlPrefix string, installs []string, err error) {
	defaultMoodleDir := previousDir
	if defaultMoodleDir == "" {
		defaultMoodleDir = filepath.Join(workspace, "www", "html", "mdle")
	}
	fmt.Fprintf(stdout, "\nDiretório de instalações do Moodle [%s]: ", defaultMoodleDir)
	if input := strings.TrimSpace(prompt.ReadLine()); input != "" {
		expanded, expErr := config.ExpandPath(input)
		if expErr != nil {
			return "", "", nil, fmt.Errorf("expandir diretorio moodle: %w", expErr)
		}
		defaultMoodleDir = expanded
	}
	dir = defaultMoodleDir

	prefix, ok := moodleURLPrefix(workspace, dir)
	if !ok {
		ui.Warning(stdout, "O diretório precisa estar dentro de "+filepath.Join(workspace, "www", "html")+" (raiz servida pelo nginx) — pulando configuração do roteador Moodle.")
		return dir, "", nil, nil
	}
	urlPrefix = prefix

	var folders []string
	if entries, readErr := os.ReadDir(dir); readErr == nil {
		for _, e := range entries {
			if e.IsDir() {
				folders = append(folders, e.Name())
			}
		}
	}
	if len(folders) == 0 {
		ui.Info(stdout, "Nenhuma pasta encontrada em "+dir+" — pulando configuração do roteador Moodle.")
		return dir, urlPrefix, nil, nil
	}

	previously := make(map[string]bool)
	for _, name := range strings.Fields(previousInstalls) {
		previously[name] = true
	}
	ui.Info(stdout, "Marque as pastas que são instalações Moodle 5.1+ (requerem o roteador r.php):")
	moodleItems := make([]ui.SelectItem, len(folders))
	for i, name := range folders {
		moodleItems[i] = ui.SelectItem{Label: name, ID: name, Selected: previously[name]}
	}
	finalMoodle, moodleConfirmed, msErr := ui.RunMultiSelect(ctx, stdin, stdout, moodleItems)
	if msErr != nil {
		return dir, urlPrefix, nil, msErr
	}
	if moodleConfirmed {
		for _, item := range finalMoodle {
			if item.Selected {
				installs = append(installs, item.ID)
			}
		}
	}
	return dir, urlPrefix, installs, nil
}

// nginxContainerRunning reports whether the "nginx" container from the
// generated stack is currently running.
func nginxContainerRunning(ctx context.Context, exe *executor.Executor) bool {
	out, err := exe.Output(ctx, executor.Options{}, "docker", "inspect", "-f", "{{.State.Running}}", "nginx")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "true"
}

// MoodleRouter reconfigures which Moodle 5.1+ installs get router (r.php)
// location blocks in the already-generated nginx/default.conf, without
// touching docker-compose.yml, .env, the PHP Dockerfile or php.ini. Meant
// for adding/removing Moodle installs after the stack is already running —
// "Criar Stack PHP" would otherwise have to be re-run just to change nginx.
func MoodleRouter(ctx context.Context, exe *executor.Executor, stdin io.Reader, stdout io.Writer) error {
	ui.PrintHeader(stdout, "Configurar Roteamento Moodle")

	cfg, err := config.Load()
	if err != nil {
		ui.Err(stdout, "Falha ao carregar config: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("carregar config: %w", err)
	}

	if cfg.DockerComposeDir == "" {
		ui.Warning(stdout, "Stack não configurada. Execute 'Criar Stack PHP' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}

	nginxPath := filepath.Join(cfg.DockerComposeDir, "nginx", "default.conf")
	if _, statErr := os.Stat(nginxPath); statErr != nil {
		ui.Warning(stdout, "Stack não configurada. Execute 'Criar Stack PHP' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}

	versions := strings.Fields(cfg.Stack.PHPVersions)
	if len(versions) == 0 {
		ui.Warning(stdout, "Stack não configurada. Execute 'Criar Stack PHP' primeiro.")
		ui.WaitEnter(stdout)
		return nil
	}

	if !hasPHPAtLeast(versions, 82) {
		ui.Info(stdout, "Moodle 5.1+ requer PHP 8.2 ou superior; a stack atual não tem nenhuma versão compatível.")
		ui.WaitEnter(stdout)
		return nil
	}

	dir, urlPrefix, installs, err := promptMoodleInstalls(ctx, stdin, stdout, cfg.WorkspacePath, cfg.Stack.MoodleDir, cfg.Stack.MoodleInstalls)
	if err != nil {
		return err
	}

	oldContent, err := os.ReadFile(nginxPath)
	if err != nil {
		ui.Err(stdout, "Falha ao ler nginx/default.conf atual: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("ler nginx atual: %w", err)
	}

	newContent := buildNginxConf(versions, installs, urlPrefix)
	if err := os.WriteFile(nginxPath, []byte(newContent), 0o644); err != nil {
		ui.Err(stdout, "Falha ao escrever nginx/default.conf: "+err.Error())
		ui.WaitEnter(stdout)
		return fmt.Errorf("escrever nginx.conf: %w", err)
	}

	if !nginxContainerRunning(ctx, exe) {
		ui.Info(stdout, "nginx/default.conf atualizado em disco. Inicie a stack para aplicar.")
	} else {
		var out bytes.Buffer
		testErr := exe.Run(ctx, executor.Options{Stdout: &out, Stderr: &out}, "docker", "exec", "nginx", "nginx", "-t")
		if testErr != nil {
			if restoreErr := os.WriteFile(nginxPath, oldContent, 0o644); restoreErr != nil {
				ui.Err(stdout, "Configuração inválida e falha ao reverter: "+restoreErr.Error()+"\n\n"+out.String())
				ui.WaitEnter(stdout)
				return fmt.Errorf("reverter nginx.conf: %w", restoreErr)
			}
			ui.Err(stdout, "Configuração inválida, alterações revertidas:\n\n"+out.String())
			ui.WaitEnter(stdout)
			return fmt.Errorf("configuracao nginx invalida: %w", testErr)
		}

		if err := exe.Run(ctx, executor.Options{Stdout: stdout, Stderr: stdout}, "docker", "exec", "nginx", "nginx", "-s", "reload"); err != nil {
			ui.Err(stdout, "Falha ao recarregar o nginx: "+err.Error())
			ui.WaitEnter(stdout)
			return fmt.Errorf("recarregar nginx: %w", err)
		}
		ui.Success(stdout, "Roteamento atualizado e aplicado sem downtime.")
	}

	cfg.Stack.MoodleDir = dir
	cfg.Stack.MoodleInstalls = strings.Join(installs, " ")
	if err := config.Save(cfg); err != nil {
		ui.Warning(stdout, "Falha ao salvar configurações: "+err.Error())
	}

	if len(installs) > 0 {
		ui.Info(stdout, "Instalações roteadas: "+strings.Join(installs, ", "))
	} else {
		ui.Info(stdout, "Nenhuma instalação marcada — roteador removido de todas as pastas.")
	}
	ui.WaitEnter(stdout)
	return nil
}
