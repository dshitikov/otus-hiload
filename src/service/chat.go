package service

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"otus-hiload/src/constants"
	"otus-hiload/src/repository"
	"strconv"
	"time"
)

type chatService struct {
	*baseService
	chatRepository    repository.IChatRepository
	messageRepository repository.IMessageRepository
	sessionManager    *scs.SessionManager
	msgLoadSize       int
}

type chatData struct {
	ID         int64
	UserID     int64
	UpdateDate time.Time
	Name       string
	LastName   string
}

type messageData struct {
	ID       int64
	LastName string
	Name     string
	Date     time.Time
	Text     string
}

type IChatService interface {
	ListChatsHandler(w http.ResponseWriter, r *http.Request)
	LoadMessagesFormHandler(w http.ResponseWriter, r *http.Request)
	LoadChatMessagesHandler(w http.ResponseWriter, r *http.Request)
	AddFirstMessageHandler(w http.ResponseWriter, r *http.Request)
	AddMessageHandler(w http.ResponseWriter, r *http.Request)
}

func NewChatService(userRepo repository.IUserRepository, chatRepo repository.IChatRepository, messageRepo repository.IMessageRepository,
	sessionManager *scs.SessionManager) IChatService {
	baseService := &baseService{sessionManager: sessionManager, userRepository: userRepo}
	return &chatService{baseService: baseService, chatRepository: chatRepo, messageRepository: messageRepo, sessionManager: sessionManager, msgLoadSize: 100}
}

func (s *chatService) ListChatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.logError("ListChatsHandler getUserFromContext", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}
		chats, err := s.chatRepository.GetAllByUser(user.ID)
		if err != nil {
			s.logError("ListChatsHandler GetAllByUser", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}
		params := make(map[string]interface{})
		if len(chats) > 0 {
			data, err := s.getChatsData(user, chats)
			if err != nil {
				s.logError("ListChatsHandler getChatsData", err)
				http.Redirect(w, r, constants.RootPath, http.StatusFound)
			}
			params["chats"] = data
		} else {
			params["chats"] = []*chatData{}
		}

		s.renderFormParams(w, "chats", params)
	}
}

func (s *chatService) getChatsData(user *repository.User, chats []*repository.Chat) ([]*chatData, error) {
	ids := make([]int64, 0, len(chats))
	usersMap := make(map[int64]*repository.User)
	for _, v := range chats {
		if v.User1ID != user.ID {
			ids = append(ids, v.User1ID)
		} else {
			ids = append(ids, v.User2ID)
		}
	}

	users, err := s.userRepository.GetByIDs(ids)

	if err != nil {
		return nil, err
	}
	for _, v := range users {
		usersMap[v.ID] = v
	}

	chatsData := make([]*chatData, 0, len(chats))
	for _, v := range chats {
		data := new(chatData)
		data.ID = v.ID
		var u *repository.User
		if v.User1ID != user.ID {
			u = usersMap[v.User1ID]
			data.UserID = v.User1ID
		} else {
			u = usersMap[v.User2ID]
			data.UserID = v.User2ID
		}
		data.UpdateDate = v.UpdatedAt
		data.Name = u.Name
		data.LastName = u.LastName
		chatsData = append(chatsData, data)
	}
	return chatsData, nil
}

func (s *chatService) LoadMessagesFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.logError("LoadMessagesFormHandler getUserFromContext", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}

		vars := mux.Vars(r)
		idStr := vars["userId"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			s.logError("LoadMessagesFormHandler userId parseInt", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}
		user2, err := s.userRepository.Get(id)
		if err != nil {
			s.logError("LoadMessagesFormHandler userRepository.Get", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}
		chat, err := s.chatRepository.Get(id, user.ID)
		if err != nil && err != sql.ErrNoRows {
			s.logError("LoadMessagesFormHandler chatRepository.Get", err)
			http.Redirect(w, r, constants.RootPath, http.StatusFound)
		}

		params := make(map[string]interface{})
		params["user"] = user
		params["user2"] = user2
		if chat != nil {
			params["chat_id"] = chat.ID
			// messages, err := s.messageRepository.GetByChat(chat.ID, 0, 100)
			// if err != nil && err != sql.ErrNoRows {
			// 	s.logError("LoadMessagesFormHandler messageRepository.GetByChat", err)
			// 	http.Redirect(w, r, constants.RootPath, http.StatusFound)
			// }
			// enrichedMsgs := s.enrichMessages(messages, user, user2)
			// params["messages"] = enrichedMsgs
		}

		s.renderFormParams(w, "messages", params)
	}
}

