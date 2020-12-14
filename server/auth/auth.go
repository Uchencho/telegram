package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Key is a representation of the expected type to decode the context passed
type Key string

var (
	signingKey        = []byte(os.Getenv("SIGNING_KEY"))
	refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))

	// FrontEndOrigin is the origin of the consumer
	FrontEndOrigin string = os.Getenv("FRONT_ORIGIN")
)

// HashPassword Hashes a password
func HashPassword(password string) (string, error) {
	if len(password) < 1 {
		return "", errors.New("Cant hash an empty string")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

// CheckPasswordHash Checks the password and the hash, returns a non nil error if not the same
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// GenerateToken Generates an acess and refresh token on authentication
func GenerateToken(email string) (string, string, error) {

	if len(email) == 0 {
		return "", "", errors.New("Can't generate token for an invalid email")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = email
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()

	accessToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshClaims["authorized"] = true
	refreshClaims["client"] = email
	refreshClaims["exp"] = time.Now().Add(time.Hour * 8).Unix()

	refreshString, err := refreshToken.SignedString(refreshSigningKey)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshString, nil
}

// Checks the accessToken for authenticity
func checkAccessToken(accessToken string) (interface{}, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("An error occured")
		}
		return signingKey, nil
	})

	if err != nil {
		return "", err
	}

	if cliams, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return cliams["client"], nil
	}
	return "", errors.New("Credentials not provided")
}

// CheckRefreshToken checks if the refresh token passed is correct
func CheckRefreshToken(refreshToken string) (interface{}, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("An error occurred")
		}
		return refreshSigningKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["client"], nil
	}
	return "", errors.New("Credentials not provided")
}

// NewAccessToken Creates a new access token only
func NewAccessToken(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = email
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()

	accessToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
