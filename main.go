package main

import (
	"log"
	"net/http"
	"os"
	"time"

	mux "github.com/gorilla/mux"
	mw "github.com/shubhamdwivedii/collab-story-assignment/pkg/middlewares"
	sv "github.com/shubhamdwivedii/collab-story-assignment/pkg/server"
	st "github.com/shubhamdwivedii/collab-story-assignment/pkg/storage/mysql"
	str "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
	ws "github.com/shubhamdwivedii/collab-story-assignment/pkg/word"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err.Error())
	}
	// InfoLogger = log.New(file, "INFO: ", log.LstdFlags|log.Lshortfile)
	InfoLogger = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.LstdFlags|log.Lshortfile)
}

func main() {
	InfoLogger.Println("This is some info...")
	ErrorLogger.Println("Some Error ?? HUH ??")

	DB_URL := "root:hesoyam@tcp(127.0.0.1:3306)/collab"
	storage, err := st.NewMySQLStorage(DB_URL, InfoLogger)
	if err != nil {
		log.Fatal("Error Initializing Storage: " + err.Error())
	}

	wrdsrv := ws.NewWordService(storage)

	strsrv := str.NewStoryService(storage)

	router := mux.NewRouter()

	service, err := sv.NewServer(wrdsrv, strsrv, InfoLogger)

	router.HandleFunc("/add", mw.ResponseTimeLogger(service.AddWordHandler, InfoLogger)).Methods("POST")
	router.HandleFunc("/stories", mw.ResponseTimeLogger(service.GetStoriesHandler, InfoLogger)).Methods("GET")
	router.HandleFunc("/stories/{story}", mw.ResponseTimeLogger(service.GetStoryHandler, InfoLogger)).Methods("GET")

	server := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(server.ListenAndServe())

}
