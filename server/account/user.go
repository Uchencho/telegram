package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	default:
		utils.MethodNotAllowedResponse(w)
		return
	}
}
