package telegram

const msgHelp = `LinksHelperBot saves links and helps you return to them later.

Commands:
/save <url> - save a link
/rnd - get a random unread link and mark it as read
/list - show latest saved links
/search <text> - search by URL or title
/delete <id> - delete a link by ID
/stats - show your link stats
/help - show this help

You can also send any http/https URL without a command.`

const msgHello = "Hi! Send me a link and I will save it for later.\n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command. Send /help to see available commands."
	msgEmptyMessage   = "Please send a command or a link."
	msgNoSavedPages   = "You have no unread saved links."
	msgSaved          = "Saved."
	msgAlreadyExists  = "This link is already in your list."
	msgInvalidURL     = "I can save only valid http/https links."
)
