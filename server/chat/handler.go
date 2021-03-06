package chat

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Uchencho/telegram/server/database"
	"github.com/Uchencho/telegram/server/utils"
)

// History is the handler that returns the list of users a specific user has contacted
func History(getThreads database.RetrieveUserThreadsFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		user := utils.GetUserFromRequestContext(w, req)

		switch req.Method {
		case http.MethodGet:
			threads, err := getThreads(user)
			if err != nil {
				utils.InternalIssues(w, err)
				return
			}
			resp := utils.GenericResponse{
				Message: "success",
			}

			if len(threads) == 0 {
				resp.Data = []database.Thread{}
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
}

// MessageHistory is a handler that return all the messages that was sent in a thread
func MessageHistory(getMessages database.RetrieveMessagesFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:

			threadPayload := database.Thread{}
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

			chatList, err := getMessages(threadPayload.ID)
			if err != nil {
				utils.InternalIssues(w, err)
				return
			}

			resp := utils.GenericResponse{
				Message: "success",
			}
			if len(chatList) == 0 {
				resp.Data = []database.Message{}
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
}
