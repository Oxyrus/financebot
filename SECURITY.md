# Security Policy

## Supported Versions

| Version            | Supported                   |
| ------------------ | --------------------------- |
| main               | ✅                          |
| Older tags / forks | ⚠️ Community supported only |

Only the latest `main` branch is actively maintained. If you deploy a pinned commit or fork, you are responsible for cherry-picking security fixes.

## Security Guidelines

- **Secrets**: Keep `.env` files and API keys outside version control. Rotate `TELEGRAM_TOKEN`, `OPENAI_API_KEY`, and database credentials immediately if compromise is suspected.
- **Transport**: The bot communicates with Telegram and OpenAI over HTTPS. When running inside Docker or on remote hosts, secure outbound traffic with firewall rules and avoid exposing SQLite over the network.
- **Storage**: SQLite data lives under `data/`. Use filesystem permissions or encrypted volumes to protect it. Backups created with `rclone` or similar tools should be encrypted server-side.
- **Dependencies**: Run `go list -m -u all` or `go mod tidy` periodically and review release notes for `github.com/sashabaranov/go-openai`, `github.com/go-telegram-bot-api/telegram-bot-api`, and `modernc.org/sqlite`. Apply security patches promptly.
- **Containers**: The published Docker image is non-root and distroless. When deploying elsewhere, maintain the same security posture (read-only root FS, drop unnecessary capabilities).
