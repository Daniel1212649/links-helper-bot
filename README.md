# LinksHelperBot

Telegram bot for saving links and returning to them later.

## Features

- Save any `http` or `https` link by sending it to the bot.
- Save explicitly with `/save <url>`.
- Get a random unread link with `/rnd`.
- List latest links with `/list`.
- Search saved links with `/search <text>`.
- Delete a link with `/delete <id>`.
- Check counters with `/stats`.
- Store data in PostgreSQL.
- Run locally or on a server with Docker Compose.

## Commands

```text
/start          show greeting and help
/help           show help
/save <url>     save a link
/rnd            get a random unread link and mark it as read
/list           show latest saved links
/search <text>  search by URL or title
/delete <id>    delete a saved link
/stats          show total, unread and read counters
```

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

PostgreSQL is initialized from `migrations/000001_init.up.sql` on the first run.

## Local Run Without Docker

You need Go 1.24+ and PostgreSQL.

```bash
export TELEGRAM_BOT_TOKEN=123456:your_token
export DATABASE_URL='postgres://links_helper:links_helper_password@localhost:5432/links_helper?sslmode=disable'
go test ./...
go run .
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
make compose-config
```

## Deployment

See [docs/deploy.md](docs/deploy.md).

## Roadmap

- Inline buttons for `Read`, `Delete`, `Another`.
- Metadata fetching for page titles.
- File-to-PostgreSQL migration command for old gob storage.
- CI workflow for tests and Docker image build.
