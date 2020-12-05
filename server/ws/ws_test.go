package ws

import (
	"log"
	"testing"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetorCreateThread(t *testing.T) {

	log.Println("Started the tests")

	testDb := testutils.CreateDB()
	defer func() {
		testutils.DropDB()
	}()

	db.MigrateDB(testDb)

	thread := chat.Thread{}
	_, err := getOrCreateThread(testDb, thread)

	if assert.Error(t, err) {
		assert.Equal(t, nil, err)
	}

}
