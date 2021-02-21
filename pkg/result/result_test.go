package result

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetLabelsByEnv(t *testing.T) {
	cleanup()
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

func TestResultDuration(t *testing.T) {
	cleanup()
	r1 := New("", nil)
	time.Sleep(time.Second)
	r1.Report(Unknown, "")
	require.Less(t, r1.StartedAt, r1.CompletedAt)

	var r2 Result
	buf := new(bytes.Buffer)
	require.NoError(t, json.NewEncoder(buf).Encode(r1))
	require.NoError(t, json.NewDecoder(buf).Decode(&r2))
	require.Less(t, r2.StartedAt, r2.CompletedAt)
}
