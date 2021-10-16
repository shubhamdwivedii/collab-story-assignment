package main

import (
	"log"
	"os"

	st "github.com/shubhamdwivedii/collab-story-assignment/pkg/storage/mysql"
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
	InfoLogger = log.New(file, "INFO: ", log.LstdFlags|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.LstdFlags|log.Lshortfile)
}

func main() {
	InfoLogger.Println("This is some info...")
	ErrorLogger.Println("Some Error ?? HUH ??")

	DB_URL := "root:admin@tcp(127.0.0.1:3306)/collab"
	storage, err := st.NewMySQLStorage(DB_URL)

	if err != nil {
		log.Fatal("Error Initializing Storage: " + err.Error())
	}

	wrdsrv := ws.NewWordService(storage)

	words := []string{
		"my", "story",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"hello", "my", "name", "is", "Slim", "shady", "hi", "kids", "do", "you", "like", "violence", "wanna", "see", "magic",
		"this", "is", "unfinished", "sentence",
	}

	for _, word := range words {
		err := wrdsrv.AddWord(word)
		if err != nil {
			log.Println("error adding Word", err.Error())
			break
		}
	}
}
