# Suggested commands
- Check status: `git status --short`
- Search tracked files: `git grep -n -I "pattern" -- path`
- Backend format: `gofmt -w <files>`
- Backend tests: `go test ./...`
- Frontend install: `cd web; bun install`
- Frontend build: `cd web; bun run build`
- Frontend lint/format check: `cd web; bun run lint`
- Frontend i18n tools: `cd web; bun run i18n:extract`, `bun run i18n:sync`, `bun run i18n:lint`