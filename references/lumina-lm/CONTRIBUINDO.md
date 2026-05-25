# Contribuindo

📄 English version: see [CONTRIBUTING.md](CONTRIBUTING.md)

## Diretrizes

- Mantenha as mudanças focadas na tarefa solicitada
- Preserve a arquitetura atual e a estrutura do projeto
- Prefira alterações pequenas e precisas em vez de refatorações amplas
- Siga as convenções de desenvolvimento shell documentadas em `AGENTS.md`
- Escreva documentação Markdown em inglês, exceto quando o arquivo for uma tradução explícita
- Escreva comentários de código em inglês

## Execução Local

```bash
bash lumina-lm.sh
```

## Validação

Execute ShellCheck nos scripts shell alterados quando disponível:

```bash
shellcheck --severity=warning --shell=bash --exclude=SC1091 lumina-lm.sh
shellcheck --severity=warning --shell=bash --exclude=SC1091 scripts/lib/utils.sh
```

## Documentação

- Mantenha `README.md` em inglês e `LEIAME.md` em português do Brasil
- Mantenha `CONTRIBUTING.md` em inglês e `CONTRIBUINDO.md` em português do Brasil
- Preserve os links entre arquivos traduzidos
- Garanta que todo arquivo Markdown termine com a assinatura do repositório

---

Made with ❤️ and AI by [Kadu Velasco](https://github.com/kaduvelasco)
