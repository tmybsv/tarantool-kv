package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type successResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Details any    `json:"details,omitempty"`
}

type errResponse struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Details string `json:"details"`
}

func writeJSONErr(log *slog.Logger, w http.ResponseWriter, statusCode int, details string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	resp := errResponse{
		Status:  "error",
		Code:    statusCode,
		Details: details,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode error response", slog.String("error", err.Error()))
	}
}

func writeJSONSuccess(log *slog.Logger, w http.ResponseWriter, statusCode int, details any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	resp := successResponse{
		Status:  "success",
		Code:    statusCode,
		Details: details,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode success response", slog.String("error", err.Error()))
	}
}
