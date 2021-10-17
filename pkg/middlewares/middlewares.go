package middlewares

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func DurationLogger(next http.HandlerFunc, logger *logrus.Logger) http.HandlerFunc {
	total := 0 // Implement better metric logic later.
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		next(w, r)
		logger.Info("Request at \"", r.URL.Path, "\" Processed in ", time.Now().Sub(startTime))
		total++
		logger.Info("Total Requests: ", total)
	}
}
