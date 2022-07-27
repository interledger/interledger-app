package env_test

import (
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	_env "gitlab.com/fynbos/env"
)

func TestFynbosEnv(t *testing.T) {
	t.Run("Deafults to prod", func(st *testing.T) {
		env, err := _env.NewFynbosEnv("")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, env.IsProd())
		assert.False(t, env.IsSandbox())
		assert.False(t, env.IsDev())
		assert.False(t, env.IsTesting())
	})

	t.Run("IsSandbox", func(st *testing.T) {
		env, err := _env.NewFynbosEnv("sandbox")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, env.IsSandbox())
		assert.False(t, env.IsProd())
		assert.False(t, env.IsDev())
		assert.False(t, env.IsTesting())
	})

	t.Run("IsProd", func(st *testing.T) {
		env, err := _env.NewFynbosEnv("prod")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, env.IsProd())
		assert.False(t, env.IsSandbox())
		assert.False(t, env.IsDev())
		assert.False(t, env.IsTesting())
	})

	t.Run("IsDev", func(st *testing.T) {
		env, err := _env.NewFynbosEnv("dev")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, env.IsDev())
		assert.False(t, env.IsSandbox())
		assert.False(t, env.IsProd())
		assert.False(t, env.IsTesting())
	})

	t.Run("IsTesting", func(st *testing.T) {
		env, err := _env.NewFynbosEnv("testing")
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, env.IsTesting())
		assert.False(t, env.IsSandbox())
		assert.False(t, env.IsProd())
		assert.False(t, env.IsDev())
	})

	t.Run("Returns error for invalid environment", func(st *testing.T) {
		envName := faker.Name()
		env, err := _env.NewFynbosEnv(envName)

		assert.EqualError(t, err, "Invalid env="+envName)
		assert.Nil(t, env)
	})
}
