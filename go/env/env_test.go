package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	_env "gitlab.com/fynbos/env"
)

func TestGetEnv(t *testing.T) {
	t.Run("Defaults to prod", func(st *testing.T) {
		os.Unsetenv("FYNBOS_ENV")

		assert.Equal(st, "prod", _env.GetEnv())
	})

	t.Run("panics if not valid env", func(st *testing.T) {
		defer func() {
			e := recover()
			if e == nil {
				st.Fatal("Should have panicked")
			} else {
				assert.Equal(st, "Invalid env=prod2", e)
			}
		}()

		os.Setenv("FYNBOS_ENV", "prod2")
		_ = _env.GetEnv()
	})
}

func TestLocal(t *testing.T) {
	os.Setenv("FYNBOS_ENV", "local")

	env := _env.GetEnv()
	assert.Equal(t, "local", env)
	assert.True(t, _env.IsLocal())
	assert.False(t, _env.IsProd())
	assert.False(t, _env.IsDev())
	assert.False(t, _env.IsSandbox())
}

func TestDev(t *testing.T) {
	os.Setenv("FYNBOS_ENV", "dev")

	env := _env.GetEnv()
	assert.Equal(t, "dev", env)
	assert.True(t, _env.IsDev())
	assert.False(t, _env.IsProd())
	assert.False(t, _env.IsLocal())
	assert.False(t, _env.IsSandbox())
}

func TestSandbox(t *testing.T) {
	os.Setenv("FYNBOS_ENV", "sandbox")

	env := _env.GetEnv()
	assert.Equal(t, "sandbox", env)
	assert.True(t, _env.IsSandbox())
	assert.False(t, _env.IsProd())
	assert.False(t, _env.IsDev())
	assert.False(t, _env.IsLocal())
}

func TestProd(t *testing.T) {
	os.Setenv("FYNBOS_ENV", "prod")

	env := _env.GetEnv()
	assert.Equal(t, "prod", env)
	assert.True(t, _env.IsProd())
	assert.False(t, _env.IsSandbox())
	assert.False(t, _env.IsDev())
	assert.False(t, _env.IsLocal())
}
