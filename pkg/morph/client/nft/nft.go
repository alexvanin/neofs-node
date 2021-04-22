package nft

import (
	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/encoding/address"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neofs-api-go/pkg/owner"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
	"github.com/pkg/errors"
)

type (
	// Fetcher is a structure that implements function to fetch data about
	// NFT and its holder.
	Fetcher struct {
		cli *client.Client
	}
)

const ownerOfMethod = "ownerOf" // defined in NEP-11

// NewFetcher is a constructor for NFT information fetcher.
func NewFetcher(cli *client.Client) *Fetcher {
	return &Fetcher{
		cli: cli,
	}
}

// Owner returns owner ID of NFT holder produced by contract.
func (f Fetcher) Owner(contract util.Uint160, tokenID []byte) (*owner.ID, error) {
	items, err := f.cli.TestInvoke(contract, ownerOfMethod, tokenID)
	if err != nil {
		return nil, errors.Wrap(err, "test invoke error")
	}

	if ln := len(items); ln != 1 {
		return nil, errors.Wrapf(err, "expected 1 stack item, got %d", ln)
	}

	u160bytes, err := client.BytesFromStackItem(items[0])
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse bytes from stack item")
	}

	u160, err := util.Uint160DecodeBytesBE(u160bytes)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode Uint160 BE value")
	}

	return uint160ToOwnerID(u160)
}

func uint160ToOwnerID(u util.Uint160) (*owner.ID, error) {
	addr, err := base58.Decode(address.Uint160ToString(u))
	if err != nil {
		return nil, errors.Wrap(err, "can't decode wallet address")
	}

	w := new(owner.NEO3Wallet)
	copy(w.Bytes(), addr)

	return owner.NewIDFromNeo3Wallet(w), nil
}
