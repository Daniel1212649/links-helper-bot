package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Daniel1212649/LinksHelperBot/lib/e"
	"github.com/Daniel1212649/LinksHelperBot/storage"
)

const (
	rndCmd    = "/rnd"
	helpCmd   = "/help"
	startCmd  = "/start"
	saveCmd   = "/save"
	listCmd   = "/list"
	searchCmd = "/search"
	deleteCmd = "/delete"
	statsCmd  = "/stats"
)

func (p *Processor) doCmd(ctx context.Context, text string, meta Meta) error {
	text = strings.TrimSpace(text)

	if text == "" {
		log.Printf("empty message from chat_id=%d username=%q", meta.ChatID, meta.Username)
		return p.tg.SendMessage(ctx, meta.ChatID, msgEmptyMessage)
	}

	log.Printf("got new command %q from chat_id=%d username=%q", text, meta.ChatID, meta.Username)

	if isAddCmd(text) {
		return p.savePage(ctx, meta, text)
	}

	command, argument := splitCommand(text)
	switch command {
	case rndCmd:
		return p.sendRandom(ctx, meta)
	case helpCmd:
		return p.sendHelp(ctx, meta.ChatID)
	case startCmd:
		return p.sendHello(ctx, meta.ChatID)
	case saveCmd:
		return p.savePage(ctx, meta, argument)
	case listCmd:
		return p.sendList(ctx, meta)
	case searchCmd:
		return p.search(ctx, meta, argument)
	case deleteCmd:
		return p.delete(ctx, meta, argument)
	case statsCmd:
		return p.sendStats(ctx, meta)
	default:
		return p.tg.SendMessage(ctx, meta.ChatID, msgUnknownCommand)
	}
}

func (p *Processor) savePage(ctx context.Context, meta Meta, pageURL string) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	if !isURL(pageURL) {
		return p.tg.SendMessage(ctx, meta.ChatID, msgInvalidURL)
	}

	_, err = p.storage.Save(ctx, userFromMeta(meta), pageURL)
	if errors.Is(err, storage.ErrPageExists) {
		return p.tg.SendMessage(ctx, meta.ChatID, msgAlreadyExists)
	}
	if err != nil {
		return err
	}

	return p.tg.SendMessage(ctx, meta.ChatID, msgSaved)
}

func (p *Processor) sendRandom(ctx context.Context, meta Meta) (err error) {
	defer func() { err = e.WrapIfErr("can't send random page", err) }()

	page, err := p.storage.PickRandom(ctx, userFromMeta(meta))
	if errors.Is(err, storage.ErrNoSavedPages) {
		return p.tg.SendMessage(ctx, meta.ChatID, msgNoSavedPages)
	}
	if err != nil {
		return err
	}

	if err := p.tg.SendMessage(ctx, meta.ChatID, formatPage(page)); err != nil {
		return err
	}

	return p.storage.MarkRead(ctx, userFromMeta(meta), page.ID)
}

func (p *Processor) sendList(ctx context.Context, meta Meta) error {
	pages, err := p.storage.List(ctx, userFromMeta(meta), 10)
	if err != nil {
		return e.Wrap("can't list pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, "Your list is empty.")
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages("Latest links:", pages))
}

func (p *Processor) search(ctx context.Context, meta Meta, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, "Usage: /search <text>")
	}

	pages, err := p.storage.Search(ctx, userFromMeta(meta), query, 10)
	if err != nil {
		return e.Wrap("can't search pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, "Nothing found.")
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages("Search results:", pages))
}

func (p *Processor) delete(ctx context.Context, meta Meta, argument string) error {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, "Usage: /delete <id>")
	}

	id, err := strconv.ParseInt(argument, 10, 64)
	if err != nil || id <= 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, "Link ID must be a positive number.")
	}

	if err := p.storage.Remove(ctx, userFromMeta(meta), id); err != nil {
		return e.Wrap("can't delete page", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, "Deleted.")
}

func (p *Processor) sendStats(ctx context.Context, meta Meta) error {
	stats, err := p.storage.Stats(ctx, userFromMeta(meta))
	if err != nil {
		return e.Wrap("can't get stats", err)
	}

	message := fmt.Sprintf("Stats:\nTotal: %d\nUnread: %d\nRead: %d", stats.Total, stats.Unread, stats.Read)
	return p.tg.SendMessage(ctx, meta.ChatID, message)
}

func (p *Processor) sendHelp(ctx context.Context, chatID int64) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp)
}

func (p *Processor) sendHello(ctx context.Context, chatID int64) error {
	return p.tg.SendMessage(ctx, chatID, msgHello)
}

func splitCommand(text string) (string, string) {
	parts := strings.SplitN(text, " ", 2)
	command := strings.ToLower(parts[0])
	if len(parts) == 1 {
		return command, ""
	}
	return command, strings.TrimSpace(parts[1])
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	_, err := storage.NormalizeURL(text)
	return err == nil
}

func formatPage(page *storage.Page) string {
	return fmt.Sprintf("#%d\n%s", page.ID, page.URL)
}

func formatPages(title string, pages []storage.Page) string {
	var builder strings.Builder
	builder.WriteString(title)
	for _, page := range pages {
		builder.WriteString(fmt.Sprintf("\n#%d [%s] %s", page.ID, page.Status, page.URL))
	}
	return builder.String()
}
