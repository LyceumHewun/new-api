# Style and conventions
- Use existing layered architecture and local patterns; keep changes surgical.
- JSON marshal/unmarshal in Go business code must use wrappers from `common/json.go`; direct `encoding/json` calls are disallowed for new business logic.
- DB code must be cross-compatible with SQLite/MySQL/PostgreSQL; prefer GORM abstractions and branch with `common.Using*` only when raw SQL is unavoidable.
- Frontend i18n uses `react-i18next`; keys are Chinese source strings, stored in `web/src/i18n/locales/{lang}.json`.
- Frontend package manager/script runner: Bun.
- Optional upstream request scalar fields must use pointers with `omitempty` to preserve explicit zero values.