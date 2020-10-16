package utils

import (
	"fmt"
	"log"
	"net/http"
)

func InvalidJsonResp(w http.ResponseWriter, err error) {
	log.Printf("error in decoding json: %v", err)
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, `{"error" : "Invalid payload"}`)
}

func MethodNotAllowedResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprint(w, `{"message" : "Method Not allowed"}`)
}
