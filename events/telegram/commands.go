package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
	noteCmd   = "/note"
	remindCmd = "/remind"
)

func (p *Processor) doCmd(ctx context.Context, text string, meta Meta) error {
	text = strings.TrimSpace(text)

	if text == "" {
		log.Printf("empty message from chat_id=%d username=%q", meta.ChatID, meta.Username)
		locale := p.locale(ctx, meta)
		return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).EmptyMessage, mainMenuKeyboard(locale))
	}

	log.Printf("got new command %q from chat_id=%d username=%q", text, meta.ChatID, meta.Username)

	if pageURL, note, remindAt, ok, err := parseSaveArgs(text, nowInMoscow()); ok {
		if err != nil {
			locale := p.locale(ctx, meta)
			return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).InvalidReminderDate, mainMenuKeyboard(locale))
		}
		return p.savePage(ctx, meta, pageURL, note, remindAt)
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
		pageURL, note, remindAt, ok, err := parseSaveArgs(argument, nowInMoscow())
		if err != nil {
			locale := p.locale(ctx, meta)
			return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).InvalidReminderDate, mainMenuKeyboard(locale))
		}
		if !ok {
			return p.savePage(ctx, meta, "", "", nil)
		}
		return p.savePage(ctx, meta, pageURL, note, remindAt)
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
	case noteCmd:
		return p.setNote(ctx, meta, argument)
	case remindCmd:
		return p.setReminder(ctx, meta, argument)
	default:
		locale := p.locale(ctx, meta)
		return p.tg.SendMessage(ctx, meta.ChatID, tr(locale).UnknownCommand, mainMenuKeyboard(locale))
	}
}

func (p *Processor) savePage(ctx context.Context, meta Meta, pageURL string, note string, remindAt *time.Time) (err error) {
	defer func() { err = e.WrapIfErr("can't save page", err) }()

	locale := p.locale(ctx, meta)
	messages := tr(locale)
	if !isURL(pageURL) {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.InvalidURL, mainMenuKeyboard(locale))
	}

	_, err = p.storage.SaveWithDetails(ctx, userFromMeta(meta), pageURL, strings.TrimSpace(note), remindAt)
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

func (p *Processor) setNote(ctx context.Context, meta Meta, argument string) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)

	id, note, ok := splitIDAndText(argument)
	if !ok || note == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.NoteUsage, mainMenuKeyboard(locale))
	}

	if err := p.storage.SetNote(ctx, userFromMeta(meta), id, note); err != nil {
		return e.Wrap("can't set note", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, messages.NoteSaved, mainMenuKeyboard(locale))
}

func (p *Processor) setReminder(ctx context.Context, meta Meta, argument string) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)

	id, rawDate, ok := splitIDAndText(argument)
	if !ok || rawDate == "" {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.ReminderUsage, mainMenuKeyboard(locale))
	}

	remindAt, err := parseReminderTime(rawDate, nowInMoscow())
	if err != nil {
		return p.tg.SendMessage(ctx, meta.ChatID, messages.InvalidReminderDate, mainMenuKeyboard(locale))
	}

	if err := p.storage.SetReminder(ctx, userFromMeta(meta), id, remindAt); err != nil {
		return e.Wrap("can't set reminder", err)
	}

	return p.tg.SendMessage(ctx, meta.ChatID, fmt.Sprintf(messages.ReminderSavedFormat, formatReminderTime(remindAt)), mainMenuKeyboard(locale))
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

func splitURLAndNote(text string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(text), " ", 2)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.TrimSpace(parts[1])
}

func parseSaveArgs(text string, now time.Time) (string, string, *time.Time, bool, error) {
	pageURL, note := splitURLAndNote(text)
	if !isURL(pageURL) {
		return "", "", nil, false, nil
	}

	note, rawReminder := splitReminderOption(note)
	if rawReminder == "" {
		return pageURL, note, nil, true, nil
	}

	remindAt, err := parseReminderTime(rawReminder, now)
	if err != nil {
		return "", "", nil, true, err
	}
	return pageURL, note, &remindAt, true, nil
}

func splitReminderOption(text string) (string, string) {
	text = strings.TrimSpace(text)
	const marker = "--remind"
	if strings.HasPrefix(text, marker+" ") {
		return "", strings.TrimSpace(strings.TrimPrefix(text, marker))
	}

	index := strings.LastIndex(text, " "+marker+" ")
	if index == -1 {
		return text, ""
	}

	note := strings.TrimSpace(text[:index])
	reminder := strings.TrimSpace(text[index+len(" "+marker+" "):])
	return note, reminder
}

func splitIDAndText(text string) (int64, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(text), " ", 2)
	if len(parts) != 2 {
		return 0, "", false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, "", false
	}
	return id, strings.TrimSpace(parts[1]), true
}

func parseReminderTime(raw string, now time.Time) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	layouts := []string{
		"2006-01-02 15:04",
		"02.01.2006 15:04",
		"2006-01-02",
		"02.01.2006",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, raw, moscowLocation)
		if err != nil {
			continue
		}
		if layout == "2006-01-02" || layout == "02.01.2006" {
			parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 9, 0, 0, 0, moscowLocation)
		}
		if parsed.Before(now) {
			return time.Time{}, fmt.Errorf("reminder time is in the past")
		}
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("unsupported reminder time format")
}

func formatReminderTime(t time.Time) string {
	return t.In(moscowLocation).Format("2006-01-02 15:04")
}

func isURL(text string) bool {
	text = strings.TrimSpace(text)
	lower := strings.ToLower(text)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") && !strings.Contains(text, ".") {
		return false
	}

	_, err := storage.NormalizeURL(text)
	return err == nil
}

func formatRandomPage(page *storage.Page) string {
	return formatPage(page)
}

func formatPages(title string, pages []storage.Page) string {
	var builder strings.Builder
	builder.WriteString(title)
	for _, page := range pages {
		builder.WriteString("\n")
		builder.WriteString(formatPage(&page))
	}
	return builder.String()
}

func formatPage(page *storage.Page) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("#%d [%s]\n%s", page.ID, page.Status, page.URL))
	if strings.TrimSpace(page.Note) != "" {
		builder.WriteString(fmt.Sprintf("\n📝 %s", page.Note))
	}
	if page.RemindAt != nil {
		builder.WriteString(fmt.Sprintf("\n⏰ %s", formatReminderTime(*page.RemindAt)))
	}
	return builder.String()
}
