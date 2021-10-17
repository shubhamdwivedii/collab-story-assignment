package mysql

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	storage     *MySQLStorage
	storyId     int32
	paragraphId int32
	sentenceId  int32
)

// Test in Separeate Test DB and Truncate Previous Test Data.

func TestNewMySQLStorage(t *testing.T) {
	// connect := "root:admin@tcp(127.0.0.1:3306)/collab"
	connect := os.Getenv("VERLOOP_DSN")
	logger := log.New()
	var err error
	storage, err = NewMySQLStorage(connect, logger)
	require.NoError(t, err)
}

func TestAddStory(t *testing.T) {
	var err error
	storyId, err = storage.AddStory()
	require.NoError(t, err)
}

func TestAddParagraph(t *testing.T) {
	var err error
	paragraphId, err = storage.AddParagraph(storyId)
	require.NoError(t, err)
}

func TestAddSentence(t *testing.T) {
	var err error
	sentenceId, err = storage.AddSentence(paragraphId, "First")
	require.NoError(t, err)
}

func TestGetUnfinishedStory(t *testing.T) {
	story, err := storage.GetUnfinishedStory()
	require.NoError(t, err)

	assert.Equal(t, storyId, story.ID, "Expected Story IDs To Match")
}

func TestGetUnfinishedParagrph(t *testing.T) {
	paragraph, err := storage.GetUnfinishedParagraph(storyId)
	require.NoError(t, err)

	assert.Equal(t, paragraphId, paragraph.ID, "Expected Paragraph IDS to Match")
}

func TestGetUnfinishedSentence(t *testing.T) {
	sentence, err := storage.GetUnfinishedSentence(paragraphId)
	require.NoError(t, err)

	assert.Equal(t, sentenceId, sentence.ID, "Expected Sentence IDs To Match")
}

func TestUpdateStoryTitle(t *testing.T) {
	err := storage.UpdateStoryTitle(storyId, "Random")
	require.NoError(t, err)
	err = storage.UpdateStoryTitle(storyId, "Title")
	require.NoError(t, err)

	// Get Story and Check Title for Furthur Tests.
}
