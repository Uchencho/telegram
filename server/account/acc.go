package account

import (
	"time"
)

// LoginInfo represents the functionality of a login information
type LoginInfo struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshT struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type accessT struct {
	AccessToken string `json:"access_token"`
}

// LoginResponse is a representation of a login response
type LoginResponse struct {
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

// RegisterUser is a representation of a request body to register a user
type RegisterUser struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	DeviceID        string `json:"device_id,omitempty"`
}
