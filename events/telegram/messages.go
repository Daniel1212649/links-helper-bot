package telegram

import (
	"fmt"

	"github.com/Daniel1212649/LinksHelperBot/storage"
)

type messages struct {
	Help                  string
	Hello                 string
	UnknownCommand        string
	EmptyMessage          string
	NoSavedPages          string
	Saved                 string
	AlreadyExists         string
	InvalidURL            string
	EmptyList             string
	LatestLinksTitle      string
	SearchResultsTitle    string
	SearchUsage           string
	SearchPrompt          string
	NothingFound          string
	SavePrompt            string
	DeleteUsage           string
	DeletePrompt          string
	InvalidLinkID         string
	Deleted               string
	NoteUsage             string
	NotePromptFormat      string
	NoteSaved             string
	ReminderUsage         string
	ReminderPromptFormat  string
	InvalidReminderDate   string
	ReminderSavedFormat   string
	ReminderMessageTitle  string
	StatsFormat           string
	InvalidCallbackLinkID string
	CouldNotMarkRead      string
	MarkedRead            string
	MarkedReadMessage     string
	CouldNotDelete        string
	DeletedCallback       string
	DeletedMessage        string
	UnknownAction         string
	ChooseLanguage        string
	LanguageUpdated       string
}

func tr(locale string) messages {
	if storage.NormalizeLocale(locale) == storage.LocaleEN {
		return enMessages
	}
	return ruMessages
}

func localeFromLanguageCode(languageCode string) string {
	return storage.NormalizeLocale(languageCode)
}

func formatStatsMessage(locale string, stats storage.Stats) string {
	return fmt.Sprintf(tr(locale).StatsFormat, stats.Total, stats.Unread, stats.Read)
}

var enMessages = messages{
	Help: `LinksHelperBot saves links and helps you return to them later.

Commands:
/save <url> [note] [--remind <date>] - save a link
/rnd - get a random unread link
/list - show latest saved links
/search <text> - search by URL or title
/delete <id> - delete a link by ID
/stats - show your link stats
/lang [ru|en] - choose interface language
/note <id> <text> - add or update a note
/remind <id> <date> - remind about a link
/help - show this help

You can also send any http/https URL without a command.

Accepted reminder date formats:
2026-07-01 09:30
2026-07-01
01.07.2026 09:30
01.07.2026
All reminder times use Moscow time (Europe/Moscow).

Save with note and reminder:
/save https://example.com Useful article --remind 2026-07-01 09:30

Use the buttons below for quick actions. After /rnd, choose Read, Delete, or Another.`,
	UnknownCommand:        "Unknown command. Send /help or use the buttons below.",
	EmptyMessage:          "Please send a command, a link, or use the buttons below.",
	NoSavedPages:          "You have no unread saved links.",
	Saved:                 "Saved.",
	AlreadyExists:         "This link is already in your list.",
	InvalidURL:            "I can save only valid http/https links.",
	EmptyList:             "Your list is empty.",
	LatestLinksTitle:      "Latest links:",
	SearchResultsTitle:    "Search results:",
	SearchUsage:           "Usage: /search <text>",
	SearchPrompt:          "Send /search <text> or tap 🔍 Search and then type your query.",
	NothingFound:          "Nothing found.",
	SavePrompt:            "Send a link or use /save <url> [note] [--remind <date>].",
	DeleteUsage:           "Usage: /delete <id>",
	DeletePrompt:          "Send /delete <id> or tap 🗑 on a link in /list.",
	InvalidLinkID:         "Link ID must be a positive number.",
	Deleted:               "Deleted.",
	NoteUsage:             "Usage: /note <id> <text>",
	NotePromptFormat:      "Send /note %d <text> to add a note.",
	NoteSaved:             "Note saved.",
	ReminderUsage:         "Usage: /remind <id> <date>. Example: /remind 12 2026-07-01 09:30",
	ReminderPromptFormat:  "Send /remind %d <date>. Example: /remind %d 2026-07-01 09:30",
	InvalidReminderDate:   "I understand dates like 2026-07-01 09:30, 2026-07-01, 01.07.2026 09:30, 01.07.2026.",
	ReminderSavedFormat:   "Reminder set for %s.",
	ReminderMessageTitle:  "Reminder:",
	StatsFormat:           "Stats:\nTotal: %d\nUnread: %d\nRead: %d",
	InvalidCallbackLinkID: "Invalid link ID",
	CouldNotMarkRead:      "Could not mark as read",
	MarkedRead:            "Marked as read",
	MarkedReadMessage:     "✅ Marked as read.",
	CouldNotDelete:        "Could not delete link",
	DeletedCallback:       "Deleted",
	DeletedMessage:        "🗑 Deleted.",
	UnknownAction:         "Unknown action",
	ChooseLanguage:        "Choose language:",
	LanguageUpdated:       "Language updated.",
}

