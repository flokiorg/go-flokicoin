// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package connmgr implements a generic Flokicoin network connection manager.

# Connection Manager Overview

Connection Manager handles all the general connection concerns such as
maintaining a set number of outbound connections, sourcing peers, banning,
limiting max connections, tor lookup, etc.
*/
package connmgr
