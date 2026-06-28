package telegram

import (
	"fmt"

	tgclient "github.com/Daniel1212649/LinksHelperBot/clients/telegram"
	"github.com/Daniel1212649/LinksHelperBot/storage"
)

const (
	cbCmdStart  = "cmd:start"
	cbCmdHelp   = "cmd:help"
	cbCmdSave   = "cmd:save"
	cbCmdRnd    = "cmd:rnd"
	cbCmdList   = "cmd:list"
	cbCmdStats  = "cmd:stats"
	cbCmdSearch = "cmd:search"
	cbCmdDelete = "cmd:delete"
	cbCmdNote   = "cmd:note"
	cbCmdRemind = "cmd:remind"
	cbCmdLang   = "cmd:lang"
	cbLangRU    = "lang:ru"
	cbLangEN    = "lang:en"
)

func cbRead(id int64) string {
	return fmt.Sprintf("read:%d", id)
}

func cbDelete(id int64) string {
	return fmt.Sprintf("del:%d", id)
}

func cbNote(id int64) string {
	return fmt.Sprintf("note:%d", id)
}

func cbRemind(id int64) string {
	return fmt.Sprintf("remind:%d", id)
}

func mainMenuKeyboard(locale string) *tgclient.InlineKeyboardMarkup {
	return &tgclient.InlineKeyboardMarkup{
		InlineKeyboard: mainMenuRows(locale),
	}
}

func languageKeyboard(locale string) *tgclient.InlineKeyboardMarkup {
	return &tgclient.InlineKeyboardMarkup{
		InlineKeyboard: append([][]tgclient.InlineKeyboardButton{
			{
				{Text: "🇷🇺 Русский", CallbackData: cbLangRU},
				{Text: "🇬🇧 English", CallbackData: cbLangEN},
			},
		}, mainMenuRows(locale)...),
	}
}

func mainMenuRows(locale string) [][]tgclient.InlineKeyboardButton {
	if storage.NormalizeLocale(locale) == storage.LocaleEN {
		return [][]tgclient.InlineKeyboardButton{
			{
				{Text: "👋 Start", CallbackData: cbCmdStart},
				{Text: "📖 Help", CallbackData: cbCmdHelp},
			},
			{
				{Text: "💾 Save", CallbackData: cbCmdSave},
				{Text: "🎲 Random", CallbackData: cbCmdRnd},
			},
			{
				{Text: "📝 Note", CallbackData: cbCmdNote},
				{Text: "⏰ Reminder", CallbackData: cbCmdRemind},
			},
			{
				{Text: "📋 List", CallbackData: cbCmdList},
				{Text: "📊 Stats", CallbackData: cbCmdStats},
			},
			{
				{Text: "🔍 Search", CallbackData: cbCmdSearch},
				{Text: "🗑 Delete", CallbackData: cbCmdDelete},
			},
			{
				{Text: "🌐 Language", CallbackData: cbCmdLang},
			},
		}
	}

	return [][]tgclient.InlineKeyboardButton{
		{
			{Text: "👋 Старт", CallbackData: cbCmdStart},
			{Text: "📖 Справка", CallbackData: cbCmdHelp},
		},
		{
			{Text: "💾 Сохранить", CallbackData: cbCmdSave},
			{Text: "🎲 Случайная", CallbackData: cbCmdRnd},
		},
		{
			{Text: "📝 Заметка", CallbackData: cbCmdNote},
			{Text: "⏰ Напомнить", CallbackData: cbCmdRemind},
		},
		{
			{Text: "📋 Список", CallbackData: cbCmdList},
			{Text: "📊 Статистика", CallbackData: cbCmdStats},
		},
		{
			{Text: "🔍 Поиск", CallbackData: cbCmdSearch},
			{Text: "🗑 Удалить", CallbackData: cbCmdDelete},
		},
		{
			{Text: "🌐 Язык", CallbackData: cbCmdLang},
		},
	}
}

func linkActionKeyboard(locale string, pageID int64) *tgclient.InlineKeyboardMarkup {
	rows := mainMenuRows(locale)
	actionRow := []tgclient.InlineKeyboardButton{
		{Text: "✅ Прочитано", CallbackData: cbRead(pageID)},
		{Text: "🗑 Удалить", CallbackData: cbDelete(pageID)},
		{Text: "🎲 Ещё", CallbackData: cbCmdRnd},
	}
	detailsRow := []tgclient.InlineKeyboardButton{
		{Text: "📝 Заметка", CallbackData: cbNote(pageID)},
		{Text: "⏰ Напомнить", CallbackData: cbRemind(pageID)},
	}
	if storage.NormalizeLocale(locale) == storage.LocaleEN {
		actionRow = []tgclient.InlineKeyboardButton{
			{Text: "✅ Read", CallbackData: cbRead(pageID)},
			{Text: "🗑 Delete", CallbackData: cbDelete(pageID)},
			{Text: "🎲 Another", CallbackData: cbCmdRnd},
		}
		detailsRow = []tgclient.InlineKeyboardButton{
			{Text: "📝 Note", CallbackData: cbNote(pageID)},
			{Text: "⏰ Remind", CallbackData: cbRemind(pageID)},
		}
	}

	rows = append([][]tgclient.InlineKeyboardButton{
		actionRow,
		detailsRow,
	}, rows...)

	return &tgclient.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func listActionKeyboard(locale string, pages []storage.Page) *tgclient.InlineKeyboardMarkup {
	rows := make([][]tgclient.InlineKeyboardButton, 0, len(pages)/2+len(mainMenuRows(locale)))

	for i := 0; i < len(pages); i += 2 {
		row := []tgclient.InlineKeyboardButton{
			{Text: fmt.Sprintf("🗑 #%d", pages[i].ID), CallbackData: cbDelete(pages[i].ID)},
		}
		if i+1 < len(pages) {
			row = append(row, tgclient.InlineKeyboardButton{
				Text:         fmt.Sprintf("🗑 #%d", pages[i+1].ID),
				CallbackData: cbDelete(pages[i+1].ID),
			})
		}
		rows = append(rows, row)
	}

	rows = append(rows, mainMenuRows(locale)...)
	return &tgclient.InlineKeyboardMarkup{InlineKeyboard: rows}
}
