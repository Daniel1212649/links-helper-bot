package telegram

const msgHelp = `LinksHelperBot saves links and helps you return to them later.

Commands:
/save <url> - save a link
/rnd - get a random unread link
/list - show latest saved links
/search <text> - search by URL or title
/delete <id> - delete a link by ID
/stats - show your link stats
/help - show this help

You can also send any http/https URL without a command.

Use the buttons below for quick actions. After /rnd, choose Read, Delete, or Another.`

const msgHello = "Hi! Send me a link and I will save it for later.\n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command. Send /help or use the buttons below."
	msgEmptyMessage   = "Please send a command, a link, or use the buttons below."
	msgNoSavedPages   = "You have no unread saved links."
	msgSaved          = "Saved."
	msgAlreadyExists  = "This link is already in your list."
	msgInvalidURL     = "I can save only valid http/https links."
	msgEmptyList      = "Your list is empty."
	msgSearchUsage    = "Usage: /search <text>"
	msgSearchPrompt   = "Send /search <text> or tap 🔍 Search and then type your query."
	msgNothingFound   = "Nothing found."
	msgSavePrompt     = "Send a link or use /save <url>."
	msgDeleteUsage    = "Usage: /delete <id>"
	msgDeletePrompt   = "Send /delete <id> or tap 🗑 on a link in /list."
	msgInvalidLinkID  = "Link ID must be a positive number."
	msgDeleted        = "Deleted."
)
