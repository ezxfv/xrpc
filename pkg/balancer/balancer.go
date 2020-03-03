package balancer

type Type int

const (
	Round Type = iota
	WeightedRound
	LeastConn
	WeightedLeastConn
	Random
	IPHash
)
