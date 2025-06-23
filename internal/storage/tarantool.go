package storage

import (
	"encoding/json"
	"errors"

	"github.com/tarantool/go-tarantool/v2"
)

var (
	// ErrInvalidDataFormat is returned when the data format is invalid.
	ErrInvalidDataFormat = errors.New("invalid data format")
	// ErrKeyNotFound is returned when the key is not found.
	ErrKeyNotFound = errors.New("key not found")
)

// Tarantool is a storage implementation that uses Tarantool as a backend.
type Tarantool struct {
	conn  *tarantool.Connection
	space string
	index string
}

// NewTarantool creates a new Tarantool storage.
func NewTarantool(conn *tarantool.Connection, space, index string) *Tarantool {
	return &Tarantool{
		conn:  conn,
		space: space,
		index: index,
	}
}

// Set stores the value for the given key.
func (s *Tarantool) Set(key string, value any) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	req := tarantool.NewInsertRequest(s.space).Tuple([]any{key, string(jsonValue)})
	if _, err := s.conn.Do(req).Get(); err != nil {
		return err
	}

	return nil
}

// Update updates the value for the given key.
func (s *Tarantool) Update(key string, value any) error {
	if _, err := s.Get(key); err != nil {
		return err
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	req := tarantool.NewReplaceRequest(s.space).Tuple([]any{key, string(jsonValue)})
	if _, err := s.conn.Do(req).Get(); err != nil {
		return err
	}

	return nil
}

// Get retrieves the value for the given key.
func (s *Tarantool) Get(key string) (any, error) {
	req := tarantool.NewSelectRequest(s.space).Index(s.index).Key(tarantool.StringKey{S: key})
	resp, err := s.conn.Do(req).Get()
	if err != nil {
		return nil, err
	}

	if len(resp) == 0 {
		return nil, ErrKeyNotFound
	}

	row, ok := resp[0].([]any)
	if !ok {
		return nil, errors.New("cannot retrieve response row")
	}

	if len(row) != 2 {
		return nil, ErrInvalidDataFormat
	}

	valueBytes := []byte(row[1].(string))
	var value any
	if err := json.Unmarshal(valueBytes, &value); err != nil {
		return nil, err
	}

	return value, nil
}

// Delete deletes the value for the given key.
func (s *Tarantool) Delete(key string) error {
	req := tarantool.NewDeleteRequest(s.space).Index(s.index).Key(tarantool.StringKey{S: key})
	if _, err := s.conn.Do(req).Get(); err != nil {
		return err
	}
	return nil
}
