package telegram

import (
	"fmt"

	"github.com/Daniel1212649/LinksHelperBot/internal/storage"
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
	GroupLinksTitleFormat string
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
	GroupUsage            string
	GroupPromptFormat     string
	GroupSaved            string
	GroupsTitle           string
	NoGroups              string
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

📋 Commands:
💾 /save <url> [note] [#group] [@date] — save a link
🎲 /rnd — random unread link
📋 /list [group] — latest links, optionally by group
🔍 /search <text> — search by URL, title, note or group
🗑 /delete <id> — delete by ID
📊 /stats — counters
🌐 /lang [ru|en] — language
📝 /note <id> <text> — add or update a note
📁 /group <id> <group> — set link group
📁 /groups — list groups
⏰ /remind <id> <date> — set reminder
📖 /help — this help

🔗 You can also send any http/https URL without a command.

⏰ Reminder date formats (Moscow time, Europe/Moscow):
• 2026-07-01 09:30
• 2026-07-01
• 01.07.2026 09:30
• 01.07.2026

💡 Save examples:
/save <url> <note text>
/save <url> <note text> @<date>
/save <url> <note text> #<group>
/save <url> <note text> #<group> @<date>

👇 Use the buttons below. After /rnd: ✅ Read, 🗑 Delete, 🎲 Another.`,
	UnknownCommand:        "Unknown command. Send /help or use the buttons below.",
	EmptyMessage:          "Please send a command, a link, or use the buttons below.",
	NoSavedPages:          "You have no unread saved links.",
	Saved:                 "Saved.",
	AlreadyExists:         "This link is already in your list.",
	InvalidURL:            "I can save only valid http/https links.",
	EmptyList:             "Your list is empty.",
	LatestLinksTitle:      "Latest links:",
	GroupLinksTitleFormat: "Links in group %q:",
	SearchResultsTitle:    "Search results:",
	SearchUsage:           "Usage: /search <text>",
	SearchPrompt:          "Send /search <text> or tap 🔍 Search and then type your query.",
	NothingFound:          "Nothing found.",
	SavePrompt:            "Send a link or use /save <url> <note> [#group] [@date].",
	DeleteUsage:           "Usage: /delete <id>",
	DeletePrompt:          "Send /delete <id> or tap 🗑 on a link in /list.",
	InvalidLinkID:         "Link ID must be a positive number.",
	Deleted:               "Deleted.",
	NoteUsage:             "Usage: /note <id> <text>",
	NotePromptFormat:      "Send /note %d <text> to add a note.",
	NoteSaved:             "Note saved.",
	GroupUsage:            "Usage: /group <id> <group>",
	GroupPromptFormat:     "Send /group %d <group> to add this link to a group.",
	GroupSaved:            "Group saved.",
	GroupsTitle:           "Groups:",
	NoGroups:              "You do not have any groups yet.",
	ReminderUsage:         "Usage: /remind <id> <date>. Example: /remind <id> <date>",
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

📋 Команды:
💾 /save <url> [заметка] [#группа] [@дата] — сохранить ссылку
🎲 /rnd — случайная непрочитанная ссылка
📋 /list [группа] — последние ссылки, можно по группе
🔍 /search <текст> — поиск по URL, названию, заметке или группе
🗑 /delete <id> — удалить по ID
📊 /stats — статистика
🌐 /lang [ru|en] — язык интерфейса
📝 /note <id> <текст> — добавить или обновить заметку
📁 /group <id> <группа> — задать группу ссылки
📁 /groups — список групп
⏰ /remind <id> <дата> — напоминание
📖 /help — эта справка

🔗 Можно просто отправить http/https ссылку без команды.

⏰ Форматы даты (московское время, Europe/Moscow):
• 2026-07-01 09:30
• 2026-07-01
• 01.07.2026 09:30
• 01.07.2026

💡 Примеры сохранения:
/save <url> <текст заметки>
/save <url> <текст заметки> @<дата>
/save <url> <текст заметки> #<группа>
/save <url> <текст заметки> #<группа> @<дата>

👇 Используй кнопки ниже. После /rnd: ✅ Прочитано, 🗑 Удалить, 🎲 Ещё.`,
	UnknownCommand:        "Неизвестная команда. Отправь /help или используй кнопки ниже.",
	EmptyMessage:          "Отправь команду, ссылку или используй кнопки ниже.",
	NoSavedPages:          "У тебя нет непрочитанных сохранённых ссылок.",
	Saved:                 "Сохранено.",
	AlreadyExists:         "Эта ссылка уже есть в твоём списке.",
	InvalidURL:            "Я могу сохранить только корректные http/https ссылки.",
	EmptyList:             "Твой список пуст.",
	LatestLinksTitle:      "Последние ссылки:",
	GroupLinksTitleFormat: "Ссылки в группе %q:",
	SearchResultsTitle:    "Результаты поиска:",
	SearchUsage:           "Использование: /search <текст>",
	SearchPrompt:          "Отправь /search <текст>, чтобы найти ссылку по URL или названию.",
	NothingFound:          "Ничего не найдено.",
	SavePrompt:            "Отправь ссылку или /save <url> <заметка> [#группа] [@дата].",
	DeleteUsage:           "Использование: /delete <id>",
	DeletePrompt:          "Отправь /delete <id> или нажми 🗑 рядом со ссылкой в /list.",
	InvalidLinkID:         "ID ссылки должен быть положительным числом.",
	Deleted:               "Удалено.",
	NoteUsage:             "Использование: /note <id> <текст>",
	NotePromptFormat:      "Отправь /note %d <текст>, чтобы добавить заметку.",
	NoteSaved:             "Заметка сохранена.",
	GroupUsage:            "Использование: /group <id> <группа>",
	GroupPromptFormat:     "Отправь /group %d <группа>, чтобы добавить ссылку в группу.",
	GroupSaved:            "Группа сохранена.",
	GroupsTitle:           "Группы:",
	NoGroups:              "У тебя пока нет групп.",
	ReminderUsage:         "Использование: /remind <id> <дата>. Пример: /remind <id> <дата>",
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
	enMessages.Hello = "👋 Hi! Send me a link and I will save it for later.\n\n" + enMessages.Help
	ruMessages.Hello = "👋 Привет! Отправь ссылку — сохраню её на потом.\n\n" + ruMessages.Help
}
