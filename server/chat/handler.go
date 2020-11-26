package chat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/utils"
)

// History is the handler that returns the list of users a specific user has contacted
func History(w http.ResponseWriter, req *http.Request) {
	user := utils.GetUserFromRequestContext(w, req)

	switch req.Method {
	case http.MethodGet:
		threads, err := chatThreadsByUser(db.Db, user)
		if err != nil {
			utils.InternalIssues(w, err)
			return
		}
		resp := utils.SuccessResponse{
			Message: "success",
		}

		if len(threads) == 0 {
			resp.Data = []Thread{}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				utils.InternalIssues(w, err)
				return
			}
			fmt.Fprint(w, string(jsonResp))
			return
		}
		resp.Data = threads
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

// MessageHistory is a handler that return all the messages that was sent in a thread
func MessageHistory(w http.ResponseWriter, req *http.Request) {

	switch req.Method {
	case http.MethodPost:

		threadPayload := Thread{}
		err := json.NewDecoder(req.Body).Decode(&threadPayload)
		if err != nil {
			utils.InvalidJSONResp(w, err)
			return
		}
		_, err = utils.ValidateInput(threadPayload)
		if err != nil {
			utils.InvalidJSONResp(w, err)
			return
		}

		chatList, err := getMessages(db.Db, threadPayload.ID)
		if err != nil {
			utils.InternalIssues(w, err)
			return
		}

		resp := utils.SuccessResponse{
			Message: "success",
		}
		if len(chatList) == 0 {
			resp.Data = []Message{}
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				utils.InternalIssues(w, err)
				return
			}
			fmt.Fprint(w, string(jsonResp))
			return
		}

		resp.Data = chatList
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
