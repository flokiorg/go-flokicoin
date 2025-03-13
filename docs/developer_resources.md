# Developer Resources

* [Code Contribution Guidelines](https://github.com/flokiorg/go-flokicoin/tree/main/docs/code_contribution_guidelines.md)

* [JSON-RPC Reference](https://github.com/flokiorg/go-flokicoin/tree/main/docs/json_rpc_api.md)
  * [RPC Examples](https://github.com/flokiorg/go-flokicoin/tree/main/docs/json_rpc_api.md#ExampleCode)

* The go-flokicoin related Go Packages:
  * [rpcclient](https://github.com/flokiorg/go-flokicoin/tree/main/rpcclient) - Implements a
    robust and easy to use Websocket-enabled Bitcoin JSON-RPC client
  * [chainjson](https://github.com/flokiorg/go-flokicoin/tree/main/chainjson) - Provides an extensive API
    for the underlying JSON-RPC command and return values
  * [wire](https://github.com/flokiorg/go-flokicoin/tree/main/wire) - Implements the
    Bitcoin wire protocol
  * [peer](https://github.com/flokiorg/go-flokicoin/tree/main/peer) -
    Provides a common base for creating and managing Bitcoin network peers.
  * [blockchain](https://github.com/flokiorg/go-flokicoin/tree/main/blockchain) -
    Implements Bitcoin block handling and chain selection rules
  * [blockchain/fullblocktests](https://github.com/flokiorg/go-flokicoin/tree/main/blockchain/fullblocktests) -
    Provides a set of block tests for testing the consensus validation rules
  * [txscript](https://github.com/flokiorg/go-flokicoin/tree/main/txscript) -
    Implements the Bitcoin transaction scripting language
  * [crypto](https://github.com/flokiorg/go-flokicoin/tree/main/crypto) - Implements
    support for the elliptic curve cryptographic functions needed for the
    Bitcoin scripts
  * [database](https://github.com/flokiorg/go-flokicoin/tree/main/database) -
    Provides a database interface for the Bitcoin block chain
  * [mempool](https://github.com/flokiorg/go-flokicoin/tree/main/mempool) -
    Package mempool provides a policy-enforced pool of unmined bitcoin
    transactions.
  * [chainutil](https://github.com/flokiorg/go-flokicoin/tree/main/chainutil) - Provides Bitcoin-specific
    convenience functions and types
  * [chainhash](https://github.com/flokiorg/go-flokicoin/tree/main/chaincfg/chainhash) -
    Provides a generic hash type and associated functions that allows the
    specific hash algorithm to be abstracted.
  * [connmgr](https://github.com/flokiorg/go-flokicoin/tree/main/connmgr) -
    Package connmgr implements a generic Bitcoin network connection manager.
