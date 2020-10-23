package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/utils"
)

// Register User endppoint
func Register(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var (
			userPayload auth.RegisterUser
			err         error
		)

		err = json.NewDecoder(req.Body).Decode(&userPayload)
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		err, aboveOneField := utils.ValidateInput(userPayload)
		if aboveOneField {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
			return
		}
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}
		user := auth.User{
			Email:      userPayload.Email,
			DateJoined: time.Now(),
			LastLogin:  time.Now(),
			IsActive:   true,
		}

		user.HashedPassword, err = auth.HashPassword(userPayload.Password)
		if err != nil {
			utils.InternalIssues(w)
			return
		}

		err = AddRecordToUserTable(db.Db, user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "User already exists, please login"}`)
			return
		}

		accessToken, refreshToken, err := auth.GenerateToken(user.Email)
		if err != nil {
			utils.InternalIssues(w)
			return
		}

		logRes := loginResponse{
			ID:           user.ID,
			Email:        user.Email,
			FirstName:    user.FirstName,
			IsActive:     user.IsActive,
			DateJoined:   user.DateJoined,
			LastLogin:    user.LastLogin,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		successResp := utils.SuccessResponse{
			Message: "success",
			Data:    logRes,
		}
		jsonResp, err := json.Marshal(successResp)
		if err != nil {
			utils.InternalIssues(w)
		}

		fmt.Fprint(w, string(jsonResp))
		return

	default:
		utils.MethodNotAllowedResponse(w)
		return
	}
}

// Login User Endpoint
func Login(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var loginDetails loginInfo

		err := json.NewDecoder(req.Body).Decode(&loginDetails)
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}
		err, aboveOneField := utils.ValidateInput(loginDetails)
		if aboveOneField {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
			return
		}
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		customer, err := GetUserLogin(db.Db, loginDetails.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "User does not exist"}`)
			return
		}

		err = auth.CheckPasswordHash(loginDetails.Password, customer.HashedPassword)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Email/Password is incorrect"}`)
			return
		}

		accessToken, refreshToken, err := auth.GenerateToken(customer.Email)
		if err != nil {
			utils.InternalIssues(w)
			return
		}

		logRes := loginResponse{
			ID:           customer.ID,
			Email:        customer.Email,
			FirstName:    customer.FirstName,
			IsActive:     customer.IsActive,
			DateJoined:   customer.DateJoined,
			LastLogin:    customer.LastLogin,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		successResp := utils.SuccessResponse{
			Message: "success",
			Data:    logRes,
		}
		jsonResp, err := json.Marshal(successResp)
		if err != nil {
			utils.InternalIssues(w)
		}

		fmt.Fprint(w, string(jsonResp))
		return

	default:
		utils.MethodNotAllowedResponse(w)
	}
}
