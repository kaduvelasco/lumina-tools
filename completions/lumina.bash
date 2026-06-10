#!/usr/bin/env bash
# Bash completion for lumina
# Install: source /path/to/lumina.bash
#          or copy to /etc/bash_completion.d/lumina

_lumina_completions() {
    local cur words cword cmd subcmd
    cur="${COMP_WORDS[COMP_CWORD]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD

    # ── Level 1: top-level commands ───────────────────────────────────────────
    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "
            system stack dev
            apps gnome ai
            db repo set
            self-update self-uninstall self-config
            version help
        " -- "$cur"))
        return 0
    fi

    cmd="${words[1]}"

    # ── Level 2: subcommands ──────────────────────────────────────────────────
    if [[ $cword -eq 2 ]]; then
        case "$cmd" in
            system)
                COMPREPLY=($(compgen -W "pos fonts templates apps gnome update ulauncher toys megasync" -- "$cur"))
                ;;
            stack)
                COMPREPLY=($(compgen -W "config start end restart log status db fix-perm" -- "$cur"))
                ;;
            dev)
                COMPREPLY=($(compgen -W "pre go llm ide term mcp update create-workspace create-stack-php" -- "$cur"))
                ;;
            apps)
                COMPREPLY=($(compgen -W "install uninstall web" -- "$cur"))
                ;;
            gnome)
                COMPREPLY=($(compgen -W "pre ext themes icons cursor flatpak" -- "$cur"))
                ;;
            ai)
                COMPREPLY=($(compgen -W "context clear" -- "$cur"))
                ;;
            db)
                COMPREPLY=($(compgen -W "backup restore remove optimize moodle" -- "$cur"))
                ;;
            repo)
                COMPREPLY=($(compgen -W "global init clone ident gitignore conduct" -- "$cur"))
                ;;
            set)
                COMPREPLY=($(compgen -W "workspace docker theme flatpak" -- "$cur"))
                ;;
        esac
        return 0
    fi

    subcmd="${words[2]}"

    # ── Level 3: arguments ────────────────────────────────────────────────────
    if [[ $cword -eq 3 ]]; then
        case "$cmd/$subcmd" in
            system/pos)
                COMPREPLY=($(compgen -W "mint zorin ubuntu fedora" -- "$cur"))
                ;;
            system/apps)
                COMPREPLY=($(compgen -W "install uninstall webapps" -- "$cur"))
                ;;
            stack/config)
                COMPREPLY=($(compgen -W "docker workspace stack" -- "$cur"))
                ;;
            set/theme)
                COMPREPLY=($(compgen -W "lumina light dracula nord tokyo gruvbox" -- "$cur"))
                ;;
            set/flatpak)
                COMPREPLY=($(compgen -W "user system" -- "$cur"))
                ;;
            set/workspace|set/docker)
                COMPREPLY=($(compgen -d -- "$cur"))
                ;;
        esac
        return 0
    fi

    COMPREPLY=()
}

complete -F _lumina_completions lumina
