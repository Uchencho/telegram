package app_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/account"
	"github.com/Uchencho/telegram/server/app"
	"github.com/Uchencho/telegram/server/testutils"
	"github.com/Uchencho/telegram/server/utils"
	"github.com/stretchr/testify/assert"
)

func setupTestEnv() (string, *sql.DB, func()) {
	sqliteForTest := testutils.CreateDB()

	driver := testutils.GetTestDriver(sqliteForTest)

	db.MigrateDB(sqliteForTest, driver, "sqlite3")

	TestApp := app.NewApp(sqliteForTest)
	url, closeServer := testutils.NewTestServer(TestApp.Handler())
	return url, sqliteForTest, closeServer

}
func TestRegisterSuccess(t *testing.T) {
	defer testutils.DropDB()

	var (
		requestBody  account.RegisterUser
		expectedResp utils.SuccessResponse
		responseBody utils.SuccessResponse
	)

	url, sqliteDB, closeServer := setupTestEnv()
	defer sqliteDB.Close()
	defer closeServer()

	testutils.FileToStruct(filepath.Join("test_data", "register_request.json"), &requestBody)

	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/register", url), bytes.NewBuffer(jsonBody))
	testutils.SetTestStandardHeaders(req)

	res, _ := http.DefaultClient.Do(req)
	testutils.GetResponseBody(res, responseBody)

	testutils.FileToStruct(filepath.Join("test_data", "register_response.json"), &expectedResp)

	t.Run("HTTP response status is 200", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
	// t.Run("Response body is as expected", func(t *testing.T) {
	// 	assert.Equal(t, expectedResp, responseBody)
	// })
}
