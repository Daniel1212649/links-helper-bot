# LinksHelperBot

Telegram bot for saving links and returning to them later.

## Features

- Save any `http` or `https` link by sending it to the bot.
- Save explicitly with `/save <url>`.
- Get a random unread link with `/rnd` or the 🎲 button.
- List latest links with `/list`.
- Search saved links with `/search <text>`.
- Delete a link with `/delete <id>` or inline 🗑 buttons.
- Check counters with `/stats`.
- Use inline buttons for quick navigation and link actions.
- Store data in PostgreSQL.
- Run locally or on a server with Docker Compose.

## Commands

```text
/start          👋 greeting and short help
/help           📖 show all commands
/save <url>     💾 save a link
/rnd            🎲 random unread link (Read / Delete / Another buttons)
/list           📋 show latest saved links
/stats          📊 show total, unread and read counters
/search <text>  🔍 search by URL or title
/delete <id>    🗑 delete a saved link by ID
```

Inline buttons: 👋 start, 📖 help, 💾 save, 🎲 random, 📋 list, 📊 stats, 🔍 search, 🗑 delete.

## Local Run With Docker

Create an environment file:

```bash
cp .env.example .env
```

Set your Telegram token in `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:your_token
```

Start the stack:

```bash
docker compose up --build
```

The stack contains:

- `bot` - Go application.
- `postgres` - PostgreSQL database with persistent Docker volume.
- `migrate` - one-shot migration runner.

PostgreSQL data is stored in the `postgres_data` Docker volume. Migrations from `migrations/` are applied automatically before the bot starts, so data survives restarts and version updates.

## Local Run Without Docker

You need Go 1.24+ and PostgreSQL.

```bash
export TELEGRAM_BOT_TOKEN=123456:your_token
export DATABASE_URL='postgres://links_helper:links_helper_password@localhost:5432/links_helper?sslmode=disable'
go test ./...
go run ./cmd/links-helper-bot
```

Apply the SQL migration from `migrations/000001_init.up.sql` before running the bot.

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `TELEGRAM_BOT_TOKEN` | yes | - | Telegram Bot API token |
| `DATABASE_URL` | yes | - | PostgreSQL connection string |
| `TELEGRAM_API_HOST` | no | `api.telegram.org` | Telegram API host |
| `APP_ENV` | no | `local` | Environment name |
| `LOG_LEVEL` | no | `info` | Reserved for structured logging |
| `POLL_BATCH_SIZE` | no | `100` | Telegram updates batch size |
| `POLL_INTERVAL` | no | `1s` | Delay between empty polling iterations |
| `HTTP_TIMEOUT` | no | `35s` | Telegram HTTP client timeout |

## Development

```bash
make test
make vet
make build
make migrate
make compose-config
```

## Deployment

See [docs/deploy.md](docs/deploy.md).

## Roadmap

- Metadata fetching for page titles.
- Postgres storage integration tests.
- Move packages to `internal/`.
