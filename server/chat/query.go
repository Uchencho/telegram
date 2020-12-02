package chat

import (
	"database/sql"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/pkg/errors"
)

// Thread is a representation of a thread stored in the db
type Thread struct {
	ID             int        `json:"id,omitempty" validate:"required"`
	FirstUserID    int        `json:"first_user_id,omitempty"`
	FirstUsername  string     `json:"first_username,omitempty"`
	SecondUserID   int        `json:"second_user_id,omitempty"`
	SecondUsername string     `json:"second_username,omitempty"`
	Updated        *time.Time `json:"updated,omitempty"`
	Created        *time.Time `json:"created,omitempty"`
}

// Message is a representation of a stored message in the DB
type Message struct {
	ID        int        `json:"id,omitempty"`
	Thread    int        `json:"thread,omitempty"`
	UserID    int        `json:"user_id,omitempty"`
	Username  string     `json:"user_name,omitempty"`
	Chatmsg   string     `json:"message,omitempty"`
	InputTime *time.Time `json:"input_time,omitempty"`
}

func chatThreadsByUser(db *sql.DB, user auth.User) ([]Thread, error) {
	query := `SELECT DISTINCT * FROM Thread WHERE firstUserID = ? OR secondUserID = ?;`
	prep, err := db.Prepare(query)
	if err != nil {
		return []Thread{}, errors.Wrap(err, "chat - could not prepare query")
	}

	records, err := prep.Query(user.ID, user.ID)
	defer records.Close()

	threadResult := Thread{}
	results := []Thread{}

	for records.Next() {
		err := records.Scan(&threadResult.ID, &threadResult.FirstUserID, &threadResult.FirstUsername,
			&threadResult.SecondUserID, &threadResult.SecondUsername, &threadResult.Updated, &threadResult.Created)
		if err != nil {
			return []Thread{}, errors.Wrap(err, "chat - could not scan thread record")
		}
		results = append(results, threadResult)
	}
	return results, nil
}

func getMessages(db *sql.DB, threadID int) ([]Message, error) {
	query := `SELECT * FROM ChatMessage WHERE thread = ?;`
	prep, err := db.Prepare(query)
	if err != nil {
		return []Message{}, errors.Wrap(err, "chat - could not prepare query")
	}

	records, err := prep.Query(threadID)
	defer records.Close()

	msgs := []Message{}
	aMsg := Message{}

	for records.Next() {
		err := records.Scan(&aMsg.ID, &aMsg.UserID, &aMsg.Username, &aMsg.Thread,
			&aMsg.Chatmsg, &aMsg.InputTime)
		if err != nil {
			return []Message{}, errors.Wrap(err, "chat - could not scan thread record")
		}
		msgs = append(msgs, aMsg)
	}
	return msgs, nil
}
