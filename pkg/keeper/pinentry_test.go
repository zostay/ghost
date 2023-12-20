package keeper_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zostay/ghost/pkg/keeper"
)

func TestPinEntry(t *testing.T) { //nolint:paralleltest
	if os.Getenv("GHOST_TEST_INTERACTIVE") == "" {
		t.Skip("normally don't test this because it requires feedback from the user")
	}

	response, err := keeper.GetPassword(
		"Title",
		"Description: If this looks good, type OK",
		"Prompt",
		"OK")
	require.NoError(t, err)
	assert.Equal(t, "OK", response)
}
