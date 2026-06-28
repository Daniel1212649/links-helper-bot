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
		return p.tg.SendMessage(ctx, meta.ChatID, msgEmptyMessage, mainMenuKeyboard())
	}

	log.Printf("got new command %q from chat_id=%d username=%q", text, meta.ChatID, meta.Username)

	if isAddCmd(text) {
		return p.savePage(ctx, meta, text)
	}

	command, argument := splitCommand(text)
	switch command {
	case rndCmd:
		return p.sendRandomPage(ctx, meta, 0)
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
		return p.tg.SendMessage(ctx, meta.ChatID, msgUnknownCommand, mainMenuKeyboard())
	}
}

func (p *Processor) savePage(ctx context.Context, meta Meta, pageURL string) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	if !isURL(pageURL) {
		return p.tg.SendMessage(ctx, meta.ChatID, msgInvalidURL, mainMenuKeyboard())
	}

	_, err = p.storage.Save(ctx, userFromMeta(meta), pageURL)
	if errors.Is(err, storage.ErrPageExists) {
		return p.tg.SendMessage(ctx, meta.ChatID, msgAlreadyExists, mainMenuKeyboard())
	}
	if err != nil {
		return err
	}

	return p.tg.SendMessage(ctx, meta.ChatID, msgSaved, mainMenuKeyboard())
}

func (p *Processor) sendRandomPage(ctx context.Context, meta Meta, editMessageID int) (err error) {
	defer func() { err = e.WrapIfErr("can't send random page", err) }()

	page, err := p.storage.PickRandom(ctx, userFromMeta(meta))
	if errors.Is(err, storage.ErrNoSavedPages) {
		text := msgNoSavedPages
		if editMessageID > 0 {
			return p.tg.EditMessageText(ctx, meta.ChatID, editMessageID, text, mainMenuKeyboard())
		}
		return p.tg.SendMessage(ctx, meta.ChatID, text, mainMenuKeyboard())
	}
	if err != nil {
		return err
	}

	text := formatRandomPage(page)
	markup := linkActionKeyboard(page.ID)
	if editMessageID > 0 {
		return p.tg.EditMessageText(ctx, meta.ChatID, editMessageID, text, markup)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, text, markup)
}

func (p *Processor) sendList(ctx context.Context, meta Meta) error {
	pages, err := p.storage.List(ctx, userFromMeta(meta), 10)
	if err != nil {
		return e.Wrap("can't list pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, msgEmptyList, mainMenuKeyboard())
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages("Latest links:", pages), listActionKeyboard(pages))
}

func (p *Processor) search(ctx context.Context, meta Meta, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, msgSearchUsage, mainMenuKeyboard())
	}

	pages, err := p.storage.Search(ctx, userFromMeta(meta), query, 10)
	if err != nil {
		return e.Wrap("can't search pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, msgNothingFound, mainMenuKeyboard())
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages("Search results:", pages), listActionKeyboard(pages))
}

func (p *Processor) delete(ctx context.Context, meta Meta, argument string) error {
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, msgDeleteUsage, mainMenuKeyboard())
	}

	id, err := strconv.ParseInt(argument, 10, 64)
	if err != nil || id <= 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, msgInvalidLinkID, mainMenuKeyboard())
	}

	if err := p.storage.Remove(ctx, userFromMeta(meta), id); err != nil {
		return e.Wrap("can't delete page", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, msgDeleted, mainMenuKeyboard())
}

func (p *Processor) sendStats(ctx context.Context, meta Meta) error {
	stats, err := p.storage.Stats(ctx, userFromMeta(meta))
	if err != nil {
		return e.Wrap("can't get stats", err)
	}

	message := fmt.Sprintf("Stats:\nTotal: %d\nUnread: %d\nRead: %d", stats.Total, stats.Unread, stats.Read)
	return p.tg.SendMessage(ctx, meta.ChatID, message, mainMenuKeyboard())
}

func (p *Processor) sendHelp(ctx context.Context, chatID int64) error {
	return p.tg.SendMessage(ctx, chatID, msgHelp, mainMenuKeyboard())
}

func (p *Processor) sendHello(ctx context.Context, chatID int64) error {
	return p.tg.SendMessage(ctx, chatID, msgHello, mainMenuKeyboard())
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

func formatRandomPage(page *storage.Page) string {
	return fmt.Sprintf("#%d [%s]\n%s", page.ID, page.Status, page.URL)
}

func formatPages(title string, pages []storage.Page) string {
	var builder strings.Builder
	builder.WriteString(title)
	for _, page := range pages {
		builder.WriteString(fmt.Sprintf("\n#%d [%s] %s", page.ID, page.Status, page.URL))
	}
	return builder.String()
}
