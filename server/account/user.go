package account

import (
	"encoding/json"
	"net/http"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/utils"
)

func Register(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		var (
			user auth.User
			err  error
		)

		err = json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			utils.InvalidJsonResp(w, err)
			return
		}

	default:
		utils.MethodNotAllowedResponse(w)
		return
	}
}
