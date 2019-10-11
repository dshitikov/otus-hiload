package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
	"os/signal"
	"otus-hiload/src/constants"
	"otus-hiload/src/file_storage"
	"otus-hiload/src/middleware"
	"otus-hiload/src/repository"
	"otus-hiload/src/service"
	"syscall"
	"time"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
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
	storageDir := os.Getenv("STORAGE_DIR")
	if len(dsn) == 0 {
		log.Fatalf("STORAGE_DIR env variable not set")
	}

	err := repository.Migrate(dsn, "migrations")
	if err != nil {
		log.Fatalf("migration error: %s", err.Error())
	}

	repo := repository.NewMysqlRepository(dsn)
	storage := file_storage.NewFileStorage(storageDir)
	userService := service.NewUserService(repo, store, storage)

	r := mux.NewRouter()
	r.Use(middleware.RecoverHandler)
	r.Handle(constants.LoginPath, http.HandlerFunc(userService.LoginHandler)).Methods("GET", "POST")
	r.Handle(constants.LogoutPath, middleware.AuthHandler(http.HandlerFunc(userService.LogoutHandler), store)).Methods("GET")
	r.Handle(constants.RegPath, http.HandlerFunc(userService.RegHandler)).Methods("GET", "POST")
	r.Handle(constants.MePath, middleware.AuthHandler(http.HandlerFunc(userService.MeHandler), store)).Methods("GET")
	r.Handle(constants.MeEditPath, middleware.AuthHandler(http.HandlerFunc(userService.EditHandler), store)).Methods("GET", "POST")
	r.Handle(constants.RootPath, middleware.AuthHandler(http.HandlerFunc(userService.RootHandler), store)).Methods("GET")
	r.Handle(constants.UserPath, middleware.AuthHandler(http.HandlerFunc(userService.UserHandler), store)).Methods("GET")
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
