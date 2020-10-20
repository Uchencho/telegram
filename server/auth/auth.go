package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUser struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	DeviceID        string `json:"device_id,omitempty"`
}

type User struct {
	ID             uint      `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
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

const (
	frontEndOrigin string = "*"
)

var (
	signingKey        = []byte(os.Getenv("SIGNING_KEY"))
	refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))
)

// Hashes a password
func HashPassword(password string) (string, error) {
	if len(password) < 1 {
		return "", errors.New("Cant hash an empty string")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

// Checks the password and the hash, returns a non nil error if not the same
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// Generates an acess and refresh token on authentication
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

// Middleware that checks if a token was passed
func BasicToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Header["Authorization"] != nil {
			if len(strings.Split(r.Header["Authorization"][0], " ")) < 2 {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token format"}`)
				return
			}

			accessToken := strings.Split(r.Header["Authorization"][0], " ")[1]
			basic_token := os.Getenv("BASIC_TOKEN")
			if basic_token == accessToken {

				//Allow CORS here By or specific origin
				w.Header().Set("Access-Control-Allow-Origin", frontEndOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				next.ServeHTTP(w, r)
				return
			} else {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token passed"}`)
				return
			}
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}

// Middleware that returns the details of the user
func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if req.Header["Authorization"] != nil {
			if len(strings.Split(req.Header["Authorization"][0], " ")) < 2 {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error" : "Invalid token format"}`)
				return
			}

			accessToken := strings.Split(req.Header["Authorization"][0], " ")[1]
			_, err := checkAccessToken(accessToken)
			if err != nil && "Token is expired" == err.Error() {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "Token has expired please login"}`)
				return
			} else if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "Invalid token"}`)
				return
			}

			// Retrieve the user and pass it into a context, to do!

			//Allow CORS here By or specific origin
			w.Header().Set("Access-Control-Allow-Origin", frontEndOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			next.ServeHTTP(w, req)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}
