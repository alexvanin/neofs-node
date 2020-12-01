package engine

import (
	objectSDK "github.com/nspcc-dev/neofs-api-go/pkg/object"
	"github.com/nspcc-dev/neofs-node/pkg/local_object_storage/shard"
	"go.uber.org/zap"
)

// InhumePrm encapsulates parameters for inhume operation.
type InhumePrm struct {
	addr, tombstone *objectSDK.Address
}

// InhumeRes encapsulates results of inhume operation.
type InhumeRes struct{}

// WithTarget sets object address that should be inhumed and tombstone address
// as the reason for inhume operation.
func (p *InhumePrm) WithTarget(addr, tombstone *objectSDK.Address) *InhumePrm {
	if p != nil {
		p.addr = addr
		p.tombstone = tombstone
	}

	return p
}

// Inhume calls metabase. Inhume method to mark object as removed. It won't be
// removed physically from shard until `Delete` operation.
func (e *StorageEngine) Inhume(prm *InhumePrm) (*InhumeRes, error) {
	shPrm := new(shard.InhumePrm).WithTarget(prm.addr, prm.tombstone)

	e.iterateOverSortedShards(prm.addr, func(_ int, sh *shard.Shard) (stop bool) {
		_, err := sh.Inhume(shPrm)
		if err != nil {
			// TODO: smth wrong with shard, need to be processed
			e.log.Warn("could not inhume object in shard",
				zap.Stringer("shard", sh.ID()),
				zap.String("error", err.Error()),
			)
		}

		return false
	})

	return nil, nil
}