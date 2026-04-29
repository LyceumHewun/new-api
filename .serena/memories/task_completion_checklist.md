# Task completion checklist
- Run focused format/build/test commands for changed area.
- For Go changes, run `gofmt` on touched Go files and targeted `go test` when possible.
- For frontend changes, prefer `bun run build` or focused i18n/lint checks as appropriate.
- Check `git diff --check` and `git status --short` before final response.
- Mention any checks that could not be run.