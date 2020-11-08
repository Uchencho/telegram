package chat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/utils"
)

// History is the handler that returns the chat history of a specific user
func History(w http.ResponseWriter, req *http.Request) {
	const userKey auth.Key = "user"
	user, ok := req.Context().Value(userKey).(auth.User)
	if !ok {
		utils.InternalIssues(w)
		return
	}

	switch req.Method {
	case http.MethodGet:
		threads, err := chatThreadsByUser(db.Db, user)
		if err != nil {
			utils.InternalIssues(w)
			return
		}
		resp := utils.SuccessResponse{
			Message: "success",
		}

		if len(threads) == 0 {
			resp.Data = thread{}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				utils.InternalIssues(w)
				return
			}
			fmt.Fprint(w, string(jsonResp))
			return
		}
		resp.Data = threads
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
