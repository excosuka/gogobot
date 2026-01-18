package telegram

const msgHelp = `I can save and keep you pages. Also I can offer you them to read.

In order to save the page, just send me link to it.

In order to watch a random page from your list, send me command /peek

In order to get a random page from your list, send me command /pick.
Caution! After that,this page will be removed from your list! 

Also You can get into main menu with command /menu

If You need help you can use command /help
`

const msgHello = "Hi there!  \n\n" + msgHelp

const (
	msgUnknownCommand = "Unknown command"
	msgNoSavedPages   = "You have no saved pages"
	msgSaved          = "Saved!"
	msgAlreadyExists  = "You have already have this page in your list"
	msgCount          = "You have this count of pages: "
	msgList           = "You already saved these pages!!!"
	msgQuery          = "Filtered pages!!!"
)
