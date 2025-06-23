package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	configContent := `
env: test
tarantool:
  host: localhost
  port: 3301
  user: guest
  password: ""
  timeout: 5s
  kv_space: kv
  kv_index: primary
http:
  port: 8080
  timeout: 30s
  kv_base_path: /api/kv
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	originalPath := os.Getenv("KV_CONFIG_PATH")
	os.Setenv("KV_CONFIG_PATH", tmpFile.Name())
	defer os.Setenv("KV_CONFIG_PATH", originalPath)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "test", cfg.Env)
	assert.Equal(t, "localhost", cfg.Tarantool.Host)
	assert.Equal(t, 3301, cfg.Tarantool.Port)
	assert.Equal(t, "guest", cfg.Tarantool.User)
	assert.Equal(t, "", cfg.Tarantool.Password)
	assert.Equal(t, 5*time.Second, cfg.Tarantool.Timeout)
	assert.Equal(t, "kv", cfg.Tarantool.KVSpace)
	assert.Equal(t, "primary", cfg.Tarantool.KVIndex)
	assert.Equal(t, 8080, cfg.HTTP.Port)
	assert.Equal(t, 30*time.Second, cfg.HTTP.Timeout)
	assert.Equal(t, "/api/kv", cfg.HTTP.KVBasePath)
}

func TestLoad_FileNotFound(t *testing.T) {
	os.Setenv("KV_CONFIG_PATH", "nonexistent-file.yml")
	defer os.Unsetenv("KV_CONFIG_PATH")

	_, err := Load()
	assert.Error(t, err)
}

func TestMustLoad_Panic(t *testing.T) {
	os.Setenv("KV_CONFIG_PATH", "nonexistent-file.yml")
	defer os.Unsetenv("KV_CONFIG_PATH")

	assert.Panics(t, func() {
		MustLoad()
	})
}
