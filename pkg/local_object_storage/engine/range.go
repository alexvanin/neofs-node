package engine

import (
	"errors"

	objectSDK "github.com/nspcc-dev/neofs-api-go/pkg/object"
	"github.com/nspcc-dev/neofs-node/pkg/core/object"
	meta "github.com/nspcc-dev/neofs-node/pkg/local_object_storage/metabase/v2"
	"github.com/nspcc-dev/neofs-node/pkg/local_object_storage/shard"
	"go.uber.org/zap"
)

// RngPrm groups the parameters of GetRange operation.
type RngPrm struct {
	off, ln uint64

	addr *objectSDK.Address
}

// GetRes groups resulting values of GetRange operation.
type RngRes struct {
	obj *object.Object
}

// WithAddress is a GetRng option to set the address of the requested object.
//
// Option is required.
func (p *RngPrm) WithAddress(addr *objectSDK.Address) *RngPrm {
	if p != nil {
		p.addr = addr
	}

	return p
}

// WithPayloadRange is a GetRange option to set range of requested payload data.
//
// Missing an option or calling with zero length is equivalent
// to getting the full payload range.
func (p *RngPrm) WithPayloadRange(off, ln uint64) *RngPrm {
	if p != nil {
		p.off, p.ln = off, ln
	}

	return p
}

// Object returns the requested object part.
//
// Instance payload contains the requested range of the original object.
func (r *RngRes) Object() *object.Object {
	return r.obj
}

// GetRange reads part of an object from local storage.
//
// Returns any error encountered that
// did not allow to completely read the object part.
//
// Returns ErrNotFound if requested object is missing in local storage.
func (e *StorageEngine) GetRange(prm *RngPrm) (*RngRes, error) {
	var (
		obj *object.Object

		alreadyRemoved = false
	)

	shPrm := new(shard.RngPrm).
		WithAddress(prm.addr).
		WithRange(prm.off, prm.ln)

	e.iterateOverSortedShards(prm.addr, func(_ int, sh *shard.Shard) (stop bool) {
		res, err := sh.GetRange(shPrm)
		if err != nil {
			switch {
			case errors.Is(err, object.ErrNotFound):
				return false // ignore, go to next shard
			case errors.Is(err, meta.ErrAlreadyRemoved):
				alreadyRemoved = true

				return true // stop, return it back
			default:
				// TODO: smth wrong with shard, need to be processed, but
				// still go to next shard
				e.log.Warn("could not get object from shard",
					zap.Stringer("shard", sh.ID()),
					zap.String("error", err.Error()),
				)

				return false
			}
		}

		obj = res.Object()

		return true
	})

	if obj == nil {
		if alreadyRemoved {
			return nil, meta.ErrAlreadyRemoved
		}

		return nil, object.ErrNotFound
	}

	return &RngRes{
		obj: obj,
	}, nil
}

// GetRange reads object payload range from local storage by provided address.
func GetRange(storage *StorageEngine, addr *objectSDK.Address, rng *objectSDK.Range) ([]byte, error) {
	res, err := storage.GetRange(new(RngPrm).
		WithAddress(addr).
		WithPayloadRange(rng.GetOffset(), rng.GetLength()),
	)
	if err != nil {
		return nil, err
	}

	return res.Object().Payload(), nil
}