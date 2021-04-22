package nft

import (
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestUint160ToOwnerID(t *testing.T) {
	const (
		u160string    = "8aebbe8c9ebe48a946057d5fa7fbceb439b4f768"
		ownerIDstring = "NVUzCUvrbuWadAm6xBoyZ2U7nCmS9QBZtb"
	)

	u160, err := util.Uint160DecodeStringLE(u160string)
	require.NoError(t, err)

	id, err := uint160ToOwnerID(u160)
	require.NoError(t, err)

	require.Equal(t, ownerIDstring, id.String())
}
