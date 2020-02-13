package chordpb

import (
	"bytes"
	"encoding/hex"
	"math/big"
)

// Message types
const (
	NodeJoin    = iota // A node is joining the network
	NodeLeave          // A node is leaving the network
	HeartBeat          // Heartbeat signal
	NodeNotify         // Notified of node existence
	NodeAnn            // A node has been announced
	KeySet             // Store key-value
	KeyGet             // Get value by key
	KeyDel             // Delete key
	SuccReq            // A request for a nodes successor
	PredReq            // A request for a nodes predecessor
	StatusError        // Response indicating an error
	StatusOk           // Simple status OK response
)

type KVStore interface {
	Set(key, value []byte) error
	Get(key []byte) (value []byte, err error)
	Del(key []byte) error
}

// Represents a Hash algorithm
type Hasher interface {
	Hash(data []byte) [128]byte
	Size() int
}

type Node struct {
	Id   NodeID
	Host string
	Port int
}

type Message struct {
	ID      NodeID
	Key     []byte
	Purpose int
	Sender  Node
	Target  Node
	Hops    int
	Body    []byte
	Errors  []string
}

type NodeID [128]byte

func (n NodeID) String() string {
	return hex.EncodeToString(n[:])
}

// Add a integer to the NodeID.
func (n NodeID) Add(i *big.Int) NodeID {
	newVal := big.NewInt(0)
	y := big.NewInt(0)
	y.SetBytes(n[:])
	nn := NodeID{}
	copy(nn[:], newVal.Add(y, i).Bytes())
	return nn
}

// Returns true iff NodeID n < id
func (n NodeID) Less(id NodeID) bool {
	return bytes.Compare(n[:], id[:]) == -1
}

// Returns true iff NodeID n > id
func (n NodeID) Greater(id NodeID) bool {
	return bytes.Compare(n[:], id[:]) == 1
}

// Returns true iff NodeID n == id
func (n NodeID) Equal(id NodeID) bool {
	return bytes.Compare(n[:], id[:]) == 0
}

// Returns true iff NodeID n <= id
func (n NodeID) LE(id NodeID) bool {
	return n.Less(id) || n.Equal(id)
}

// Returns true iff NodeID n >= id
func (n NodeID) GE(id NodeID) bool {
	return n.Greater(id) || n.Equal(id)
}
