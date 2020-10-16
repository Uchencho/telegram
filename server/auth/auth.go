package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	ID             uint           `json:"id"`
	Email          string         `json:"email"`
	HashedPassword string         `json:"password"`
	FirstName      sql.NullString `json:"first_name"`
	PhoneNumber    sql.NullString `json:"phone_number"`
	UserAddress    sql.NullString `json:"user_address"`
	IsActive       bool           `json:"is_active"`
	DateJoined     time.Time      `json:"date_joined"`
	LastLogin      sql.NullTime   `json:"last_login"`
	Longitude      sql.NullString `json:"longitude"`
	Latitude       sql.NullString `json:"latitude"`
	DeviceID       sql.NullString `json:"device_id"`
}

var (
	signingKey = []byte(os.Getenv("SIGNING_KEY"))
	// refreshSigningKey = []byte(os.Getenv("REFRESH_SIGNING_KEY"))
)

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

			next.ServeHTTP(w, req)
			return
		}
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error" : "Token not passed"}`)
	})
}
