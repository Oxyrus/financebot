# FinanceBot

FinanceBot is a Telegram assistant that leverages OpenAI to categorize expenses from natural language messages. The bot extracts category, amount, and description, echoes a confirmation, and stores the entry in a local SQLite database for future dashboarding.

## Features
- Telegram message polling restricted to approved usernames
- Expense extraction via OpenAI Chat Completions with strict JSON responses
 - Modular Go packages for configuration, extraction, storage (SQLite), and Telegram handling
 - Makefile workflow for build, run, test, and formatting tasks

## Prerequisites
- Go 1.25 or newer
- OpenAI API key with access to `gpt-4o-mini`
- Telegram bot token created via [BotFather](https://core.telegram.org/bots#botfather)

## Setup
1. Clone the repository and install dependencies:
   ```sh
   go mod download
   ```
2. Create a `.env` (or export env vars) with:
   ```sh
   TELEGRAM_TOKEN=your-telegram-token
   OPENAI_API_KEY=your-openai-key
   AUTHORIZED_USERS=iamoxyrus,anotheruser
   DATABASE_PATH=data/financebot.db
   ```
3. Use the Makefile for common workflows:
   ```sh
   make build   # compile to bin/financebot
   make run     # start the bot locally (creates data/financebot.db by default)
   make test    # run unit/integration tests
   make fmt     # gofmt cmd/ and internal/
   ```

## Development Notes
 - Storage uses SQLite via `internal/storage/sqlite` (pure Go driver). The database file defaults to `data/financebot.db`; override with `DATABASE_PATH`. Keep backups outside the repo.
- Telemetry and structured logging hooks can be added in `internal/bot` once persistence is in place.
- Keep OpenAI prompts and Telegram responses as package-level constants to simplify testing.

## Roadmap
- [ ] Expand SQLite migrations to handle schema changes (e.g., budgets, tags)
- [ ] Build expense dashboard leveraging the stored data
