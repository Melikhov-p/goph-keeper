package config_test

import (
	"testing"

	"github.com/Melikhov-p/goph-keeper/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, err := config.Load()
		require.NoError(t, err)
		require.NotNil(t, cfg)
	})
}
