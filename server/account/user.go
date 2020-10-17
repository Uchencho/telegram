package account

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/utils"
)

func Register(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var (
			userPayload auth.RegisterUser
			user        auth.User
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

		user.Email = userPayload.Email
		user.DateJoined = time.Now()
		user.LastLogin = time.Now()
		user.HashedPassword, err = auth.HashPassword(userPayload.Password)
		if err != nil {
			utils.InternalIssues(w)
			return
		}

		err = AddRecordToUserTable(db.Db, user)
		if err != nil {
			log.Printf("Error occured in adding details to db, %v", err)
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
