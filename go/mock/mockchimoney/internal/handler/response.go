package handler

import (
	"encoding/json"
	"net/http"

	"gitlab.com/fynbos/mock/mockchimoney/internal/models"
)

func (h *Handler) respondOK(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(models.APIResponse{
		Status: "success",
		Data:   data,
	})
}

func (h *Handler) respondErr(w http.ResponseWriter, statusCode int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(models.APIResponse{
		Status: "error",
		Error:  msg,
	})
}
