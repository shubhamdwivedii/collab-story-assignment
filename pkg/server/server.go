package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	str "github.com/shubhamdwivedii/collab-story-assignment/pkg/story"
	wrd "github.com/shubhamdwivedii/collab-story-assignment/pkg/word"
)

// type Service interface {
// 	AddWord(word string) (*wrd.WordResponse, error)
// 	GetAllStories(limit int32, offset int32) (*str.StoriesResponse, error)
// }

type Server struct {
	wrdSrv *wrd.WordService
	strSrv *str.StoryService
	// router *mux.Router
	logger *log.Logger
}

func NewServer(wrdSrv *wrd.WordService, strSrv *str.StoryService, logger *log.Logger) (*Server, error) {
	// var err error
	sv := new(Server)
	sv.wrdSrv = wrdSrv
	sv.strSrv = strSrv
	sv.logger = logger
	return sv, nil
}

func (s *Server) AddWordHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	// body is []byte

	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return // always return after sending response.
	}

	contentType := r.Header.Get("content-type")

	if contentType != "application/json" {
		s.RespondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}

	var wrdReq wrd.WordRequest
	err = json.Unmarshal(body, &wrdReq)

	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	wrdRes, err := s.wrdSrv.AddWord(wrdReq.Word) // Also Verifies Word
	if err != nil {
		// s.RespondWithError(w, http.StatusBadRequest, err.Error())
		wrdErr := wrd.WordError{Error: err.Error()}
		s.RespondWithJSON(w, http.StatusBadRequest, wrdErr)
		return
	}

	fmt.Println("WordRes", *wrdRes)
	s.RespondWithJSON(w, http.StatusCreated, *wrdRes)
}

func (s *Server) GetStoriesHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.Println("Story Get All Stories")

	storiesRes, err := s.strSrv.GetAllStories(5, 0)

	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.RespondWithJSON(w, http.StatusAccepted, *storiesRes)
	return
}

func (s *Server) GetStoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyId := vars["story"]
	s.logger.Println("Story Get", storyId)

	// get story by id
	// get paragraphs by story id
	// get sentences by paragraph id

	id, err := strconv.ParseInt(storyId, 10, 32)

	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "Invalid Id")
	}

	storyRes, err := s.strSrv.GetStoryDetail(int32(id))

	s.RespondWithJSON(w, http.StatusCreated, *storyRes)
}

func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string) {
	s.RespondWithJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		s.logger.Println("Error Marshalling Response Data", err)
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	// As soon as w.Write() is executed, the Server will send the response.
}
