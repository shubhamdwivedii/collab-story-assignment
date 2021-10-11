package main

import (
	"log"

	st "github.com/shubhamdwivedii/collab-story-assignment/pkg/storage"
	ws "github.com/shubhamdwivedii/collab-story-assignment/pkg/word"
)

func main() {
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
