package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testFindOrCreateByName    = "libsacloud_test_note_name"
	testFindOrCreateByContent = "libsacloud_test_note_content"
)

func TestCRUDByNoteAPI(t *testing.T) {
	noteAPI := client.Note

	//CREATE
	var note = noteAPI.New()
	note.Name = testTargetNoteName
	note.Content = testTargetNoteContentBefore

	res, err := noteAPI.Create(note)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)

	//for READ
	var id = res.ID

	//READ
	res, err = noteAPI.Read(id)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.NotEmpty(t, res.Content)

	//UPDATE
	note.Content = testTargetNoteContentAfter

	res, err = noteAPI.Update(id, note)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.NotEmpty(t, res.Content)
	assert.Equal(t, res.Content, testTargetNoteContentAfter)

	//DELETE
	res, err = noteAPI.Delete(id)
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func init() {
	testSetupHandlers = append(testSetupHandlers, cleanupTestNote)
	testSetupHandlers = append(testSetupHandlers, cleanupFindOrCreateByNote)

	testTearDownHandlers = append(testTearDownHandlers, cleanupTestNote)
	testTearDownHandlers = append(testTearDownHandlers, cleanupFindOrCreateByNote)
}

func cleanupFindOrCreateByNote() {
	noteAPI := client.Note
	res, _ := noteAPI.withNameLike(testFindOrCreateByName).Find()
	if res.Count > 0 {
		noteAPI.Delete(res.Notes[0].ID)
	}
}
