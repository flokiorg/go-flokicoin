package mempool

// RemovalReason indicates why a transaction left the mempool. It is used by
// the fee estimator to decide whether to treat a removal as a confirmation or
// a failure.
type RemovalReason int

const (
	RemovalReasonUnknown RemovalReason = iota
	// RemovalReasonBlock indicates the transaction was mined in a block.
	RemovalReasonBlock
	// RemovalReasonConflict indicates a removal due to a conflicting
	// transaction (including RBF and block-confirmed conflicts).
	RemovalReasonConflict
	// RemovalReasonReorg indicates a removal due to a chain reorg.
	RemovalReasonReorg
	// RemovalReasonEvicted indicates eviction due to mempool limits/expiry.
	RemovalReasonEvicted
	// RemovalReasonRejected indicates explicit validation failure or RPC
	// rejection after temporary admission.
	RemovalReasonRejected
)
