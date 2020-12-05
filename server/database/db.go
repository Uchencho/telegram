package database

import (
	"time"
)

// User is a representation of a user
type User struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email" validate:"required,email"`
	HashedPassword string    `json:"password,omitempty"`
	FirstName      string    `json:"first_name"`
	PhoneNumber    string    `json:"phone_number"`
	UserAddress    string    `json:"user_address"`
	IsActive       bool      `json:"is_active"`
	DateJoined     time.Time `json:"date_joined"`
	LastLogin      time.Time `json:"last_login"`
	Longitude      string    `json:"longitude"`
	Latitude       string    `json:"latitude"`
	DeviceID       string    `json:"device_id"`
}

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

// RetrieveUserLoginDetailsFunc returns the ability to retrieve a user's details from the database
type RetrieveUserLoginDetailsFunc func(string) (User, error)

// UpdateUserDetailsFunc provides the ability to update a user's details
type UpdateUserDetailsFunc func(User) error

// AddUserToDBFunc provides the ability to save a user's details to the DB
type AddUserToDBFunc func(User) error

// RetrieveUserThreads provides the ability to retrieve a user's thread from the DB
type RetrieveUserThreads func(User) ([]Thread, error)

// RetrieveMessages provides the functionality to retrieve messages in a thread
type RetrieveMessages func(int) ([]Message, error)
