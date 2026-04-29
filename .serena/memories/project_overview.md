# Project overview
- Go AI API gateway/proxy with Gin, GORM, multiple upstream provider adapters, user/billing/rate-limit/admin dashboard.
- Frontend: React 18 + Vite + Semi UI under `web/`, Bun preferred.
- Backend layout: `router/`, `controller/`, `service/`, `model/`, `relay/`, `middleware/`, `setting/`, `common/`, `dto/`, `types/`.
- DB support must remain SQLite, MySQL >= 5.7.8, PostgreSQL >= 9.6.
- Important project policy: do not remove/rename protected `new-api` / `QuantumNous` branding or metadata.