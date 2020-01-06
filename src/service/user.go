package service

import (
	"errors"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"net/http"
	"otus-hiload/src/constants"
	"otus-hiload/src/file_storage"
	"otus-hiload/src/repository"
)

type userService struct {
	UserRepository      repository.IUserRepository
	tarantoolRepository repository.ITarantoolRepository
	sessionManager      *scs.SessionManager
	storage             file_storage.IFileStorage
	searchPageSize      int
}

type IUserService interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	RegHandler(w http.ResponseWriter, r *http.Request)
	IPageService
}

func NewUserService(repository repository.IUserRepository, tRepository repository.ITarantoolRepository, sessionManager *scs.SessionManager,
	storage file_storage.IFileStorage) IUserService {
	return &userService{UserRepository: repository, tarantoolRepository: tRepository, sessionManager: sessionManager,
		storage: storage, searchPageSize: 1000}
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
		err = s.setAuthenticated(r.Context(), user)
		if err != nil {
			s.logError("sessionManager.RenewToken error: %s", err)
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
		lastName := r.FormValue("last_name")
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password2")

		params := make(map[string]string)
		params["login"] = login
		params["name"] = name
		params["last_name"] = lastName

		if len(login) == 0 || len(name) == 0 || len(lastName) == 0 || len(password) == 0 || len(passwordConfirm) == 0 {
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
			params["error"] = fmt.Sprintf("логин пользователя [%s] уже занят", login)
			s.renderFormParams(w, "reg", params)
			return
		}

		user := new(repository.User)
		user.Login = login
		user.Name = name
		user.LastName = lastName
		user.Password = password
		err = s.UserRepository.Create(user)
		s.logError("UserRepository.Create error: %s", err)
		if err != nil {
			params["error"] = "внутренняя ошибка сервера"
			s.renderFormParams(w, "reg", params)
			return
		}
		//
		err = s.setAuthenticated(r.Context(), user)
		s.logError("sessionManager.RenewToken error: %s", err)
		if err != nil {
			params["error"] = "внутренняя ошибка сервера"
			s.renderFormParams(w, "reg", params)
			return
		}
		http.Redirect(w, r, constants.MePath, http.StatusFound)
	}
}

func (s *userService) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	err := s.setUnauthenticated(r.Context())
	s.logError("setUnauthenticated", err)
	http.Redirect(w, r, constants.RootPath, http.StatusFound)
}
