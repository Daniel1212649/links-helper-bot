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
)

func cbRead(id int64) string {
	return fmt.Sprintf("read:%d", id)
}

func cbDelete(id int64) string {
	return fmt.Sprintf("del:%d", id)
}

func mainMenuKeyboard() *tgclient.InlineKeyboardMarkup {
	return &tgclient.InlineKeyboardMarkup{
		InlineKeyboard: mainMenuRows(),
	}
}

func mainMenuRows() [][]tgclient.InlineKeyboardButton {
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
			{Text: "📋 Список", CallbackData: cbCmdList},
			{Text: "📊 Статистика", CallbackData: cbCmdStats},
		},
		{
			{Text: "🔍 Поиск", CallbackData: cbCmdSearch},
			{Text: "🗑 Удалить", CallbackData: cbCmdDelete},
		},
	}
}

func linkActionKeyboard(pageID int64) *tgclient.InlineKeyboardMarkup {
	rows := mainMenuRows()
	rows = append([][]tgclient.InlineKeyboardButton{
		{
			{Text: "✅ Прочитано", CallbackData: cbRead(pageID)},
			{Text: "🗑 Удалить", CallbackData: cbDelete(pageID)},
			{Text: "🎲 Ещё", CallbackData: cbCmdRnd},
		},
	}, rows...)

	return &tgclient.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func listActionKeyboard(pages []storage.Page) *tgclient.InlineKeyboardMarkup {
	rows := make([][]tgclient.InlineKeyboardButton, 0, len(pages)/2+len(mainMenuRows()))

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

	rows = append(rows, mainMenuRows()...)
	return &tgclient.InlineKeyboardMarkup{InlineKeyboard: rows}
}
