package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	mux "github.com/gorilla/mux"
	mw "github.com/shubhamdwivedii/collab-story/pkg/middlewares"
	sv "github.com/shubhamdwivedii/collab-story/pkg/server"
	st "github.com/shubhamdwivedii/collab-story/pkg/storage/mysql"
	str "github.com/shubhamdwivedii/collab-story/pkg/story"
	wrd "github.com/shubhamdwivedii/collab-story/pkg/word"
)

var (
	logger *logrus.Logger
)

func init() {
	logger = logrus.New()
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logger.Fatal(err.Error())
	}
	// debugMode := "TruE"
	debugMode := os.Getenv("LOGS_ENABLE")
	mw := io.MultiWriter(os.Stdout, file)
	logger.Out = mw
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		ForceColors:   true,
		FullTimestamp: false,
	})

	if debugMode == "1" || strings.ToLower(debugMode) == "true" {
		logger.Level = logrus.DebugLevel
	} else {
		logger.Level = logrus.FatalLevel
	}
}

func main() {
	DB_URL := os.Getenv("DB_URL")
	// DB_URL := "root:admin@tcp(127.0.0.1:3306)/collab"
	storage, err := st.NewMySQLStorage(DB_URL, logger)
	if err != nil {
		logger.Fatal(err)
	}

	wordService := wrd.NewWordService(storage, logger)
	storyService := str.NewStoryService(storage, logger)
	router := mux.NewRouter()

	server, err := sv.NewServer(wordService, storyService, logger)

	router.HandleFunc("/add", mw.DurationLogger(server.AddWordHandler, logger)).Methods("POST")
	router.HandleFunc("/stories", mw.DurationLogger(server.GetStoriesHandler, logger)).Methods("GET")
	router.HandleFunc("/stories/{story}", mw.DurationLogger(server.GetStoryHandler, logger)).Methods("GET")

	httpServer := &http.Server{
		Handler:      router,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(httpServer.ListenAndServe())
}
