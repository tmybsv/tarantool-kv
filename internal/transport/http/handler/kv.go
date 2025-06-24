package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/tmybsv/tarantool-kv/internal/storage"
)

// KVStorage is the contract for the KV storage.
type KVStorage interface {
	// Set sets the value for the key or an error if the key is already present.
	Set(key string, value any) error
	// Update updates the value for the key or an error if the key is not found.
	Update(key string, value any) error
	// Delete removes the key from the storage or an error if the key is not found.
	Delete(key string) error
	// Get returns the value for the key or an error if the key is not found.
	Get(key string) (any, error)
}

// KV is the HTTP handler for the KV storage.
type KV struct {
	log      *slog.Logger
	storage  KVStorage
	basePath string
}

// NewKV creates a new HTTP handler for the KV storage.
func NewKV(log *slog.Logger, storage KVStorage, basePath string) *KV {
	return &KV{
		basePath: basePath,
		storage:  storage,
		log:      log,
	}
}

// Set sets the value for the key.
func (h *KV) Set(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONErr(h.log, w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Key == "" {
		writeJSONErr(h.log, w, http.StatusBadRequest, "key cannot be empty")
		return
	}

	if err := h.storage.Set(req.Key, req.Value); err != nil {
		h.log.Error("failed to set key", slog.String("error", err.Error()))
		h.handleStorageError(w, err)
		return
	}

	writeJSONSuccess(h.log, w, http.StatusCreated, map[string]any{"key": req.Key})
}

// Get returns the value for the key.
func (h *KV) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	value, err := h.storage.Get(key)
	if err != nil {
		h.log.Error("failed to get key", slog.String("error", err.Error()))
		h.handleStorageError(w, err)
		return
	}

	writeJSONSuccess(h.log, w, http.StatusOK, map[string]any{"key": key, "value": value})
}

// Update updates the value for the key.
func (h *KV) Update(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var req struct {
		Value any `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONErr(h.log, w, http.StatusUnprocessableEntity, fmt.Sprintf("decode request: %s", err.Error()))
		return
	}

	if err := h.storage.Update(key, req.Value); err != nil {
		h.log.Error("failed to update key", slog.String("error", err.Error()))
		h.handleStorageError(w, err)
		return
	}

	writeJSONSuccess(h.log, w, http.StatusOK, map[string]any{"key": key})
}

// Delete removes the key from the storage.
func (h *KV) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := h.storage.Delete(key); err != nil {
		h.log.Error("failed to delete key", slog.String("error", err.Error()))
		h.handleStorageError(w, err)
		return
	}

	writeJSONSuccess(h.log, w, http.StatusOK, map[string]any{"key": key})
}

func (h *KV) handleStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrKeyNotFound):
		writeJSONErr(h.log, w, http.StatusNotFound, "key not found")
	case errors.Is(err, storage.ErrInvalidDataFormat):
		writeJSONErr(h.log, w, http.StatusBadGateway, "storage error")
	case errors.Is(err, storage.ErrKeyAlreadyExists):
		writeJSONErr(h.log, w, http.StatusConflict, "key already exists")
	default:
		writeJSONErr(h.log, w, http.StatusInternalServerError, "internal error")
	}
}
