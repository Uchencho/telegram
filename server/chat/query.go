package chat

import (
	"database/sql"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/pkg/errors"
)

type thread struct {
	ID           int        `json:"id,omitempty"`
	FirstUserID  int        `json:"first_user_id,omitempty"`
	SecondUserID int        `json:"second_user_id,omitempty"`
	Updated      *time.Time `json:"updated,omitempty"`
	Created      *time.Time `json:"created,omitempty"`
}

func chatThreadsByUser(db *sql.DB, user auth.User) ([]thread, error) {
	query := `SELECT DISTINCT * FROM Thread WHERE firstUser = ? OR secondUser = ?;`
	prep, err := db.Prepare(query)
	if err != nil {
		return []thread{}, errors.Wrap(err, "chat - could not prepare query")
	}

	records, err := prep.Query(user.ID, user.ID)
	defer records.Close()

	threadResult := thread{}
	results := []thread{}

	for records.Next() {
		err := records.Scan(&threadResult.ID, &threadResult.FirstUserID, &threadResult.SecondUserID,
			&threadResult.Updated, &threadResult.Created)
		if err != nil {
			return []thread{}, errors.Wrap(err, "chat - could not scan thread record")
		}
		results = append(results, threadResult)
	}
	return results, nil

}
