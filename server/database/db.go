package database

import (
	"github.com/Uchencho/telegram/server/auth"
)

// RetrieveUserLoginDetailsFunc returns the ability to retrieve a user's details from the database
type RetrieveUserLoginDetailsFunc func(string) (auth.User, error)

// UpdateUserDetailsFunc provides the ability to update a user's details
type UpdateUserDetailsFunc func(auth.User) error

// AddUserToDBFunc provides the ability to save a user's details to the DB
type AddUserToDBFunc func(user auth.User) error
