package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Uchencho/telegram/server/database"
)

// BasicToken is a Middleware that checks if a token was passed
func BasicToken(next http.Handler) http.HandlerFunc {
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
func UserMiddleware(getUserDetails database.RetrieveUserLoginDetailsFunc, next http.Handler) http.HandlerFunc {
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
			user, err := getUserDetails(fmt.Sprint(email))
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
func WebsocketAuthMiddleware(getUserDetails database.RetrieveUserLoginDetailsFunc, next http.Handler) http.HandlerFunc {
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
			user, err := getUserDetails(fmt.Sprint(email))
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
