package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
)

// AddRecordToUserTable adds a record to the db
func AddRecordToUserTable(dbinterface interface{}) AddUserToDBFunc {
	return func(user User) error {
		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return errors.New("Cannot use non sql type in this function")
		}

		query := `INSERT INTO Users (
			email, hashed_password, date_joined, last_login, is_active, first_name,
			phone_number, longitude, latitude, device_id
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		);`

		prep, err := db.Prepare(query)
		if err != nil {
			return errors.Wrap(err, "Error occured in preparing query")
		}

		_, err = prep.Exec(user.Email, user.HashedPassword,
			user.DateJoined, user.LastLogin, true, user.FirstName,
			user.PhoneNumber, user.Longitude, user.Latitude, user.DeviceID)
		return err
	}
}

// GetUserLogin Queries the customer's entire details and updates the laast login field if customer exist, using db transactions
func GetUserLogin(dbinterface interface{}) RetrieveUserLoginDetailsFunc {
	return func(email string) (User, error) {
		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return User{}, errors.New("Cannot use non sql type in this function")
		}

		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return User{}, err
		}

		query := `SELECT id, email, hashed_password, first_name, phone_number, user_address,
				is_active, date_joined, last_login FROM Users WHERE email = ?;`

		var (
			user    User
			address interface{}
		)

		prep, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return User{}, errors.Wrap(err, "Error occured in preparing query")
		}
		row := prep.QueryRowContext(ctx, email)
		switch err := row.Scan(&user.ID, &user.Email, &user.HashedPassword,
			&user.FirstName, &user.PhoneNumber, &address,
			&user.IsActive, &user.DateJoined, &user.LastLogin); err {
		case sql.ErrNoRows:
			_ = tx.Rollback()
			return User{}, err
		case nil:
			if address == nil {
				user.UserAddress = ""
			} else {
				user.UserAddress = fmt.Sprint(address)
			}
		default:
			log.Printf("Getting user returned uncaught error, %v", err)
			return User{}, err
		}

		query = `UPDATE Users SET last_login = ? WHERE email = ?;`
		prep, err = db.PrepareContext(ctx, query)
		if err != nil {
			log.Printf("Error occured in preparing query, %v", err)
			return User{}, err
		}
		_, err = prep.ExecContext(ctx, time.Now(), email)
		if err != nil {
			_ = tx.Rollback()
			return User{}, err
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("Error occured in commiting the transaction, %v", err)
			return User{}, err
		}
		return user, nil
	}
}

// UpdateUserRecord updates the first name and phone number of a user
func UpdateUserRecord(dbinterface interface{}) UpdateUserDetailsFunc {
	return func(user User) error {
		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return errors.New("Cannot use non sql type in this function")
		}

		query := `UPDATE Users SET first_name = ?, phone_number = ? WHERE email = ?;`
		prep, err := db.Prepare(query)
		if err != nil {
			return errors.Wrap(err, "account - Could not prepare query")
		}
		_, err = prep.Exec(user.FirstName, user.PhoneNumber, user.Email)
		if err != nil {
			return errors.Wrap(err, "account - Could not execute query")
		}
		return nil
	}
}

// GetUser retrieves the user details for the auth Middleware
func GetUser(dbinterface interface{}) RetrieveUserLoginDetailsFunc {
	return func(email string) (User, error) {

		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return User{}, errors.New("ONLY sql type allowed in this function")
		}

		query := `SELECT id, email, first_name, phone_number, user_address, 
	is_active, date_joined, last_login, hashed_password FROM Users WHERE email = ?;`

		prep, err := db.Prepare(query)
		if err != nil {
			return User{}, err
		}

		row := prep.QueryRow(email)
		var (
			user User
			add  interface{}
		)

		switch err := row.Scan(&user.ID, &user.Email, &user.FirstName,
			&user.PhoneNumber, &add, &user.IsActive,
			&user.DateJoined, &user.LastLogin, &user.HashedPassword); err {
		case sql.ErrNoRows:
			return User{}, err
		case nil:
			if add == nil {
				user.UserAddress = ""
			} else {
				user.UserAddress = fmt.Sprint(add)
			}
			return user, err
		default:
			return User{}, err
		}
	}
}

// ChatThreadsByUser retrieves the thread a user has participated in
func ChatThreadsByUser(dbinterface interface{}) RetrieveUserThreadsFunc {
	return func(user User) ([]Thread, error) {

		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return []Thread{}, errors.New("ONLY sql type allowed in this function")
		}

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
}

// GetMessages retrieves a list of messages in a specific thread
func GetMessages(dbinterface interface{}) RetrieveMessagesFunc {
	return func(threadID int) ([]Message, error) {

		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return []Message{}, errors.New("ONLY sql type allowed in this function")
		}

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
}

// StoreMessage inserts a message into the DB. Error checking is NOT properly handled
func StoreMessage(dbinterface interface{}) InsertMessageFunc {
	return func(msg Message) {

		db, ok := dbinterface.(*sql.DB)
		if !ok {
			log.Println("ONLY sql type allowed in this function")
			return
		}

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
}

// GetOrCreateThread retrieves a thread from the DB or creates one
func GetOrCreateThread(dbinterface interface{}) GetorCreateThreadFunc {
	return func(thread Thread) (threadID int, err error) {

		db, ok := dbinterface.(*sql.DB)
		if !ok {
			return 0, errors.New("ONLY sql type aloowed with this function")
		}

		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return 0, errors.Wrap(err, "ws - could not start a database transaction")
		}

		query := `SELECT DISTINCT id FROM Thread WHERE firstUserID = ? AND secondUserID = ? OR firstUserID = ? AND secondUserID = ?;`
		prep, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return 0, errors.Wrap(err, "ws - could not prepare query within transaction")
		}

		row := prep.QueryRowContext(ctx, thread.FirstUserID, thread.SecondUserID, thread.SecondUserID, thread.FirstUserID)
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
}