func (s *chatService) LoadChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.logError("LoadChatMessagesHandler getUserFromContext", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		vars := mux.Vars(r)

		chatIdStr := vars["chatId"]
		chatID, err := strconv.ParseInt(chatIdStr, 10, 64)
		if err != nil {
			s.logError("LoadChatMessagesHandler parseInt", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		chat, err := s.chatRepository.GetByID(chatID)
		if err != nil {
			s.logError("LoadChatMessagesHandler chatRepository.GetByID", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var user2ID int64
		if chat.User1ID == user.ID {
			user2ID = chat.User2ID
		} else {
			user2ID = chat.User1ID
		}
		user2, err := s.userRepository.Get(user2ID)
		if err != nil {
			s.logError("LoadChatMessagesHandler userRepository.Get", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		minIdStr := vars["minId"]
		minMsgID, err := strconv.ParseInt(minIdStr, 10, 64)
		if err != nil {
			s.logError("LoadChatMessagesHandler parseInt", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		messages, err := s.messageRepository.GetByChat(chatID, minMsgID, 100)

		if err != nil {
			s.logError("LoadChatMessagesHandler messageRepository.GetByChat", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		enrichedMsgs := s.enrichMessages(messages, user, user2)

		js, err := json.Marshal(enrichedMsgs)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

func (s *chatService) enrichMessages(messages []*repository.Message, user *repository.User, user2 *repository.User) []*messageData {
	enrichedMsgs := make([]*messageData, 0, len(messages))
	for _, msg := range messages {
		enrichedMsg := new(messageData)
		if msg.UserID == user.ID {
			enrichedMsg.Name = user.Name
			enrichedMsg.LastName = user.LastName
		} else {
			enrichedMsg.Name = user2.Name
			enrichedMsg.LastName = user2.LastName
		}
		enrichedMsg.ID = msg.ID
		enrichedMsg.Date = msg.CreatedAt
		enrichedMsg.Text = msg.Text
		enrichedMsgs = append(enrichedMsgs, enrichedMsg)
	}
	return enrichedMsgs
}

func (s *chatService) AddMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.logError("AddMessageHandler getUserFromContext", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		vars := mux.Vars(r)
		idStr := vars["chatId"]
		chatID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			s.logError("AddMessageHandler parseInt", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.logError("AddMessageHandler ReadAll", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		err = s.messageRepository.Create(chatID, user.ID, string(b))
		if err != nil {
			s.logError("AddMessageHandler messageRepository.Create", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

func (s *chatService) AddFirstMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user, err := s.getUserFromContext(r.Context())
		if err != nil {
			s.logError("AddFirstMessageHandler getUserFromContext", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		vars := mux.Vars(r)
		idStr := vars["userId"]
		user2ID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			s.logError("AddFirstMessageHandler parseInt", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if user.ID == user2ID {
			s.logError("AddFirstMessageHandler attempt to chat with itself", err)
			http.Error(w, "same user", http.StatusBadRequest)
			return
		}

		chatID, err := s.chatRepository.Start(user.ID, user2ID)
		if err != nil {
			s.logError("AddFirstMessageHandler getUserFromContext", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.logError("AddFirstMessageHandler ReadAll", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		err = s.messageRepository.Create(chatID, user.ID, string(b))
		if err != nil {
			s.logError("AddFirstMessageHandler messageRepository.Create", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// return chat id to client
		_, err = fmt.Fprintf(w, "%d", chatID)
		if err != nil {
			s.logError("AddFirstMessageHandler Fprintf", err)
		}
	}
}
