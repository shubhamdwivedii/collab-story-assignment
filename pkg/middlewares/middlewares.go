package middlewares

import (
	"log"
	"net/http"
	"time"
)

func ResponseTimeLogger(next http.HandlerFunc, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next(w, r)
		logger.Printf("Request Processed in %s", time.Now().Sub(startTime))
	}
}
