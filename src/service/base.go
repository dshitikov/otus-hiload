package service

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
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

func (s *userService) deleteFile(path string) {
	if len(path) == 0 {
		return
	}

	var _, err = os.Stat(path)
	if err != nil {
		log.Printf("deleteFile Stat error: %s", err.Error())
		return
	}

	err = os.Remove(path)
	if err != nil {
		log.Printf("deleteFile remove error: %s", err.Error())
	}
}

func (s *userService) checkFileType(reader io.Reader) error {
	buff := make([]byte, 512)
	_, err := reader.Read(buff)

	if err != nil {
		return err
	}

	fileType := http.DetectContentType(buff)

	switch fileType {
	case "image/jpeg", "image/jpg":
	case "image/gif":
	case "image/png":
	default:
		return fmt.Errorf("unknown file type uploaded: %s", fileType)
	}

	return nil
}
