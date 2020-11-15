package account

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/pkg/errors"
)

type loginInfo struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshT struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type accessT struct {
	AccessToken string `json:"access_token"`
}

type loginResponse struct {
	ID           uint      `json:"id,omitempty"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	PhoneNumber  string    `json:"phone_number"`
	UserAddress  string    `json:"user_address"`
	IsActive     bool      `json:"is_active"`
	DateJoined   time.Time `json:"date_joined"`
	LastLogin    time.Time `json:"last_login"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

// AddRecordToUserTable adds a record to the db
func AddRecordToUserTable(db *sql.DB, user auth.User) error {
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

// GetUserLogin Queries the customer's entire details and updates the laast login field if customer exist, using db transactions
func GetUserLogin(db *sql.DB, email string) (auth.User, error) {

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return auth.User{}, err
	}

	query := `SELECT id, email, hashed_password, first_name, phone_number, user_address,
				is_active, date_joined, last_login FROM Users WHERE email = ?;`

	var (
		user    auth.User
		address interface{}
	)

	prep, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return auth.User{}, errors.Wrap(err, "Error occured in preparing query")
	}
	row := prep.QueryRowContext(ctx, email)
	switch err := row.Scan(&user.ID, &user.Email, &user.HashedPassword,
		&user.FirstName, &user.PhoneNumber, &address,
		&user.IsActive, &user.DateJoined, &user.LastLogin); err {
	case sql.ErrNoRows:
		_ = tx.Rollback()
		return auth.User{}, err
	case nil:
		if address == nil {
			user.UserAddress = ""
		} else {
			user.UserAddress = fmt.Sprint(address)
		}
	default:
		log.Printf("Getting user returned uncaught error, %v", err)
		return auth.User{}, err
	}

	query = `UPDATE Users SET last_login = ? WHERE email = ?;`
	prep, err = db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error occured in preparing query, %v", err)
		return auth.User{}, err
	}
	_, err = prep.ExecContext(ctx, time.Now(), email)
	if err != nil {
		_ = tx.Rollback()
		return auth.User{}, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error occured in commiting the transaction, %v", err)
		return auth.User{}, err
	}
	return user, nil
}

// UpdateUserRecord updates the first name and phone number of a user
func UpdateUserRecord(db *sql.DB, user auth.User) error {
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
