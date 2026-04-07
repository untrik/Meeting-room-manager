package helper

import (
	"encoding/json"
	"net/http"

	"github.com/avito-internships/test-backend-1-untrik/internal/dto"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
func WriteError(w http.ResponseWriter, status int, code dto.ErrorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	var resp dto.ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	_ = json.NewEncoder(w).Encode(resp)
}
func WriteInternalError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	var resp dto.InternalErrorResponse
	resp.Error.Code = dto.ErrorCodeInternalError
	resp.Error.Message = "internal server error"
	_ = json.NewEncoder(w).Encode(resp)
}
