package result

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetLabelsByEnv(t *testing.T) {
	t.Run("WithoutEnv", func(t *testing.T) {
		r1 := New("foo", nil)
		require.Empty(t, r1.Labels)
	})

	require.NoError(t, os.Setenv(EnvTestLabelPrefix+"universe.answer", "42"))

	t.Run("WithEnv", func(t *testing.T) {
		r2 := New("bar", nil)
		require.Equal(t, "42", r2.Labels["universe.answer"])
	})
}
