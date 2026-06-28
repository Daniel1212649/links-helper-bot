package telegram

import (
	"context"
	"errors"

	tgclient "github.com/Daniel1212649/LinksHelperBot/clients/telegram"
	"github.com/Daniel1212649/LinksHelperBot/events"
	"github.com/Daniel1212649/LinksHelperBot/lib/e"
	"github.com/Daniel1212649/LinksHelperBot/storage"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrUnknownMetaType  = errors.New("unknown meta type")
)

type Meta struct {
	ChatID       int64
	TelegramID   int64
	Username     string
	LanguageCode string
}

type Processor struct {
	tg      *tgclient.Client
	offset  int
	storage storage.Storage
}

func New(client *tgclient.Client, storage storage.Storage) *Processor {
	return &Processor{
		tg:      client,
		storage: storage,
	}
}

func (p *Processor) Fetch(ctx context.Context, limit int) ([]events.Event, error) {
	updates, err := p.tg.Updates(ctx, p.offset, limit)
	if err != nil {
		return nil, e.Wrap("can't get events", err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))
	for _, update := range updates {
		if event, ok := mapUpdate(update); ok {
			res = append(res, event)
		}
	}
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
	case events.CallbackQuery:
		return p.processCallback(ctx, event.Meta.(CallbackMeta))
	default:
		return e.Wrap("can't process event", ErrUnknownEventType)
	}
}

func (p *Processor) processMessage(ctx context.Context, event events.Event) error {
	meta, err := meta(event)
	if err != nil {
		return e.Wrap("can't process message", err)
	}

	if err := p.doCmd(ctx, event.Text, meta); err != nil {
		return e.Wrap("can't process message", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, e.Wrap("can't get meta", ErrUnknownMetaType)
	}
	return res, nil
}

func mapUpdate(update tgclient.Update) (events.Event, bool) {
	if update.CallbackQuery != nil {
		cb := update.CallbackQuery
		if cb.Message == nil {
			return events.Event{}, false
		}

		return events.Event{
			Type: events.CallbackQuery,
			Meta: CallbackMeta{
				CallbackQueryID: cb.ID,
				ChatID:          cb.Message.Chat.ID,
				MessageID:       cb.Message.MessageID,
				TelegramID:      cb.From.ID,
				Username:        cb.From.Username,
				LanguageCode:    cb.From.LanguageCode,
				Data:            cb.Data,
			},
		}, true
	}

	if update.Message == nil {
		return events.Event{}, false
	}

	return events.Event{
		Type: events.Message,
		Text: update.Message.Text,
		Meta: Meta{
			ChatID:       update.Message.Chat.ID,
			TelegramID:   update.Message.From.ID,
			Username:     update.Message.From.Username,
			LanguageCode: update.Message.From.LanguageCode,
		},
	}, true
}

func userFromMeta(meta Meta) storage.User {
	return storage.User{
		TelegramID: meta.TelegramID,
		ChatID:     meta.ChatID,
		Username:   meta.Username,
		Locale:     localeFromLanguageCode(meta.LanguageCode),
	}
}
