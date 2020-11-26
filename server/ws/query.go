package ws

import (
	"context"
	"database/sql"
	"log"

	"github.com/Uchencho/telegram/server/chat"
	"github.com/pkg/errors"
)

func storeMessage(db *sql.DB, msg chat.Message) {
	query := `INSERT INTO ChatMessage (
		userID, username, thread, chatmsg
	) VALUES (
		?, ?, ?, ?
	);`

	prep, err := db.Prepare(query)
	if err != nil {
		log.Println("ws - Could not prepare query")
	}

	_, err = prep.Exec(msg.UserID, msg.Username, msg.Thread, msg.Chatmsg)
	if err != nil {
		log.Println("ws - Could not execute query")
	}
}

func getOrCreateThread(db *sql.DB, thread chat.Thread) (threadID int, err error) {
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "ws - could not start a database transaction")
	}

	query := `SELECT DISTINCT id FROM Thread WHERE firstUserID = ? OR secondUserID = ?;`
	prep, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, errors.Wrap(err, "ws - could not prepare query within transaction")
	}

	row := prep.QueryRowContext(ctx, thread.FirstUserID, thread.SecondUserID)
	switch err = row.Scan(&threadID); err {
	case sql.ErrNoRows:
		threadID = 0
	case nil:
		if threadID != 0 {
			return threadID, nil
		}
	}

	if threadID == 0 {
		query = `INSERT INTO Thread (
			firstUserID, firstUsername, secondUserID, secondUsername
		) VALUES (
			?, ?, ?, ?
		);`

		prep, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return 0, errors.Wrap(err, "ws - could not prepare query within transaction")
		}
		result, err := prep.ExecContext(ctx, thread.FirstUserID, thread.FirstUsername, thread.SecondUserID,
			thread.SecondUsername)

		lastID, err := result.LastInsertId()
		if err != nil {
			return 0, errors.Wrap(err, "ws - could not retrieve last inserted ID")
		}
		err = tx.Commit()
		if err != nil {
			return 0, errors.Wrap(err, "ws - could not commit transaction into DB")
		}
		return int(lastID), nil
	}
	return threadID, err
}
