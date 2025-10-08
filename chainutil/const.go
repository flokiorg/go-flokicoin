// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainutil

import "math"

const (
	// LokiPerFlokicent is the number of loki in one flokicoin cent.
	LokiPerFlokicent = 1e6

	// LokiPerFlokicoin is the number of loki in one flokicoin (1 FLC).
	LokiPerFlokicoin = 1e8

	// MaxLoki defines the maximum amount that can be expressed without
	// overflowing a signed 64-bit integer. Flokicoin uses perpetual inflation,
	// so we reserve the entire positive int64 space for amounts.
	MaxLoki int64 = math.MaxInt64
)
