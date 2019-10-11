package middleware

import (
	"errors"
	"log"
	"net/http"
	"runtime"
)

const (
	StackSize = 4 << 10 // 4 KB
)

func RecoverHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				stack := make([]byte, StackSize)
				length := runtime.Stack(stack, false)
				log.Printf("[PANIC RECOVER]: %v %s\n", err, stack[:length])
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}
