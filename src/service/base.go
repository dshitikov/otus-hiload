package service

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"otus-hiload/src/constants"
	"otus-hiload/src/repository"
)

func (s *userService) logError(msg string, err error) {
	if err != nil {
		log.Printf(msg+": %s", err.Error())
	}
}

func (s *userService) renderForm(w http.ResponseWriter, form string, error error) {
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

func (s *userService) renderFormParams(w http.ResponseWriter, form string, params interface{}) {
	t, _ := template.ParseFiles("templates/" + form + ".html")
	err := t.Execute(w, params)
	s.logError(form+" template execute error: %s", err)
}

func (s *userService) getUserFromContext(ctx context.Context) (*repository.User, error) {
	userId := (ctx.Value(constants.CtxUserId)).(int64)
	return s.UserRepository.Get(userId)
}

func (s *userService) setAuthenticated(ctx context.Context, user *repository.User) error {
	err := s.sessionManager.RenewToken(ctx)
	if err != nil {
		return err
	}
	s.sessionManager.Put(ctx, constants.CtxAuthenticated, true)
	s.sessionManager.Put(ctx, constants.CtxUserId, user.ID)
	return nil
}

func (s *userService) setUnauthenticated(ctx context.Context) error {
	err := s.sessionManager.RenewToken(ctx)
	if err != nil {
		return err
	}
	s.sessionManager.Put(ctx, constants.CtxAuthenticated, false)
	s.sessionManager.Put(ctx, constants.CtxUserId, nil)
	return nil
}
