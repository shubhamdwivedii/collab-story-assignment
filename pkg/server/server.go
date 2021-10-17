package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	str "github.com/shubhamdwivedii/collab-story/pkg/story"
	wrd "github.com/shubhamdwivedii/collab-story/pkg/word"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	wordService  *wrd.WordService
	storyService *str.StoryService
	logger       *log.Logger
}

func NewServer(wordService *wrd.WordService, storyService *str.StoryService, logger *log.Logger) (*Server, error) {
	sv := new(Server)
	sv.wordService = wordService
	sv.storyService = storyService
	sv.logger = logger
	return sv, nil
}

// Adds a Word to Story/Paragraph/Sentence in Storage.
func (s *Server) AddWordHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

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

	wrdRes, err := s.wordService.AddWord(wrdReq.Word) // Also Verifies Word
	if err != nil {
		s.logger.Error("AddWord Failed:" + err.Error())
		// s.RespondWithError(w, http.StatusBadRequest, err.Error())
		wrdErr := wrd.WordError{Error: err.Error()}
		s.RespondWithJSON(w, http.StatusBadRequest, wrdErr)
		return
	}

	s.RespondWithJSON(w, http.StatusCreated, *wrdRes)
}

// Get All Stories from Storage
func (s *Server) GetStoriesHandler(w http.ResponseWriter, r *http.Request) {
	limitQry := r.URL.Query().Get("limit")
	offsetQry := r.URL.Query().Get("offset")

	limit, err := strconv.ParseInt(limitQry, 10, 32)
	if err != nil {
		limit = 10 // default value
	}
	offset, err := strconv.ParseInt(offsetQry, 10, 32)
	if err != nil {
		offset = 0 // default value
	}

	storiesRes, err := s.storyService.GetAllStories(int32(limit), int32(offset))

	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.RespondWithJSON(w, http.StatusAccepted, *storiesRes)
	return
}

// Get Story by ID from Storage
func (s *Server) GetStoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	storyId := vars["story"]
	id, err := strconv.ParseInt(storyId, 10, 32)

	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "Invalid Id")
	}

	storyRes, err := s.storyService.GetStoryDetail(int32(id))
	s.RespondWithJSON(w, http.StatusCreated, *storyRes)
}

func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string) {
	s.RespondWithJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Error Marshalling Response Data", err)
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	// As soon as w.Write() is executed, the Server will send the response.
}
