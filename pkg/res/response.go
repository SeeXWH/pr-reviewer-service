package res

import (
	"encoding/json"
	"net/http"
)

type errorWrapper struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	errResponse := errorWrapper{
		Error: errorDetail{
			Code:    code,
			Message: message,
		},
	}
	JSON(w, status, errResponse)
}
