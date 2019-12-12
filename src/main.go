package main

import (
	"context"
	"fmt"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"otus-hiload/src/constants"
	"otus-hiload/src/fake"
	"otus-hiload/src/file_storage"
	"otus-hiload/src/middleware"
	"otus-hiload/src/repository"
	"otus-hiload/src/service"
	"syscall"
	"time"
)

func main() {
	port := os.Getenv("SERVICE_PORT")
	if len(port) == 0 {
		log.Fatalf("SERVICE_PORT env variable not set")
	}
	dsn := os.Getenv("DB_URI")
	if len(dsn) == 0 {
		log.Fatalf("DB_URI env variable not set")
	}
	shardedDsn := os.Getenv("SHARDED_DB_URI")
	if len(shardedDsn) == 0 {
		log.Fatalf("SHARDED_DB_URI env variable not set")
	}
	storageDir := os.Getenv("STORAGE_DIR")
	if len(dsn) == 0 {
		log.Fatalf("STORAGE_DIR env variable not set")
	}
	generateStr := os.Getenv("GENERATE_FAKE_DATA")
	generate := false
	if len(generateStr) > 0 {
		generate = true
	}

	err := repository.Migrate(dsn, "migrations")
	if err != nil {
		log.Fatalf("migration error: %s", err.Error())
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	userRepo := repository.NewUserRepository(dsn)
	chatRepo := repository.NewChatRepository(dsn)
	msgRepo := repository.NewMessageRepository(shardedDsn)

	if generate {
		log.Println("start generation")
		count := 1000000
		users := make([]*repository.User, count, count)
		for i := 1; i <= count; i++ {
			name, lastName := fake.GetRandomName()
			user := new(repository.User)
			user.Login = fmt.Sprintf("login%d", i)
			user.Name = name
			user.LastName = lastName
			user.PasswordHash = ""
			user.Description = fmt.Sprintf("описание %s %s", name, lastName)
			users[i-1] = user
		}
		log.Println("generation finished, start saving")
		userRepo.BulkCreate(users)
		log.Println("saving finished")
	}

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(userRepo.GetDB())

	storage := file_storage.NewFileStorage(storageDir)
	userService := service.NewUserService(userRepo, sessionManager, storage)
	chatService := service.NewChatService(userRepo, chatRepo, msgRepo, sessionManager)

	r := mux.NewRouter()
	r.Use(middleware.RecoverHandler)
	r.Use(sessionManager.LoadAndSave)

	r.Handle(constants.RegPath, middleware.NotAuthHandler(http.HandlerFunc(userService.RegHandler), sessionManager)).Methods("GET", "POST")
	r.Handle(constants.LoginPath, middleware.NotAuthHandler(http.HandlerFunc(userService.LoginHandler), sessionManager)).Methods("GET", "POST")

	r.Handle(constants.LogoutPath, middleware.AuthHandler(http.HandlerFunc(userService.LogoutHandler), sessionManager)).Methods("GET")
	r.Handle(constants.MePath, middleware.AuthHandler(http.HandlerFunc(userService.MeHandler), sessionManager)).Methods("GET")
	r.Handle(constants.MeEditPath, middleware.AuthHandler(http.HandlerFunc(userService.EditHandler), sessionManager)).Methods("GET", "POST")
	r.Handle(constants.RootPath, middleware.AuthHandler(http.HandlerFunc(userService.RootHandler), sessionManager)).Methods("GET")
	r.Handle(constants.UserPath, middleware.AuthHandler(http.HandlerFunc(userService.UserHandler), sessionManager)).Methods("GET")
	r.Handle(constants.SearchPath, http.HandlerFunc(userService.SearchHandler)).Methods("GET")

	r.Handle(constants.ChatsPath, middleware.AuthHandler(http.HandlerFunc(chatService.ListChatsHandler), sessionManager)).Methods("GET")
	r.Handle(constants.ChatFormPath, middleware.AuthHandler(http.HandlerFunc(chatService.LoadMessagesFormHandler), sessionManager)).Methods("GET")
	r.Handle(constants.LoadMessagePath, middleware.AuthHandler(http.HandlerFunc(chatService.LoadChatMessagesHandler), sessionManager)).Methods("GET")
	// add
	r.Handle(constants.FirstMessagePath, middleware.AuthHandler(http.HandlerFunc(chatService.AddFirstMessageHandler), sessionManager)).Methods("POST")
	r.Handle(constants.MessagePath, middleware.AuthHandler(http.HandlerFunc(chatService.AddMessageHandler), sessionManager)).Methods("POST")

	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir(storageDir))))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	listen(srv, 5)
}

func listen(srv *http.Server, timeout time.Duration) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	<-done
	log.Print("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}
	log.Print("Server Exited Properly")
}
