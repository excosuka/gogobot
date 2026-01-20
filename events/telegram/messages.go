package telegram

const msgHelp = `I can save and keep you pages. Also I can offer you them to read.

In order to save the page, just send me link to it with hashtags to search it in future.

In order to watch a random page from your list, send me command /peek

In order to get a random page from your list, send me command /pick.
Caution! After that,this page will be removed from your list! 

You can get into main menu with command /menu

You can use /list to get all articles from your storage! (In Future will be added pagination on 10 or more pages)

You can use /search #YourHashTag...  to have page with your hash tag. 

`

const msgHello = "Hi there!👋🏻  \n\n" + "This bot can help You to save all articles those You want to read, by don`t got enough time in moment.\n Just drop the link on article and add hashtags to search from storage.\nAlways you can send me command: '/help' to get advices."

const (
	msgUnknownCommand     = "Unknown command"
	msgNoSavedPages       = "You have no saved pages!"
	msgSaved              = "Saved!"
	msgAlreadyExists      = "You have already have this page in your list"
	msgCount              = "You have this count of pages: "
	msgList               = "You already saved these pages: "
	msgQuery              = "Filtered pages!!!"
	msgEmptyHashTags      = "You didn't specify hashtags!"
	msgEmptyFilteredPages = "You didn`t have pages with current hashtags!"
	msgToSpecifyHashTags  = "Specify hashtags to search by thems!"
)
