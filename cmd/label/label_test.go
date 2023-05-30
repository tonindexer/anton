package label

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchTonscanLabels(t *testing.T) {
	label, err := FetchTonscanLabels()
	require.Nil(t, err)

	for _, l := range label {
		t.Logf("%+v", l)
	}
}
