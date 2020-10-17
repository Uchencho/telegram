package account

import (
	"database/sql"
	"log"
	"time"

	"github.com/Uchencho/telegram/server/auth"
)

type loginResponse struct {
	ID           uint      `json:"id,omitempty"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	PhoneNumber  string    `json:"phone_number"`
	UserAddress  string    `json:"user_address"`
	IsActive     bool      `json:"is_active"`
	DateJoined   time.Time `json:"date_joined"`
	LastLogin    time.Time `json:"last_login"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}

func AddRecordToUserTable(db *sql.DB, user auth.User) error {
	query := `INSERT INTO Users (
		email, hashed_password, date_joined, last_login, is_active, first_name,
		phone_number, longitude, latitude, device_id
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?
	);`

	prep, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error occured in preparing query, %v", err)
		return err
	}

	_, err = prep.Exec(user.Email, user.HashedPassword,
		user.DateJoined, user.LastLogin, true, user.FirstName,
		user.PhoneNumber, user.Longitude, user.Latitude, user.DeviceID)
	return err
}
