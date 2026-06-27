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
	ChatID     int64
	TelegramID int64
	Username   string
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
		res = append(res, event(update))
	}
	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(ctx context.Context, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(ctx, event)
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

func event(update tgclient.Update) events.Event {
	updateType := fetchType(update)

	res := events.Event{
		Type: updateType,
		Text: fetchText(update),
	}

	if updateType == events.Message {
		res.Meta = Meta{
			ChatID:     update.Message.Chat.ID,
			TelegramID: update.Message.From.ID,
			Username:   update.Message.From.Username,
		}
	}

	return res
}

func fetchText(update tgclient.Update) string {
	if update.Message == nil {
		return ""
	}
	return update.Message.Text
}

func fetchType(update tgclient.Update) events.Type {
	if update.Message == nil {
		return events.Unknown
	}
	return events.Message
}

func userFromMeta(meta Meta) storage.User {
	return storage.User{
		TelegramID: meta.TelegramID,
		ChatID:     meta.ChatID,
		Username:   meta.Username,
	}
}
