// Copyright (c) 2013-2014 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainutil

const (
	// LokiPerFlokicent is the number of loki in one flokicoin cent.
	LokiPerFlokicent = 1e6

	// LokiPerFlokicoin is the number of loki in one flokicoin (1 FLC).
	LokiPerFlokicoin = 1e8

	// MaxLoki is the maximum transaction amount allowed in loki.
	MaxLoki = 21e6 * LokiPerFlokicoin
)
