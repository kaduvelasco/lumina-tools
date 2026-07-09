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
// The outer rewrite maps every request under the install's URL prefix into
// its public/ subtree *before* try_files ever touches disk. Without it, $uri
// would resolve relative to the shared global root (the install's legacy
// root, not public/) — meaning any file that happens to exist outside
// public/ (config.php, composer.json, version.php, ...) would be served
// as a static file straight off disk, bypassing the router entirely. That's
// both a real information-disclosure bug (config.php source, credentials
// included) and exactly what Moodle 5.1+'s own installer/upgrade check
// ("rootdirpublic") detects and refuses to proceed past. Routing everything
// through public/ first closes both problems at once.
//
// That outer rewrite never runs for a request whose URI already ends in
// .php: nginx resolves the nested `location ~ <phpRegex>` directly against
// the *original* URI (it's the more specific match), bypassing the parent
// block's rewrite/try_files entirely — confirmed empirically against a real
// nginx+php-fpm pair, not just inferred from the docs. Since every real
// Moodle entry point (install.php, index.php, course/view.php, ...) is a
// genuine file requested with a .php URI, that path was completely unrouted:
// $uri stayed e.g. "/mdle/dev-501/install.php", SCRIPT_FILENAME pointed at a
// path that only exists one level down under public/, and PHP-FPM answered
// "File not found." The dispatch templates below carry their own copy of the
// same rewrite (via the {{NESTED_REWRITE}} placeholder, filled in here) so
// the nested location normalizes the path itself instead of relying on the
// parent ever being reached. The negative lookahead (?!public/) makes it a
// no-op when the URI already has /public/ in it — necessary because a
// request can also reach this nested location by falling through the outer
// rewrite+try_files first (e.g. non-.php request that hits it after a
// previous MOODLE_LOCATIONS pass), and rewriting twice would double the
// segment into ".../public/public/...".
//
// try_files includes the `$uri/` directory-match alternative, matching
// Moodle's own official multi-site nginx recipe. This was dropped in
// [2.2.8] over a since-disproven security concern: the worry was that
// `$uri/` matching an existing directory (the install's public/ itself,
// or any subdirectory under it, e.g. public/my/) would hand the request to
// nginx's `index` directive and serve that directory's index.php directly,
// "instead of going through the router". But by the time try_files runs
// here, the outer rewrite has already confined $uri to the public/ subtree
// — there is no directory `$uri/` could ever match that lives outside
// public/, so serving its index.php directly is exactly as safe as routing
// through r.php first; both ultimately execute the same file.
// Dropping `$uri/` had a real, user-facing regression instead: Moodle's
// router (public/lib/classes/router.php) is a Slim Framework app with an
// explicit route table — it does NOT implicitly resolve arbitrary
// directory paths to their index.php the way a plain web server does.
// Legacy-style pages Moodle still ships as real files (my/index.php,
// course/index.php, admin/index.php, ...) have no matching Slim route, so
// forcing every directory-style request through r.php made them all 404
// with Slim's generic "resource could not be found" page — confirmed
// empirically against a real Moodle 5.1 install (accessing /my/ after a
// completed installation). Restoring `$uri/` lets nginx resolve these the
// normal way, and r.php is only reached for paths matching no real file or
// directory at all (Slim's own registered routes, clean/pretty URLs).
func buildMoodleLocations(installs []string, urlPrefix, phpRegex, dispatch string) string {
	if len(installs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, name := range installs {
		path := moodleURLPath(urlPrefix, name)
		publicPath := path + "public/"
		nestedRewrite := fmt.Sprintf(`rewrite ^%s(?!public/)(.*)$ %s$1 break;`, regexp.QuoteMeta(path), publicPath)
		installDispatch := strings.ReplaceAll(dispatch, "{{NESTED_REWRITE}}", nestedRewrite)
		fmt.Fprintf(&b, `
    location ^~ %s {
        rewrite ^%s(.*)$ %s$1 break;
        try_files $uri $uri/ %sr.php$is_args$args;

        location ~ %s {
            %s
        }
    }
`, path, regexp.QuoteMeta(path), publicPath, publicPath, phpRegex, installDispatch)
	}
	return b.String()
}

// SCRIPT_NAME is left pointing at the internal, public/-inclusive path by
// default (nginx derives it from the already-rewritten $uri), and that alone
// breaks Moodle: install.php guesses $CFG->wwwroot from SCRIPT_NAME and
// refuses to proceed once it ends in "/public" (its own check for a
// misconfigured web server) — confirmed empirically against a real Moodle
// checkout, not just the generic file-exposure risk from CHANGELOG [2.2.8].
// $moodle_clean_script_name (map defined once in nginxConfTpl, see
// moodleScriptNameMap) strips the "/public" segment back out for every
// Moodle location, regardless of whether the request arrived here directly
// (install.php, index.php, ...) or via the r.php fallback.
const moodleDefaultDispatch = `{{NESTED_REWRITE}}
            include fastcgi_params;
            fastcgi_pass {{MOODLE_PHP}}:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            fastcgi_param SCRIPT_NAME $moodle_clean_script_name;`

// if/set must stay ahead of {{NESTED_REWRITE}}: rewrite's "break" flag stops
// processing of every later ngx_http_rewrite_module directive in this
// location (if/set/rewrite itself), so $php_upstream would be left
// uninitialized if the rewrite came first — confirmed empirically (nginx
// logs "using uninitialized \"php_upstream\" variable" / "no host in
// upstream" when ordered the other way around).
const moodleVersionedDispatch = `if ($p_ver = "") { set $p_ver {{MOODLE_PHP_VER}}; }
            set $php_upstream php$p_ver:9000;
            {{NESTED_REWRITE}}
            fastcgi_split_path_info ^(.+\.php)(/.+)$;
            include fastcgi_params;
            fastcgi_pass $php_upstream;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            fastcgi_param PATH_INFO $fastcgi_path_info;
            fastcgi_param SCRIPT_NAME $moodle_clean_script_name;`

// moodleScriptNameMap is emitted once at the top of nginx/default.conf
// (outside both server{} blocks — map is only valid at http context, and
// this file gets spliced into http{} via the base image's `include
// conf.d/*.conf`) whenever at least one Moodle install is configured. It
// captures everything before "/public/" and everything after, and glues
// them back together without it — see moodleDefaultDispatch's doc comment
// for why this is needed. A single generic regex handles every install's
// prefix uniformly, so this never needs to be generated per-install.
const moodleScriptNameMap = `map $fastcgi_script_name $moodle_clean_script_name {
    "~^(?<prefix>/.+)/public/(?<rest>.*)$"  "$prefix/$rest";
    default $fastcgi_script_name;
}

`

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

	scriptNameMap := ""
	if len(moodleInstalls) > 0 {
		scriptNameMap = moodleScriptNameMap
	}

	nginxConf := nginxConfTpl
	nginxConf = strings.ReplaceAll(nginxConf, "{{MOODLE_SCRIPT_NAME_MAP}}", scriptNameMap)
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

// nginxSeesFile reports whether the "nginx" container's bind-mounted
// nginx/default.conf currently matches want.
//
// This exists because "running" isn't the same as "actually seeing the file
// we just wrote" (found 2026-07-09). docker-compose bind-mounts
// nginx/default.conf by path, but the kernel binds that mount to the
// *inode* current at container-creation time. If the file on disk later
// gets replaced by one with a different inode — e.g. the user deletes and
// regenerates the whole docker-compose directory (`rm -rf` + "Criar Stack
// PHP" again) while an nginx container from a *previous* stack is still
// up — the running container keeps serving the old, now-orphaned inode's
// content. `docker restart`/a host reboot doesn't fix this either: neither
// recreates the container, so the stale bind mount survives both. Only
// actually recreating the container (`docker compose up -d --force-recreate
// nginx`, or a full `down`+`up`) re-resolves the mount.
// Without this check, `docker exec nginx nginx -t` and `nginx -s reload`
// would run inside that same stale mount namespace and "successfully"
// validate/reload the *old* content — MoodleRouter would report "Roteamento
// atualizado e aplicado sem downtime." while nothing observable actually
// changed. See CHANGELOG [2.2.9] and the lumina-tools project vault
// (decisions.md, 2026-07-09) for the full incident writeup. Trailing
// newlines are ignored in the comparison since they don't affect nginx's
// behavior and `cat`/exe.Output round-tripping isn't guaranteed to preserve
// them byte-for-byte.
func nginxSeesFile(ctx context.Context, exe *executor.Executor, want string) bool {
	out, err := exe.Output(ctx, executor.Options{}, "docker", "exec", "nginx", "cat", "/etc/nginx/conf.d/default.conf")
	if err != nil {
		return false
	}
	return strings.TrimRight(out, "\n") == strings.TrimRight(want, "\n")
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
	} else if !nginxSeesFile(ctx, exe, newContent) {
		ui.Err(stdout, "O container nginx está rodando, mas ainda enxerga uma versão antiga do nginx/default.conf "+
			"(bind mount desatualizado — comum depois de apagar/recriar a pasta da stack com o container antigo "+
			"ainda de pé). Reiniciar o container ou o computador não resolve; é preciso recriá-lo:\n\n"+
			"  cd "+cfg.DockerComposeDir+" && docker compose up -d --force-recreate nginx\n\n"+
			"O arquivo em disco já está atualizado — rode o comando acima e tente de novo.")
		ui.WaitEnter(stdout)
		return fmt.Errorf("bind mount do nginx desatualizado")
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
