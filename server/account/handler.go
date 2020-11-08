package account

import (
	"encoding/json"
	"errors"
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

// RefreshToken is the endpoint for refreshing an access token
func RefreshToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	switch req.Method {
	case http.MethodPost:

		refreshToken := refreshT{}
		err := json.NewDecoder(req.Body).Decode(&refreshToken)
		if err != nil && err.Error() == "EOF" {
			utils.InvalidJsonResp(w, errors.New("No input was passed"))
			return
		} else if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		err, _ = utils.ValidateInput(refreshToken)
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		email, err := auth.CheckRefreshToken(refreshToken.RefreshToken)
		if err != nil && "Token is expired" == err.Error() {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"error" : "Token has expired, please login"}`)
			return
		} else if err != nil {
			auth.UnauthorizedResponse(w)
			return
		}

		accessToken, err := auth.NewAccessToken(fmt.Sprint(email))
		if err != nil {
			utils.InternalIssues(w)
			return
		}

		data := accessT{AccessToken: accessToken}
		resp := utils.SuccessResponse{
			Message: "success",
			Data:    data,
		}

		jsonResp, err := json.Marshal(resp)
		if err != nil {
			utils.InternalIssues(w)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", auth.FrontEndOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		fmt.Fprint(w, string(jsonResp))
		return
	default:
		utils.MethodNotAllowedResponse(w)
		return
	}
}

// UserProfile is the endpoint that is used for user details
func UserProfile(w http.ResponseWriter, req *http.Request) {
	const userKey auth.Key = "user"
	user, ok := req.Context().Value(userKey).(auth.User)
	if !ok {
		utils.InternalIssues(w)
		return
	}

	switch req.Method {
	case http.MethodGet:

		data := loginResponse{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			PhoneNumber: user.PhoneNumber,
			UserAddress: user.UserAddress,
			IsActive:    user.IsActive,
			DateJoined:  user.DateJoined,
			LastLogin:   user.LastLogin,
		}
		successR := utils.SuccessResponse{
			Message: "success",
			Data:    data,
		}
		jsonResp, err := json.Marshal(successR)
		if err != nil {
			utils.InternalIssues(w)
			return
		}
		fmt.Fprint(w, string(jsonResp))
		return

	case http.MethodPut:

		userProfile := auth.User{}
		userProfile.Email = user.Email
		err := json.NewDecoder(req.Body).Decode(&userProfile)
		if err != nil && err.Error() == "EOF" {
			utils.InvalidJsonResp(w, errors.New("No input was passed"))
			return
		} else if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		err, _ = utils.ValidateInput(userProfile)
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

		if userProfile.UserAddress == "" && userProfile.PhoneNumber == "" && userProfile.FirstName == "" {
			utils.InvalidJsonResp(w, errors.New("No input field was passed"))
			return
		}

		err = UpdateUserRecord(db.Db, userProfile)
		if err != nil {
			utils.InternalIssues(w)
			return
		}
		user.FirstName = userProfile.FirstName
		user.UserAddress = userProfile.UserAddress
		user.PhoneNumber = userProfile.PhoneNumber
		user.HashedPassword = ""
		resp := utils.SuccessResponse{
			Message: "success",
			Data:    user,
		}
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			utils.InternalIssues(w)
			return
		}
		fmt.Fprint(w, string(jsonResp))
		return

	default:
		utils.MethodNotAllowedResponse(w)
		return
	}
}
