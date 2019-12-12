package service

import (
	"context"
	"github.com/alexedwards/scs/v2"
	"html/template"
	"log"
	"net/http"
	"otus-hiload/src/constants"
	"otus-hiload/src/repository"
)

type baseService struct {
	userRepository repository.IUserRepository
	sessionManager *scs.SessionManager
}

func (s *baseService) logError(msg string, err error) {
	if err != nil {
		log.Printf(msg+": %s", err.Error())
	}
}

func (s *baseService) renderForm(w http.ResponseWriter, form string, error error) {
	params := make(map[string]string)
	if error != nil {
		s.logError("renderForm "+form, error)
		params["error"] = error.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	t, _ := template.ParseFiles("templates/" + form + ".html")
	err := t.Execute(w, params)
	s.logError(form+" template execute error: %s", err)
}

func (s *baseService) renderFormParams(w http.ResponseWriter, form string, params interface{}) {
	t, _ := template.ParseFiles("templates/" + form + ".html")
	err := t.Execute(w, params)
	s.logError(form+" template execute error: %s", err)
}

func (s *baseService) getUserFromContext(ctx context.Context) (*repository.User, error) {
	userId := (ctx.Value(constants.CtxUserId)).(int64)
	return s.userRepository.Get(userId)
}
