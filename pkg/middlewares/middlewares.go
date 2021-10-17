package middlewares

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func DurationLogger(next http.HandlerFunc, logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next(w, r)
		logger.Info("Request at \"%s\" Processed in %s", r.URL.Path, time.Now().Sub(startTime))
	}
}
