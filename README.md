# FinanceBot

FinanceBot is a Telegram assistant that leverages OpenAI to categorize expenses from natural language messages. The bot extracts category, amount, and description, echoes a confirmation, and stores the entry (currently in-memory with a pluggable storage layer ready for a database backend).

## Features
- Telegram message polling restricted to approved usernames
- Expense extraction via OpenAI Chat Completions with strict JSON responses
- Modular Go packages for configuration, extraction, storage, and Telegram handling
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
   ```
3. Use the Makefile for common workflows:
   ```sh
   make build   # compile to bin/financebot
   make run     # start the bot locally
   make test    # run unit tests
   make fmt     # gofmt cmd/ and internal/
   ```

## Development Notes
- Storage currently defaults to `internal/storage/memory`. Replace it with a database-backed implementation under `internal/storage/<db>`; expose configuration via `DATABASE_URL` and migrations.
- Telemetry and structured logging hooks can be added in `internal/bot` once persistence is in place.
- Keep OpenAI prompts and Telegram responses as package-level constants to simplify testing.

## Roadmap
- [ ] Introduce persistent storage (Postgres/SQLite) with migrations
- [ ] Add unit tests for extractor, bot handler, and storage adapters
- [ ] Build expense dashboard leveraging the stored data
