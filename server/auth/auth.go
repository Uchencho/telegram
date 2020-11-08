package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Uchencho/telegram/db"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUser is a representation of a request body to register a user
type RegisterUser struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"eqfield=Password"`
	DeviceID        string `json:"device_id,omitempty"`
}

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

// Key is a representation of the expected type to decode the context passed
type Key string

const (
	// FrontEndOrigin represents the origin of the consumer of the API
	FrontEndOrigin string = "*"
)

var (
	signingKey        = []byte(os.Getenv("SIGNING_KEY"))
	refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))
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

// retrieves the user details for an endpoint to use. Middleware things
func getUser(db *sql.DB, email string) (User, error) {
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

// BasicToken is a Middleware that checks if a token was passed
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
			basicToken := os.Getenv("BASIC_TOKEN")
			if basicToken == accessToken {

				//Allow CORS here By or specific origin
				w.Header().Set("Access-Control-Allow-Origin", FrontEndOrigin)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error" : "Invalid token passed"}`)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}

// UserMiddleware is a Middleware that returns the details of the user
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
			email, err := checkAccessToken(accessToken)
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
			user, err := getUser(db.Db, fmt.Sprint(email))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "User does not exist"}`)
				return
			}

			const userKey Key = "user"
			ctx := context.WithValue(req.Context(), userKey, user)

			//Allow CORS here By or specific origin
			w.Header().Set("Access-Control-Allow-Origin", FrontEndOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			next.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}

// WebsocketAuthMiddleware retrieves the user details using authentication specific for websocket requests
func WebsocketAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		urlVals := r.URL.Query()
		if token := urlVals.Get("token"); token != "" {

			email, err := checkAccessToken(token)
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
			user, err := getUser(db.Db, fmt.Sprint(email))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error" : "User does not exist"}`)
				return
			}

			const userKey Key = "user"
			ctx := context.WithValue(r.Context(), userKey, user)

			//Allow CORS here By or specific origin
			w.Header().Set("Access-Control-Allow-Origin", FrontEndOrigin)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		log.Println("User details not passed")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}

// UnauthorizedResponse is a Forbidden utility response
func UnauthorizedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, `{"error" : "Invalid authentication credentials"}`)
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
