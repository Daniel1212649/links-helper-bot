# Production Deploy

This guide describes a simple VPS deployment with Docker Compose and Telegram long polling.

## Server Requirements

- Ubuntu 22.04/24.04 or another Linux server.
- Docker Engine.
- Docker Compose plugin.
- Outbound HTTPS access to `api.telegram.org`.

You do not need a domain or reverse proxy for long polling.

## First Deploy

Create an application directory:

```bash
sudo mkdir -p /opt/links-helper-bot
sudo chown "$USER":"$USER" /opt/links-helper-bot
cd /opt/links-helper-bot
```

Clone the repository:

```bash
git clone https://github.com/Daniel1212649/LinksHelperBot.git .
```

Create the environment file:

```bash
cp .env.example .env
```

Edit `.env`:

```env
TELEGRAM_BOT_TOKEN=123456:your_token
POSTGRES_DB=links_helper
POSTGRES_USER=links_helper
POSTGRES_PASSWORD=change_this_password
DATABASE_URL=postgres://links_helper:change_this_password@postgres:5432/links_helper?sslmode=disable
APP_ENV=production
LOG_LEVEL=info
```

Start services:

```bash
docker compose up -d --build
```

Check logs:

```bash
docker compose logs -f bot
```

## Update

```bash
cd /opt/links-helper-bot
git pull
docker compose up -d --build
```

## Stop

```bash
docker compose down
```

This keeps the PostgreSQL Docker volume. To remove data too, run `docker compose down -v`.

## Backup

Create a database dump:

```bash
docker compose exec postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > links_helper_backup.sql
```

Restore into a fresh database:

```bash
docker compose exec -T postgres psql -U "$POSTGRES_USER" "$POSTGRES_DB" < links_helper_backup.sql
```

## Logs

```bash
docker compose logs -f bot
docker compose logs -f postgres
```

## Notes

- Keep `.env` private. Never commit real tokens or passwords.
- Rotate `POSTGRES_PASSWORD` before production use.
- Long polling is enough for a small personal bot. Use webhooks later only if you need lower latency or a reverse-proxy-based setup.
