package network

import (
	"strings"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/require"
)

func TestAddress_NetAddr(t *testing.T) {
	ip := "127.0.0.1"
	port := "8080"

	ma, err := multiaddr.NewMultiaddr(strings.Join([]string{
		"/ip4",
		ip,
		"tcp",
		port,
	}, "/"))

	require.NoError(t, err)

	addr, err := AddressFromString(ma.String())
	require.NoError(t, err)

	netAddr, err := addr.IPAddrString()
	require.NoError(t, err)
	require.Equal(t, ip+":"+port, netAddr)
}

func TestAddress_HostAddrString(t *testing.T) {
	t.Run("valid addresses", func(t *testing.T) {
		testcases := []struct {
			ma  multiaddr.Multiaddr
			exp string
		}{
			{buildMultiaddr("/dns4/neofs.bigcorp.com/tcp/8080", t), "neofs.bigcorp.com:8080"},
			{buildMultiaddr("/ip4/172.16.14.1/tcp/8080", t), "172.16.14.1:8080"},
		}

		for _, testcase := range testcases {
			addr := Address{testcase.ma}

			got, err := addr.HostAddrString()
			require.NoError(t, err)

			require.Equal(t, testcase.exp, got)
		}
	})

	t.Run("invalid addresses", func(t *testing.T) {
		testcases := []multiaddr.Multiaddr{
			buildMultiaddr("/tcp/8080", t),
		}

		for _, testcase := range testcases {
			addr := Address{testcase}
			_, err := addr.HostAddrString()
			require.Error(t, err)
		}
	})
}

func buildMultiaddr(s string, t *testing.T) multiaddr.Multiaddr {
	ma, err := multiaddr.NewMultiaddr(s)
	require.NoError(t, err)
	return ma
}
