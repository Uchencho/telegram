package account

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/database"
	"github.com/Uchencho/telegram/server/utils"
)

// Register User endppoint
func Register(insertRecord database.AddUserToDBFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:

			var (
				userPayload auth.RegisterUser
				err         error
			)

			err = json.NewDecoder(req.Body).Decode(&userPayload)
			if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}

			aboveOneField, err := utils.ValidateInput(userPayload)
			if aboveOneField {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
				return
			}
			if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}
			user := database.User{
				Email:      userPayload.Email,
				DateJoined: time.Now(),
				LastLogin:  time.Now(),
				IsActive:   true,
			}

			user.HashedPassword, err = auth.HashPassword(userPayload.Password)
			if err != nil {
				utils.InternalIssues(w, err)
				return
			}

			err = insertRecord(user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error" : "User already exists, please login"}`)
				return
			}

			accessToken, refreshToken, err := auth.GenerateToken(user.Email)
			if err != nil {
				utils.InternalIssues(w, err)
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
				utils.InternalIssues(w, err)
			}

			fmt.Fprint(w, string(jsonResp))
			return

		default:
			utils.MethodNotAllowedResponse(w)
			return
		}
	}
}

// Login User Endpoint
func Login(getLoginDetails database.RetrieveUserLoginDetailsFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:

			var loginDetails loginInfo

			err := json.NewDecoder(req.Body).Decode(&loginDetails)
			if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}
			aboveOneField, err := utils.ValidateInput(loginDetails)
			if aboveOneField {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error" : "Invalid Payload"}`)
				return
			}
			if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}

			customer, err := getLoginDetails(loginDetails.Email)
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
				utils.InternalIssues(w, err)
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
				utils.InternalIssues(w, err)
			}

			fmt.Fprint(w, string(jsonResp))
			return

		default:
			utils.MethodNotAllowedResponse(w)
		}
	}
}

// RefreshToken is the endpoint for refreshing an access token
func RefreshToken() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application-json")
		switch req.Method {
		case http.MethodPost:

			refreshToken := refreshT{}
			err := json.NewDecoder(req.Body).Decode(&refreshToken)
			if err != nil && err.Error() == "EOF" {
				utils.InvalidJSONResp(w, errors.New("No input was passed"))
				return
			} else if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}

			_, err = utils.ValidateInput(refreshToken)
			if err != nil {
				utils.InvalidJSONResp(w, err)
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
				utils.InternalIssues(w, err)
				return
			}

			data := accessT{AccessToken: accessToken}
			resp := utils.SuccessResponse{
				Message: "success",
				Data:    data,
			}

			jsonResp, err := json.Marshal(resp)
			if err != nil {
				utils.InternalIssues(w, err)
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
}

// UserProfile is the endpoint that is used for user details
func UserProfile(updateUser database.UpdateUserDetailsFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user := utils.GetUserFromRequestContext(w, req)

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
				utils.InternalIssues(w, err)
				return
			}
			fmt.Fprint(w, string(jsonResp))
			return

		case http.MethodPut:

			userProfile := database.User{}
			userProfile.Email = user.Email
			err := json.NewDecoder(req.Body).Decode(&userProfile)
			if err != nil && err.Error() == "EOF" {
				utils.InvalidJSONResp(w, errors.New("No input was passed"))
				return
			} else if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}

			_, err = utils.ValidateInput(userProfile)
			if err != nil {
				utils.InvalidJSONResp(w, err)
				return
			}

			if userProfile.UserAddress == "" && userProfile.PhoneNumber == "" && userProfile.FirstName == "" {
				utils.InvalidJSONResp(w, errors.New("No input field was passed"))
				return
			}

			err = updateUser(userProfile)
			if err != nil {
				utils.InternalIssues(w, err)
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
				utils.InternalIssues(w, err)
				return
			}
			fmt.Fprint(w, string(jsonResp))
			return

		default:
			utils.MethodNotAllowedResponse(w)
			return
		}
	}
}
