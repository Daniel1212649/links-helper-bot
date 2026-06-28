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
	langCmd   = "/lang"
)

func (p *Processor) doCmd(ctx context.Context, text string, meta Meta) error {
	text = strings.TrimSpace(text)

	if text == "" {
		log.Printf("empty message from chat_id=%d username=%q", meta.ChatID, meta.Username)
		locale := p.locale(ctx, meta)
		return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).EmptyMessage, mainMenuKeyboard(locale))
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
		return p.sendHelp(ctx, meta)
	case startCmd:
		return p.sendHello(ctx, meta)
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
	case langCmd:
		return p.setLocaleCommand(ctx, meta, argument)
	default:
		locale := p.locale(ctx, meta)
		return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).UnknownCommand, mainMenuKeyboard(locale))
	}
}

func (p *Processor) savePage(ctx context.Context, meta Meta, pageURL string) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	locale := p.locale(ctx, meta)
	messages := tr(locale)
	if !isURL(pageURL) {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.InvalidURL, mainMenuKeyboard(locale))
	}

	_, err = p.storage.Save(ctx, userFromMeta(meta), pageURL)
	if errors.Is(err, storage.ErrPageExists) {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.AlreadyExists, mainMenuKeyboard(locale))
	}
	if err != nil {
		return err
	}

	return p.tg.SendMessage(ctx, meta.ChatID, messages.Saved, mainMenuKeyboard(locale))
}

func (p *Processor) sendRandomPage(ctx context.Context, meta Meta, editMessageID int) (err error) {
	defer func() { err = e.WrapIfErr("can't send random page", err) }()

	locale := p.locale(ctx, meta)
	messages := tr(locale)
	page, err := p.storage.PickRandom(ctx, userFromMeta(meta))
	if errors.Is(err, storage.ErrNoSavedPages) {
		text := messages.NoSavedPages
		if editMessageID > 0 {
			return p.tg.EditMessageText(ctx, meta.ChatID, editMessageID, text, mainMenuKeyboard(locale))
		}
		return p.tg.SendMessage(ctx, meta.ChatID, text, mainMenuKeyboard(locale))
	}
	if err != nil {
		return err
	}

	text := formatRandomPage(page)
	markup := linkActionKeyboard(locale, page.ID)
	if editMessageID > 0 {
		return p.tg.EditMessageText(ctx, meta.ChatID, editMessageID, text, markup)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, text, markup)
}

func (p *Processor) sendList(ctx context.Context, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	pages, err := p.storage.List(ctx, userFromMeta(meta), 10)
	if err != nil {
		return e.Wrap("can't list pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.EmptyList, mainMenuKeyboard(locale))
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages(messages.LatestLinksTitle, pages), listActionKeyboard(locale, pages))
}

func (p *Processor) search(ctx context.Context, meta Meta, query string) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	query = strings.TrimSpace(query)
	if query == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.SearchUsage, mainMenuKeyboard(locale))
	}

	pages, err := p.storage.Search(ctx, userFromMeta(meta), query, 10)
	if err != nil {
		return e.Wrap("can't search pages", err)
	}
	if len(pages) == 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.NothingFound, mainMenuKeyboard(locale))
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatPages(messages.SearchResultsTitle, pages), listActionKeyboard(locale, pages))
}

func (p *Processor) delete(ctx context.Context, meta Meta, argument string) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	argument = strings.TrimSpace(argument)
	if argument == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.DeleteUsage, mainMenuKeyboard(locale))
	}

	id, err := strconv.ParseInt(argument, 10, 64)
	if err != nil || id <= 0 {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.InvalidLinkID, mainMenuKeyboard(locale))
	}

	if err := p.storage.Remove(ctx, userFromMeta(meta), id); err != nil {
		return e.Wrap("can't delete page", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, messages.Deleted, mainMenuKeyboard(locale))
}

func (p *Processor) sendStats(ctx context.Context, meta Meta) error {
	locale := p.locale(ctx, meta)
	stats, err := p.storage.Stats(ctx, userFromMeta(meta))
	if err != nil {
		return e.Wrap("can't get stats", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, formatStatsMessage(locale, stats), mainMenuKeyboard(locale))
}

func (p *Processor) sendHelp(ctx context.Context, meta Meta) error {
	locale := p.locale(ctx, meta)
	return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).Help, mainMenuKeyboard(locale))
}

func (p *Processor) sendHello(ctx context.Context, meta Meta) error {
	locale := p.locale(ctx, meta)
	return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).Hello, mainMenuKeyboard(locale))
}

func (p *Processor) sendLanguagePicker(ctx context.Context, meta Meta) error {
	locale := p.locale(ctx, meta)
	return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).ChooseLanguage, languageKeyboard(locale))
}

func (p *Processor) setLocaleCommand(ctx context.Context, meta Meta, argument string) error {
	locale := storage.NormalizeLocale(argument)
	if strings.TrimSpace(argument) == "" {
		return p.sendLanguagePicker(ctx, meta)
	}
	if err := p.storage.SetLocale(ctx, userFromMeta(meta), locale); err != nil {
		return err
	}
	return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).LanguageUpdated, mainMenuKeyboard(locale))
}

func (p *Processor) locale(ctx context.Context, meta Meta) string {
	locale, err := p.storage.GetLocale(ctx, userFromMeta(meta))
	if err != nil {
		return localeFromLanguageCode(meta.LanguageCode)
	}
	return storage.NormalizeLocale(locale)
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
