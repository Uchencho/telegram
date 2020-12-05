package ws

import (
	"log"
	"testing"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/testutils"
)

func TestGetorCreateThread(t *testing.T) {

	log.Println("Started the tests")

	testDb := testutils.CreateDB()
	defer func() {
		testutils.DropDB()
	}()

	db.MigrateDB(testDb)

}
