package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/database"
	"github.com/go-playground/validator"
)

var (
	validate = validator.New()
)

// GenericResponse is a representation of a successful response
type GenericResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// InvalidJSONResp is a representation of invalid json error
func InvalidJSONResp(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	res := GenericResponse{Error: fmt.Sprintf("error in decoding json: %s", err.Error())}
	jsonResp, err := json.Marshal(res)
	if err != nil {
		InternalIssues(w, err)
	}
	fmt.Fprint(w, string(jsonResp))
}

// MethodNotAllowedResponse indicates when a request method is not allowed
func MethodNotAllowedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	res := GenericResponse{Error: "Method Not allowed"}
	jsonResp, err := json.Marshal(res)
	if err != nil {
		InternalIssues(w, err)
	}
	fmt.Fprint(w, string(jsonResp))
}

// InternalIssues denotes an internal unwxpected issue occured
func InternalIssues(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	res := GenericResponse{Error: "Something went wrong"}
	jsonResp, err := json.Marshal(res)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(jsonResp))
}

// NotAvailabe is a handler that handles invalid urls
func NotAvailabe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	res := GenericResponse{Error: "Resource not found"}
	jsonResp, err := json.Marshal(res)
	if err != nil {
		InternalIssues(w, err)
	}
	fmt.Fprint(w, string(jsonResp))
}

// ValidateInput validates the input struct
func ValidateInput(object interface{}) (bool, error) {

	err := validate.Struct(object)
	if err != nil {

		//Validation syntax is invalid
		if _, ok := err.(*validator.InvalidValidationError); ok {
			log.Println(err)
			return false, err
		}

		if len(err.(validator.ValidationErrors)) > 1 {
			log.Println("Error is more than one")
			return true, err
		}

		for _, err := range err.(validator.ValidationErrors) {

			// Retrieve json field
			reflectedValue := reflect.ValueOf(object)
			field, _ := reflectedValue.Type().FieldByName(err.StructField())

			var name string
			if name = field.Tag.Get("json"); name == "" {
				name = strings.ToLower(err.StructField())
			}

			switch err.Tag() {
			case "required":
				return false, fmt.Errorf("%s is required", name)
			case "email":
				return false, fmt.Errorf("%s should be a valid email address", name)
			case "eqfield":
				return false, fmt.Errorf("%s should be the same as %s", name, err.Param())
			default:
				return false, fmt.Errorf("%s is Invalid", name)
			}
		}
		return false, err
	}
	return false, nil
}

// GetUserFromRequestContext retrieves the user details from the request context
func GetUserFromRequestContext(w http.ResponseWriter, req *http.Request) database.User {
	const userKey auth.Key = "user"
	user, ok := req.Context().Value(userKey).(database.User)
	if !ok {
		InternalIssues(w, errors.New("Cannot decode user details from middleware"))
		return database.User{}
	}
	return user
}

// BadRequest indicates when a bad request occurs
func BadRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	res := GenericResponse{Error: err.Error()}
	jsonResp, err := json.Marshal(res)
	if err != nil {
		InternalIssues(w, err)
	}
	fmt.Fprint(w, string(jsonResp))
}
