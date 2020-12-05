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

// RetrieveUserLoginDetailsFunc returns the ability to retrieve a user's details from the database
type RetrieveUserLoginDetailsFunc func(string) (User, error)

// UpdateUserDetailsFunc provides the ability to update a user's details
type UpdateUserDetailsFunc func(User) error

// AddUserToDBFunc provides the ability to save a user's details to the DB
type AddUserToDBFunc func(user User) error
