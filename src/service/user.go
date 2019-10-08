package service

import (
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"otus-hiload/src/constants"
	"otus-hiload/src/repository"
)

type userService struct {
	UserRepository repository.IUserRepository
	store          *sessions.CookieStore
	storageDir     string
}

type IUserService interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	RegHandler(w http.ResponseWriter, r *http.Request)
	IPageService
}

func NewUserService(repository repository.IUserRepository, store *sessions.CookieStore, storageDir string) IUserService {
	return &userService{UserRepository: repository, store: store, storageDir: storageDir}
}

func (s *userService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.renderForm(w, "login", nil)
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		s.logError("login form parse error: %s", err)

		login := r.FormValue("login")
		password := r.FormValue("password")

		if len(login) == 0 || len(password) == 0 {
			s.renderForm(w, "login", errors.New("все поля должны быть заполнены"))
			return
		}

		user, err := s.UserRepository.FindByLoginAndPassword(login, password)
		if err != nil {
			s.logError("UserRepository.FindByLoginAndPassword error: %s", err)
			s.renderForm(w, "login", errors.New("комбинация логин/пароль не существует"))
			return
		}
		//
		session, _ := s.store.Get(r, constants.CookieName)
		// Set user as authenticated
		session.Values["authenticated"] = true
		session.Values["user_id"] = user.Id
		err = session.Save(r, w)
		s.logError("store.Save error: %s", err)
		if err != nil {
			s.renderForm(w, "login", errors.New("внутренняя ошибка сервера"))
			return
		}
		//
		http.Redirect(w, r, constants.MePath, http.StatusFound)
	}
}

func (s *userService) RegHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.renderForm(w, "reg", nil)
	}

	if r.Method == "POST" {
		err := r.ParseForm()
		s.logError("reg form parse error: %s", err)

		login := r.FormValue("login")
		name := r.FormValue("name")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password2")

		params := make(map[string]string)
		params["login"] = login
		params["name"] = name

		if len(login) == 0 || len(name) == 0 || len(password) == 0 || len(passwordConfirm) == 0 {
			params["error"] = "все поля должны быть заполнены"
			s.renderFormParams(w, "reg", params)
			return
		}

		if password != passwordConfirm {
			params["error"] = "пароль должен быть равен подтверждению"
			s.renderFormParams(w, "reg", params)
			return
		}

		if s.UserRepository.IsLoginExist(login) {
			params["error"] = fmt.Sprintf("имя пользователя [%s] уже занято", login)
			s.renderFormParams(w, "reg", params)
			return
		}

		user := new(repository.User)
		user.Login = login
		user.Name = name
		user.Password = password
		err = s.UserRepository.Create(user)
		s.logError("UserRepository.Create error: %s", err)
		if err != nil {
			params["error"] = "внутренняя ошибка сервера"
			s.renderFormParams(w, "reg", params)
			return
		}
		//
		session, _ := s.store.Get(r, "auth")
		session.Values["authenticated"] = true
		session.Values["user_id"] = user.Id
		err = session.Save(r, w)
		s.logError("store.Save error: %s", err)
		http.Redirect(w, r, constants.MePath, http.StatusFound)
	}
}

func (s *userService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.store.Get(r, constants.CookieName)
	session.Values["authenticated"] = false
	session.Values["user_id"] = nil
	err := session.Save(r, w)
	if err != nil {
		log.Printf("logout session save error: %s", err.Error())
	}
	http.Redirect(w, r, constants.RootPath, http.StatusFound)
}
