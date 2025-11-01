# Repository Guidelines

## Project Structure & Module Organization
The module is `github.com/Oxyrus/financebot`. Runtime entry stays in `cmd/main.go`, but long-term logic should move into packages such as `internal/bot` (Telegram plumbing), `internal/expense` (domain models + validation), and `internal/storage` (database adapters). Keep shared utilities in `pkg/` if you need to export them. Add `configs/` for sample `.env` files and create a `migrations/` folder when the expense database is introduced.

## Build, Test, and Development Commands
- `go run ./cmd` boots the Telegram bot against your local environment variables.
- `go build ./cmd` produces a deployable binary; combine with a systemd service or container for prod.
- `go test ./...` runs every unit test; gate pull requests on a clean run.

## Coding Style & Naming Conventions
Run `gofmt` (tabs; idiomatic Go spacing) before committing. Use CamelCase for exported identifiers (`ExpenseRepository`) and mixedCaps for internal helpers. Keep prompt templates and message text in package-level constants. Prefer constructor functions that accept interfaces (`NewBotHandler(client OpenAI, repo ExpenseStore)`) to support mocks and future adapters.

## Testing Guidelines
Rely on the standard `testing` package with table-driven cases. Store test doubles under `internal/mocks`. Name files `*_test.go` and functions `TestFeatureName`. Cover message parsing, expense extraction edge cases, and storage failures. For async bot loops, wrap handlers so they can be invoked synchronously by tests.

## Data & Persistence Practices
SQLite persistence now lives in `internal/storage/sqlite` using the pure Go `modernc.org/sqlite` driver. The database file defaults to `data/financebot.db`; override with the `DATABASE_PATH` env var. Migrations run automatically on startup—extend the schema statements as new features land. Keep `.db` artifacts out of git but schedule periodic backups (e.g., cron to Google Drive). Rotate API keys promptly if leaked.

## Commit & Pull Request Guidelines
Write imperative, concise commit subjects (~50 chars). Include body details when you add a feature, fix a bug, or change schema. PRs should describe the problem, the solution, testing evidence (`go test ./...` output), and screenshots/logs when behavior changes. Keep PRs scoped; submit refactors separately from feature work.