var ruMessages = messages{
	Help: `LinksHelperBot сохраняет ссылки и помогает вернуться к ним позже.

Команды:
/save <url> [заметка] [--remind <дата>] - сохранить ссылку
/rnd - случайная непрочитанная ссылка
/list - последние сохранённые ссылки
/search <text> - поиск по URL или названию
/delete <id> - удалить ссылку по ID
/stats - статистика ссылок
/lang [ru|en] - выбрать язык интерфейса
/note <id> <текст> - добавить или обновить заметку
/remind <id> <дата> - напомнить о ссылке
/help - показать справку

Также можно просто отправить любую http/https ссылку без команды.

Форматы дат для напоминаний:
2026-07-01 09:30
2026-07-01
01.07.2026 09:30
01.07.2026
Все напоминания считаются по московскому времени (Europe/Moscow).

Сохранить сразу с заметкой и напоминанием:
/save https://example.com Полезная статья --remind 2026-07-01 09:30

Используй кнопки ниже для быстрых действий. После /rnd выбери Прочитано, Удалить или Ещё.`,
	UnknownCommand:        "Неизвестная команда. Отправь /help или используй кнопки ниже.",
	EmptyMessage:          "Отправь команду, ссылку или используй кнопки ниже.",
	NoSavedPages:          "У тебя нет непрочитанных сохранённых ссылок.",
	Saved:                 "Сохранено.",
	AlreadyExists:         "Эта ссылка уже есть в твоём списке.",
	InvalidURL:            "Я могу сохранить только корректные http/https ссылки.",
	EmptyList:             "Твой список пуст.",
	LatestLinksTitle:      "Последние ссылки:",
	SearchResultsTitle:    "Результаты поиска:",
	SearchUsage:           "Использование: /search <текст>",
	SearchPrompt:          "Отправь /search <текст>, чтобы найти ссылку по URL или названию.",
	NothingFound:          "Ничего не найдено.",
	SavePrompt:            "Отправь ссылку или используй /save <url> [заметка] [--remind <дата>].",
	DeleteUsage:           "Использование: /delete <id>",
	DeletePrompt:          "Отправь /delete <id> или нажми 🗑 рядом со ссылкой в /list.",
	InvalidLinkID:         "ID ссылки должен быть положительным числом.",
	Deleted:               "Удалено.",
	NoteUsage:             "Использование: /note <id> <текст>",
	NotePromptFormat:      "Отправь /note %d <текст>, чтобы добавить заметку.",
	NoteSaved:             "Заметка сохранена.",
	ReminderUsage:         "Использование: /remind <id> <дата>. Пример: /remind 12 2026-07-01 09:30",
	ReminderPromptFormat:  "Отправь /remind %d <дата>. Пример: /remind %d 2026-07-01 09:30",
	InvalidReminderDate:   "Я понимаю даты вида 2026-07-01 09:30, 2026-07-01, 01.07.2026 09:30, 01.07.2026.",
	ReminderSavedFormat:   "Напоминание установлено на %s.",
	ReminderMessageTitle:  "Напоминание:",
	StatsFormat:           "Статистика:\nВсего: %d\nНепрочитано: %d\nПрочитано: %d",
	InvalidCallbackLinkID: "Некорректный ID ссылки",
	CouldNotMarkRead:      "Не удалось пометить прочитанной",
	MarkedRead:            "Помечено прочитанной",
	MarkedReadMessage:     "✅ Помечено прочитанной.",
	CouldNotDelete:        "Не удалось удалить ссылку",
	DeletedCallback:       "Удалено",
	DeletedMessage:        "🗑 Удалено.",
	UnknownAction:         "Неизвестное действие",
	ChooseLanguage:        "Выбери язык:",
	LanguageUpdated:       "Язык обновлён.",
}

func init() {
	enMessages.Hello = "Hi! Send me a link and I will save it for later.\n\n" + enMessages.Help
	ruMessages.Hello = "Привет! Отправь мне ссылку, и я сохраню её на потом.\n\n" + ruMessages.Help
}
