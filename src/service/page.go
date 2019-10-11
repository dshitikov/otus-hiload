package service

import (
	"errors"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
	"otus-hiload/src/constants"
	"strconv"
)

type IPageService interface {
	EditHandler(w http.ResponseWriter, r *http.Request)
	MeHandler(w http.ResponseWriter, r *http.Request)
	UserHandler(w http.ResponseWriter, r *http.Request)
	RootHandler(w http.ResponseWriter, r *http.Request)
}

func (s *userService) EditHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromContext(r.Context())
	if err != nil {
		s.logError("EditHandler getUserFromContext", err)
		http.Redirect(w, r, constants.RootPath, http.StatusFound)
	}

	if r.Method == "GET" {
		s.renderForm(w, "edit", nil)
	}

	if r.Method == "POST" {
		// max 10 MB
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			s.logError("EditHandler ParseMultipartForm", err)
			s.renderForm(w, "edit", errors.New("ошибка обработки формы"))
			return
		}

		description := r.FormValue("descr")
		if len(description) < 20 {
			s.renderForm(w, "edit", errors.New("заполните описание (не менее 20 символов)"))
			return
		}
		description = html.EscapeString(description)

		file, header, err := r.FormFile("photo")
		if err != nil {
			s.logError("EditHandler formFile", err)
			s.renderForm(w, "edit", errors.New("не выбран файл фото"))
			return
		}
		defer file.Close()

		log.Printf("Uploaded File: %+v\n", header.Filename)
		log.Printf("File Size: %+v\n", header.Size)
		log.Printf("MIME Header: %+v\n", header.Header)

		fName, err := s.storage.SaveFile(file, header.Filename)
		if err != nil {
			s.logError("saveFile", err)
			s.renderForm(w, "edit", errors.New("ошибка загрузки файла"))
			return
		}

		var oldFile = user.PhotoFile

		user.Description = description
		user.PhotoFile = fName
		err = s.UserRepository.Update(user)
		if err != nil {
			s.storage.DeleteFile(fName)
			s.logError("EditHandler UpdateUser", err)
			s.renderForm(w, "edit", errors.New("внутренняя ошибка сервера"))
			return
		}

		s.storage.DeleteFile(oldFile)
		http.Redirect(w, r, constants.MePath, http.StatusFound)
	}
}

func (s *userService) MeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromContext(r.Context())
	if err != nil {
		s.logError("MeHandler getUserFromContext", err)
		http.Redirect(w, r, constants.RootPath, http.StatusFound)
	}

	if len(user.Description) == 0 {
		log.Printf("user has no description, go to edit")
		http.Redirect(w, r, constants.MeEditPath, http.StatusFound)
	}

	params := make(map[string]string)
	params["description"] = user.Description
	params["name"] = user.Name
	params["image"] = user.PhotoFile

	s.renderFormParams(w, "me", params)
}

func (s *userService) UserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.logError("UserHandler parseInt", err)
		http.Redirect(w, r, constants.RootPath, http.StatusFound)
	}

	user, err := s.UserRepository.Get(id)
	if err != nil {
		s.logError("UserHandler UserRepository.Get", err)
		http.Redirect(w, r, constants.RootPath, http.StatusFound)
	}

	params := make(map[string]string)
	params["description"] = user.Description
	params["name"] = user.Name
	params["image"] = user.PhotoFile

	s.renderFormParams(w, "user", params)
}

func (s *userService) RootHandler(w http.ResponseWriter, r *http.Request) {
	user, err := s.getUserFromContext(r.Context())
	params := make(map[string]interface{})
	if err != nil {
		s.logError("RootHandler getUserFromContext", err)
		s.renderForm(w, "root", errors.New("внутренняя ошибка сервера"))
		return
	}

	users, err := s.UserRepository.GetAll()
	if err != nil {
		s.logError("UserRepository.GetAll", err)
		s.renderForm(w, "root", errors.New("внутренняя ошибка сервера"))
		return
	}
	params["users"] = users
	params["myId"] = user.Id
	s.renderFormParams(w, "root", params)
}
