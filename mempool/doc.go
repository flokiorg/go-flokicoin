// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package mempool provides a policy-enforced pool of unmined flokicoin transactions.

A key responsibility of the flokicoin network is mining user-generated transactions
into blocks.  In order to facilitate this, the mining process relies on having a
readily-available source of transactions to include in a block that is being
solved.

At a high level, this package satisfies that requirement by providing an
in-memory pool of fully validated transactions that can also optionally be
further filtered based upon a configurable policy.

One of the policy configuration options controls whether or not "standard"
transactions are accepted.  In essence, a "standard" transaction is one that
satisfies a fairly strict set of requirements that are largely intended to help
provide fair use of the system to all users.  It is important to note that what
is considered a "standard" transaction changes over time.  For some insight, at
the time of this writing, an example of SOME of the criteria that are required
for a transaction to be considered standard are that it is of the most-recently
supported version, finalized, does not exceed a specific size, and only consists
of specific script forms.

Since this package does not deal with other flokicoin specifics such as network
communication and transaction relay, it returns a list of transactions that were
accepted which gives the caller a high level of flexibility in how they want to
proceed.  Typically, this will involve things such as relaying the transactions
to other peers on the network and notifying the mining process that new
transactions are available.

# Feature Overview

The following is a quick overview of the major features.  It is not intended to
be an exhaustive list.

  - Maintain a pool of fully validated transactions
    1. Reject non-fully-spent duplicate transactions
    2. Reject coinbase transactions
    3. Reject double spends (both from the chain and other transactions in pool)
    4. Reject invalid transactions according to the network consensus rules
    5. Full script execution and validation with signature cache support
    6. Individual transaction query support
  - Orphan transaction support (transactions that spend from unknown outputs)
    1. Configurable limits (see transaction acceptance policy)
    2. Automatic addition of orphan transactions that are no longer orphans as new
    transactions are added to the pool
    3. Individual orphan transaction query support
  - Configurable transaction acceptance policy
    1. Option to accept or reject standard transactions
    2. Option to accept or reject transactions based on priority calculations
    3. Rate limiting of low-fee and free transactions
    4. Non-zero fee threshold
    5. Max signature operations per transaction
    6. Max orphan transaction size
    7. Max number of orphan transactions allowed
  - Additional metadata tracking for each transaction
    1. Timestamp when the transaction was added to the pool
    2. Most recent block height when the transaction was added to the pool
    3. The fee the transaction pays
    4. The starting priority for the transaction
  - Manual control of transaction removal
    1. Recursive removal of all dependent transactions

# Errors

Errors returned by this package are either the raw errors provided by underlying
calls or of type mempool.RuleError.  Since there are two classes of rules
(mempool acceptance rules and blockchain (consensus) acceptance rules), the
mempool.RuleError type contains a single Err field which will, in turn, either
be a mempool.TxRuleError or a blockchain.RuleError.  The first indicates a
violation of mempool acceptance rules while the latter indicates a violation of
consensus acceptance rules.  This allows the caller to easily differentiate
between unexpected errors, such as database errors, versus errors due to rule
violations through type assertions.  In addition, callers can programmatically
determine the specific rule violation by type asserting the Err field to one of
the aforementioned types and examining their underlying ErrorCode field.
*/
package mempool
