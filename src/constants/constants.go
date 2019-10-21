package constants

const (
	LoginPath  = "/login"
	LogoutPath = "/logout"
	RegPath    = "/reg"
	MePath     = "/me"
	MeEditPath = "/me/edit"
	UserPath   = "/user/{id:[0-9]+}"
	SearchPath = "/search/{name}/{lastname}"
	RootPath   = "/"

	CtxUserId        = "userID"
	CtxAuthenticated = "authenticated"
)
