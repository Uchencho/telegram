package db

import (
	"testing"

	"github.com/Uchencho/telegram/server/testutils"
)

func TestMigrateDBSuccess(t *testing.T) {
	defer testutils.DropDB()
	sqliteForTest := testutils.CreateDB()

	driver := testutils.GetTestDriver(sqliteForTest)
	MigrateDB(sqliteForTest, driver, "sqlite3")
}
