package constants

const (
	LoginPath        = "/login"
	LogoutPath       = "/logout"
	RegPath          = "/reg"
	MePath           = "/me"
	MeEditPath       = "/me/edit"
	UserPath         = "/user/{id:[0-9]+}"
	SearchPath       = "/search"
	ChatsPath        = "/chats"
	ChatFormPath     = "/chats/{userId:[0-9]+}"
	FirstMessagePath = "/chats/{userId:[0-9]+}/firstmessage"
	MessagePath      = "/chats/{chatId:[0-9]+}/messages"
	LoadMessagePath  = "/chats/{chatId:[0-9]+}/messages/{minId:[0-9]+}"
	RootPath         = "/"

	CtxUserId        = "userID"
	CtxAuthenticated = "authenticated"
)
