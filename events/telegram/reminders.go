package telegram

import (
	"context"
	"log"
	"time"

	tgclient "github.com/Daniel1212649/LinksHelperBot/clients/telegram"
	"github.com/Daniel1212649/LinksHelperBot/storage"
)

func StartReminderScheduler(ctx context.Context, client *tgclient.Client, store storage.Storage, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			if err := sendDueReminders(ctx, client, store); err != nil {
				log.Printf("can't send due reminders: %v", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func sendDueReminders(ctx context.Context, client *tgclient.Client, store storage.Storage) error {
	reminders, err := store.DueReminders(ctx, time.Now(), 25)
	if err != nil {
		return err
	}

	for _, reminder := range reminders {
		locale := storage.NormalizeLocale(reminder.User.Locale)
		text := tr(locale).ReminderMessageTitle + "\n" + formatPage(&reminder.Page)
		if err := client.SendMessage(ctx, reminder.User.ChatID, text, linkActionKeyboard(locale, reminder.Page.ID)); err != nil {
			return err
		}
		if err := store.MarkReminded(ctx, reminder.Page.ID); err != nil {
			return err
		}
	}

	return nil
}
