package chat

import (
	"database/sql"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/pkg/errors"
)

type thread struct {
	ID             int        `json:"id,omitempty" validate:"required"`
	FirstUserID    int        `json:"first_user_id,omitempty"`
	FirstUsername  string     `json:"first_username,omitempty"`
	SecondUserID   int        `json:"second_user_id,omitempty"`
	SecondUsername string     `json:"second_username,omitempty"`
	Updated        *time.Time `json:"updated,omitempty"`
	Created        *time.Time `json:"created,omitempty"`
}

type message struct {
	ID        int        `json:"id,omitempty"`
	Thread    int        `json:"thread,omitempty"`
	UserID    int        `json:"user_id,omitempty"`
	Username  string     `json:"user_name,omitempty"`
	Chatmsg   string     `json:"message,omitempty"`
	InputTime *time.Time `json:"input_time,omitempty"`
}

func chatThreadsByUser(db *sql.DB, user auth.User) ([]thread, error) {
	query := `SELECT DISTINCT * FROM Thread WHERE firstUserID = ? OR secondUserID = ?;`
	prep, err := db.Prepare(query)
	if err != nil {
		return []thread{}, errors.Wrap(err, "chat - could not prepare query")
	}

	records, err := prep.Query(user.ID, user.ID)
	defer records.Close()

	threadResult := thread{}
	results := []thread{}

	for records.Next() {
		err := records.Scan(&threadResult.ID, &threadResult.FirstUserID, &threadResult.FirstUsername,
			&threadResult.SecondUserID, &threadResult.SecondUsername, &threadResult.Updated, &threadResult.Created)
		if err != nil {
			return []thread{}, errors.Wrap(err, "chat - could not scan thread record")
		}
		results = append(results, threadResult)
	}
	return results, nil
}

func getMessages(db *sql.DB, threadID int) ([]message, error) {
	query := `SELECT * FROM ChatMessage WHERE thread = ?;`
	prep, err := db.Prepare(query)
	if err != nil {
		return []message{}, errors.Wrap(err, "chat - could not prepare query")
	}

	records, err := prep.Query(threadID)
	defer records.Close()

	msgs := []message{}
	aMsg := message{}

	for records.Next() {
		err := records.Scan(&aMsg.ID, &aMsg.UserID, &aMsg.Username, &aMsg.Thread,
			&aMsg.Chatmsg, &aMsg.InputTime)
		if err != nil {
			return []message{}, errors.Wrap(err, "chat - could not scan thread record")
		}
		msgs = append(msgs, aMsg)
	}
	return msgs, nil
}
