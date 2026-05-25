# Contributing

📄 Portuguese version: see [CONTRIBUINDO.md](CONTRIBUINDO.md)

## Guidelines

- Keep changes focused on the requested task
- Preserve the current architecture and project structure
- Prefer small, precise edits over broad refactors
- Follow the shell development conventions documented in `AGENTS.md`
- Write Markdown documentation in English unless the file is an explicit translation
- Write code comments in English

## Local Run

```bash
bash lumina-lm.sh
```

## Validation

Run ShellCheck on changed shell scripts when available:

```bash
shellcheck --severity=warning --shell=bash --exclude=SC1091 lumina-lm.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 scripts/lib/utils.sh
```

## Documentation

- Keep `README.md` in English and `LEIAME.md` in Brazilian Portuguese
- Keep `CONTRIBUTING.md` in English and `CONTRIBUINDO.md` in Brazilian Portuguese
- Preserve counterpart links between translated files
- Ensure every Markdown file ends with the repository signature

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
