package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Daniel1212649/LinksHelperBot/internal/lib/e"
	"github.com/Daniel1212649/LinksHelperBot/internal/storage"
)

type CallbackMeta struct {
	CallbackQueryID string
	ChatID          int64
	MessageID       int
	TelegramID      int64
	Username        string
	LanguageCode    string
	Data            string
}

func (p *Processor) processCallback(ctx context.Context, event CallbackMeta) error {
	meta := Meta{
		ChatID:       event.ChatID,
		TelegramID:   event.TelegramID,
		Username:     event.Username,
		LanguageCode: event.LanguageCode,
	}
	locale := p.locale(ctx, meta)
	messages := tr(locale)

	switch event.Data {
	case cbCmdStart:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendHello(ctx, meta)
		})
	case cbCmdHelp:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendHelp(ctx, meta)
		})
	case cbCmdSave:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.SavePrompt, mainMenuKeyboard(locale))
		})
	case cbCmdRnd:
		if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
			return err
		}
		return p.sendRandomPage(ctx, meta, event.MessageID)
	case cbCmdList:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendList(ctx, meta, "")
		})
	case cbCmdStats:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendStats(ctx, meta)
		})
	case cbCmdSearch:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.SearchPrompt, mainMenuKeyboard(locale))
		})
	case cbCmdDelete:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.DeletePrompt, mainMenuKeyboard(locale))
		})
	case cbCmdNote:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.NoteUsage, mainMenuKeyboard(locale))
		})
	case cbCmdRemind:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.ReminderUsage, mainMenuKeyboard(locale))
		})
	case cbCmdGroups:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendGroups(ctx, meta)
		})
	case cbCmdLang:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, messages.ChooseLanguage, languageKeyboard(locale))
		})
	case cbLangRU:
		return p.setLocaleCallback(ctx, event, meta, storage.LocaleRU)
	case cbLangEN:
		return p.setLocaleCallback(ctx, event, meta, storage.LocaleEN)
	default:
		if strings.HasPrefix(event.Data, "read:") {
			return p.markReadCallback(ctx, event, meta)
		}
		if strings.HasPrefix(event.Data, "del:") {
			return p.deleteCallback(ctx, event, meta)
		}
		if strings.HasPrefix(event.Data, "note:") {
			return p.notePromptCallback(ctx, event, meta)
		}
		if strings.HasPrefix(event.Data, "remind:") {
			return p.reminderPromptCallback(ctx, event, meta)
		}
		if strings.HasPrefix(event.Data, "group:") {
			return p.groupPromptCallback(ctx, event, meta)
		}
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.UnknownAction)
	}
}

func (p *Processor) answerAndSend(ctx context.Context, event CallbackMeta, action func() error) error {
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
		return err
	}
	return action()
}

func (p *Processor) setLocaleCallback(ctx context.Context, event CallbackMeta, meta Meta, locale string) error {
	locale = storage.NormalizeLocale(locale)
	if err := p.storage.SetLocale(ctx, userFromMeta(meta), locale); err != nil {
		return err
	}
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, tr(locale).LanguageUpdated); err != nil {
		return err
	}

	message := tr(locale).LanguageUpdated
	if event.MessageID > 0 {
		return p.tg.EditMessageText(ctx, event.ChatID, event.MessageID, message, mainMenuKeyboard(locale))
	}
	return p.tg.SendMessage(ctx, event.ChatID, message, mainMenuKeyboard(locale))
}

func (p *Processor) markReadCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	id, err := parseCallbackID(event.Data, "read:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.InvalidCallbackLinkID)
	}

	if err := p.storage.MarkRead(ctx, userFromMeta(meta), id); err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.CouldNotMarkRead)
	}

	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.MarkedRead); err != nil {
		return err
	}

	if event.MessageID == 0 {
		return nil
	}

	text := fmt.Sprintf("#%d\n%s", id, messages.MarkedReadMessage)
	return p.tg.EditMessageText(ctx, event.ChatID, event.MessageID, text, mainMenuKeyboard(locale))
}

func (p *Processor) deleteCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	id, err := parseCallbackID(event.Data, "del:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.InvalidCallbackLinkID)
	}

	if err := p.storage.Remove(ctx, userFromMeta(meta), id); err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.CouldNotDelete)
	}

	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.DeletedCallback); err != nil {
		return err
	}

	if event.MessageID == 0 {
		return p.tg.SendMessage(ctx, event.ChatID, fmt.Sprintf("#%d\n%s", id, messages.DeletedMessage), mainMenuKeyboard(locale))
	}

	text := fmt.Sprintf("#%d\n%s", id, messages.DeletedMessage)
	return p.tg.EditMessageText(ctx, event.ChatID, event.MessageID, text, mainMenuKeyboard(locale))
}

func (p *Processor) notePromptCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	id, err := parseCallbackID(event.Data, "note:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.InvalidCallbackLinkID)
	}
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
		return err
	}
	return p.tg.SendMessage(ctx, event.ChatID, fmt.Sprintf(messages.NotePromptFormat, id), mainMenuKeyboard(locale))
}

func (p *Processor) reminderPromptCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	id, err := parseCallbackID(event.Data, "remind:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.InvalidCallbackLinkID)
	}
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
		return err
	}
	return p.tg.SendMessage(ctx, event.ChatID, fmt.Sprintf(messages.ReminderPromptFormat, id, id), mainMenuKeyboard(locale))
}

func (p *Processor) groupPromptCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	locale := p.locale(ctx, meta)
	messages := tr(locale)
	id, err := parseCallbackID(event.Data, "group:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, messages.InvalidCallbackLinkID)
	}
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
		return err
	}
	return p.tg.SendMessage(ctx, event.ChatID, fmt.Sprintf(messages.GroupPromptFormat, id), mainMenuKeyboard(locale))
}

func parseCallbackID(data string, prefix string) (int64, error) {
	raw := strings.TrimPrefix(data, prefix)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, e.Wrap("invalid callback id", err)
	}
	return id, nil
}
