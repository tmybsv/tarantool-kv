package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tmybsv/tarantool-kv/internal/storage"
)

type MockKVStorage struct {
	mock.Mock
}

func (m *MockKVStorage) Set(key string, value any) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockKVStorage) Update(key string, value any) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockKVStorage) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockKVStorage) Get(key string) (any, error) {
	args := m.Called(key)
	return args.Get(0), args.Error(1)
}

func TestKV_Set(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(*MockKVStorage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful set",
			requestBody: map[string]any{
				"key":   "test-key",
				"value": "test-value",
			},
			mockSetup: func(storage *MockKVStorage) {
				storage.On("Set", "test-key", "test-value").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "empty key",
			requestBody: map[string]any{
				"key":   "",
				"value": "test-value",
			},
			mockSetup:      func(storage *MockKVStorage) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(storage *MockKVStorage) {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockKVStorage{}
			tt.mockSetup(mockStorage)

			handler := NewKV(log, mockStorage, "/api/v1/kv")

			var body bytes.Buffer
			if tt.name == "invalid JSON" {
				body.WriteString(tt.requestBody.(string))
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/kv", &body)
			w := httptest.NewRecorder()

			handler.Set(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestKV_Get(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	tests := []struct {
		name           string
		key            string
		mockSetup      func(*MockKVStorage)
		expectedStatus int
	}{
		{
			name: "successful get",
			key:  "test-key",
			mockSetup: func(storage *MockKVStorage) {
				storage.On("Get", "test-key").Return("test-value", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "key not found",
			key:  "nonexistent-key",
			mockSetup: func(ms *MockKVStorage) {
				ms.On("Get", "nonexistent-key").Return(nil, storage.ErrKeyNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockKVStorage{}
			tt.mockSetup(mockStorage)

			handler := NewKV(logger, mockStorage, "/api/v1/kv")

			req := httptest.NewRequest(http.MethodGet, "/api/v1/kv/"+tt.key, nil)
			req.SetPathValue("key", tt.key)
			w := httptest.NewRecorder()

			handler.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockStorage.AssertExpectations(t)
		})
	}
}
