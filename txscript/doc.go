// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2024 The Flokicoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package txscript implements the flokicoin transaction script language.

A complete description of the script language used by flokicoin can be found at
https://en.bitcoin.it/wiki/Script.  The following only serves as a quick
overview to provide information on how to use the package.

This package provides data structures and functions to parse and execute
flokicoin transaction scripts.

# Script Overview

Flokicoin transaction scripts are written in a stack-base, FORTH-like language.

The flokicoin script language consists of a number of opcodes which fall into
several categories such as pushing and popping data to and from the stack,
performing basic and bitwise arithmetic, conditional branching, comparing
hashes, and checking cryptographic signatures.  Scripts are processed from left
to right and intentionally do not provide loops.

The vast majority of Flokicoin scripts at the time of this writing are of several
standard forms which consist of a spender providing a public key and a signature
which proves the spender owns the associated private key.  This information
is used to prove the spender is authorized to perform the transaction.

One benefit of using a scripting language is added flexibility in specifying
what conditions must be met in order to spend flokicoins.

# Errors

Errors returned by this package are of type txscript.Error.  This allows the
caller to programmatically determine the specific error by examining the
ErrorCode field of the type asserted txscript.Error while still providing rich
error messages with contextual information.  A convenience function named
IsErrorCode is also provided to allow callers to easily check for a specific
error code.  See ErrorCode in the package documentation for a full list.
*/
package txscript
