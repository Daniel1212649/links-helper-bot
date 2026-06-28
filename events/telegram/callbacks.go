package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Daniel1212649/LinksHelperBot/lib/e"
)

type CallbackMeta struct {
	CallbackQueryID string
	ChatID          int64
	MessageID       int
	TelegramID      int64
	Username        string
	Data            string
}

func (p *Processor) processCallback(ctx context.Context, event CallbackMeta) error {
	meta := Meta{
		ChatID:     event.ChatID,
		TelegramID: event.TelegramID,
		Username:   event.Username,
	}

	switch event.Data {
	case cbCmdStart:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendHello(ctx, meta.ChatID)
		})
	case cbCmdHelp:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendHelp(ctx, meta.ChatID)
		})
	case cbCmdSave:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, msgSavePrompt, mainMenuKeyboard())
		})
	case cbCmdRnd:
		if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
			return err
		}
		return p.sendRandomPage(ctx, meta, event.MessageID)
	case cbCmdList:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendList(ctx, meta)
		})
	case cbCmdStats:
		return p.answerAndSend(ctx, event, func() error {
			return p.sendStats(ctx, meta)
		})
	case cbCmdSearch:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, msgSearchPrompt, mainMenuKeyboard())
		})
	case cbCmdDelete:
		return p.answerAndSend(ctx, event, func() error {
			return p.tg.SendMessage(ctx, meta.ChatID, msgDeletePrompt, mainMenuKeyboard())
		})
	default:
		if strings.HasPrefix(event.Data, "read:") {
			return p.markReadCallback(ctx, event, meta)
		}
		if strings.HasPrefix(event.Data, "del:") {
			return p.deleteCallback(ctx, event, meta)
		}
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Unknown action")
	}
}

func (p *Processor) answerAndSend(ctx context.Context, event CallbackMeta, action func() error) error {
	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, ""); err != nil {
		return err
	}
	return action()
}

func (p *Processor) markReadCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	id, err := parseCallbackID(event.Data, "read:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Invalid link ID")
	}

	if err := p.storage.MarkRead(ctx, userFromMeta(meta), id); err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Could not mark as read")
	}

	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Marked as read"); err != nil {
		return err
	}

	if event.MessageID == 0 {
		return nil
	}

	text := fmt.Sprintf("#%d\n✅ Marked as read.", id)
	return p.tg.EditMessageText(ctx, event.ChatID, event.MessageID, text, mainMenuKeyboard())
}

func (p *Processor) deleteCallback(ctx context.Context, event CallbackMeta, meta Meta) error {
	id, err := parseCallbackID(event.Data, "del:")
	if err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Invalid link ID")
	}

	if err := p.storage.Remove(ctx, userFromMeta(meta), id); err != nil {
		return p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Could not delete link")
	}

	if err := p.tg.AnswerCallbackQuery(ctx, event.CallbackQueryID, "Deleted"); err != nil {
		return err
	}

	if event.MessageID == 0 {
		return p.tg.SendMessage(ctx, event.ChatID, fmt.Sprintf("#%d deleted.", id), mainMenuKeyboard())
	}

	text := fmt.Sprintf("#%d\n🗑 Deleted.", id)
	return p.tg.EditMessageText(ctx, event.ChatID, event.MessageID, text, mainMenuKeyboard())
}

func parseCallbackID(data string, prefix string) (int64, error) {
	raw := strings.TrimPrefix(data, prefix)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, e.Wrap("invalid callback id", err)
	}
	return id, nil
}
