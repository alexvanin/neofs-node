package netmap

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
)

// GetNetMapArgs groups the arguments
// of get network map test invoke call.
type GetNetMapArgs struct {
}

// GetSnapshotArgs groups the arguments
// of get netmap snapshot test invoke call.
type GetSnapshotArgs struct {
	diff uint64
}

// GetNetMapValues groups the stack parameters
// returned by get network map test invoke.
type GetNetMapValues struct {
	peers [][]byte
}

// EpochSnapshotArgs groups the arguments
// of snapshot by epoch test invoke call.
type EpochSnapshotArgs struct {
	epoch uint64
}

// EpochSnapshotValues groups the stack parameters
// returned by snapshot by epoch test invoke.
type EpochSnapshotValues struct {
	*GetNetMapValues
}

// GetNetMapCandidatesArgs groups the arguments
// of get network map candidates test invoke call.
type GetNetMapCandidatesArgs struct {
}

// GetNetMapCandidatesValues groups the stack parameters
// returned by get network map candidates test invoke.
type GetNetMapCandidatesValues struct {
	netmapNodes []*PeerWithState
}

func (g GetNetMapCandidatesValues) NetmapNodes() []*PeerWithState {
	return g.netmapNodes
}

// State is an enumeration of various states of the NeoFS node.
type State int64

const (
	// Undefined is unknown state.
	Undefined State = iota

	// Online is network unavailable state.
	Online

	// Offline is an active state in the network.
	Offline
)

// PeerWithState groups information about peer
// and its state in network map.
type PeerWithState struct {
	peer  []byte
	state State
}

func (ps PeerWithState) State() State {
	return ps.state
}

func (ps PeerWithState) Peer() []byte {
	return ps.peer
}

const (
	nodeInfoFixedPrmNumber = 1

	peerWithStateFixedPrmNumber = 2
)

// SetDiff sets argument for snapshot method of
// netmap contract.
func (g *GetSnapshotArgs) SetDiff(d uint64) {
	g.diff = d
}

// SetEpoch sets epoch number to get snapshot.
func (a *EpochSnapshotArgs) SetEpoch(d uint64) {
	a.epoch = d
}

// Peers return the list of peers from
// network map in a binary format.
func (g GetNetMapValues) Peers() [][]byte {
	return g.peers
}

// NetMap performs the test invoke of get network map
// method of NeoFS Netmap contract.
func (c *Client) NetMap(_ GetNetMapArgs) (*GetNetMapValues, error) {
	prms, err := c.client.TestInvoke(
		c.netMapMethod,
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform test invocation (%s): %w",
			c.netMapMethod, err)
	}

	return peersFromStackItems(prms, c.netMapMethod)
}

// Snapshot performs the test invoke of get snapshot of network map
// from NeoFS Netmap contract. Contract saves only one previous epoch,
// so all invokes with diff > 1 return error.
func (c *Client) Snapshot(a GetSnapshotArgs) (*GetNetMapValues, error) {
	prms, err := c.client.TestInvoke(
		c.snapshotMethod,
		int64(a.diff),
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform test invocation (%s): %w",
			c.netMapMethod, err)
	}

	return peersFromStackItems(prms, c.snapshotMethod)
}

// EpochSnapshot performs the test invoke of get snapshot of network map by epoch
// from NeoFS Netmap contract.
func (c *Client) EpochSnapshot(args EpochSnapshotArgs) (*EpochSnapshotValues, error) {
	prms, err := c.client.TestInvoke(
		c.epochSnapshotMethod,
		int64(args.epoch),
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform test invocation (%s): %w",
			c.epochSnapshotMethod, err)
	}

	nmVals, err := peersFromStackItems(prms, c.epochSnapshotMethod)
	if err != nil {
		return nil, err
	}

	return &EpochSnapshotValues{
		GetNetMapValues: nmVals,
	}, nil
}

func (c *Client) Candidates(_ GetNetMapCandidatesArgs) (*GetNetMapCandidatesValues, error) {
	prms, err := c.client.TestInvoke(
		c.netMapCandidatesMethod,
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform test invocation (%s): %w", c.netMapCandidatesMethod, err)
	}

	candVals, err := peersWithStateFromStackItems(prms, c.netMapCandidatesMethod)
	if err != nil {
		return nil, fmt.Errorf("could not parse contract response: %w", err)
	}

	return candVals, nil
}

func peersWithStateFromStackItems(stack []stackitem.Item, method string) (*GetNetMapCandidatesValues, error) {
	if ln := len(stack); ln != 1 {
		return nil, fmt.Errorf("unexpected stack item count (%s): %d", method, ln)
	}

	netmapNodes, err := client.ArrayFromStackItem(stack[0])
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array from stack item (%s): %w", method, err)
	}

	res := &GetNetMapCandidatesValues{
		netmapNodes: make([]*PeerWithState, 0, len(netmapNodes)),
	}

	for i := range netmapNodes {
		node, err := peerWithStateFromStackItem(netmapNodes[i])
		if err != nil {
			return nil, fmt.Errorf("could not parse stack item (Peer #%d): %w", i, err)
		}

		res.netmapNodes = append(res.netmapNodes, node)
	}

	return res, nil
}

func peerWithStateFromStackItem(prm stackitem.Item) (*PeerWithState, error) {
	prms, err := client.ArrayFromStackItem(prm)
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array (PeerWithState): %w", err)
	} else if ln := len(prms); ln != peerWithStateFixedPrmNumber {
		return nil, fmt.Errorf(
			"unexpected stack item count (PeerWithState): expected %d, has %d",
			peerWithStateFixedPrmNumber,
			ln,
		)
	}

	var res PeerWithState

	// peer
	if res.peer, err = peerInfoFromStackItem(prms[0]); err != nil {
		return nil, fmt.Errorf("could not get bytes from 'node' field of PeerWithState: %w", err)
	}

	// state
	state, err := client.IntFromStackItem(prms[1])
	if err != nil {
		return nil, fmt.Errorf("could not get int from 'state' field of PeerWithState: %w", err)
	}

	switch state {
	case 1:
		res.state = Online
	case 2:
		res.state = Offline
	default:
		res.state = Undefined
	}

	return &res, nil
}

func peersFromStackItems(stack []stackitem.Item, method string) (*GetNetMapValues, error) {
	if ln := len(stack); ln != 1 {
		return nil, fmt.Errorf("unexpected stack item count (%s): %d",
			method, ln)
	}

	peers, err := client.ArrayFromStackItem(stack[0])
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array from stack item (%s): %w",
			method, err)
	}

	res := &GetNetMapValues{
		peers: make([][]byte, 0, len(peers)),
	}

	for i := range peers {
		peer, err := peerInfoFromStackItem(peers[i])
		if err != nil {
			return nil, fmt.Errorf("could not parse stack item (Peer #%d): %w", i, err)
		}

		res.peers = append(res.peers, peer)
	}

	return res, nil
}

func peerInfoFromStackItem(prm stackitem.Item) ([]byte, error) {
	prms, err := client.ArrayFromStackItem(prm)
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array (PeerInfo): %w", err)
	} else if ln := len(prms); ln != nodeInfoFixedPrmNumber {
		return nil, fmt.Errorf(
			"unexpected stack item count (PeerInfo): expected %d, has %d",
			nodeInfoFixedPrmNumber,
			ln,
		)
	}

	return client.BytesFromStackItem(prms[0])
}
