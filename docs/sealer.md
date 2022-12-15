# Sealer

Swiss Army Knife for block building.

## Motivation

Most of the block builders around are integrated into the execution clients.
This makes rolling out changes and testing new code is painful, since it requires restarts of the whole Geth and loosing the content of the mempool.

Sealer service allows to create a builder external to Geth.

## RPC

Sealer provides only one method:

### `sealer_sealBlock`

Builds a new block out of given transactions, optionally filling up the block with transactions from the mempool.

#### Parameters

1. `Object` - Block Parameters for the block

- `parent: DATA, 32 Bytes` - hash of the parent block. This block must be present on the node executing the RPC.
- `coinbase: DATA, 20 Bytes` - reward address for the gas tips of the block.
- `timestamp: QUANTITY` - timestamp of the new block.
- `gasLimit: QUANTITY` - gas limit of the new block.
- `random: DATA, 32 Bytes` RANDAO (MixDigest) of the new block.
- `extraData: DATA` - value of the extraData of the block. If not provided, will be set to `Manifold`.

2. `Object|Array` - list of transactions (JSON encoded) to be included into the block (if possible)

3. `Boolean` - if `true` will fill the block with transactions from mempool after including provided transactions at the top of the block.

4. `Boolean` - if `true` will include call traces of all transactions included in the block.

#### Returns: `Object` - Sealed block, receipts, reasons for not inclusion of transactions and optional traces.

- `executableData: Object` - `engine.ExecutionPayloadEnvelope` of the sealed block.
  - `executionPayload: Object` - `engine.ExecutableData`
    - `parentHash: DATA, 32 Bytes`
    - `feeRecipient: DATA, 20 Bytes`
    - `stateRoot: DATA, 32 Bytes`
    - `receiptsRoot: DATA, 32 Bytes`
    - `logsBloom: DATA, 256 Bytes`
    - `prevRandao: DATA, 32 Bytes`
    - `blockNumber: QUANTITY`
    - `gasLimit: QUANTITY`
    - `gasUsed: QUANTITY`
    - `timestamp: QUANTITY`
    - `extraData: DATA`
    - `baseFeePerGas: QUANTITY`
    - `blockHash: DATA, 32 Bytes`
    - `transactions: DATA|ARRAY`
    - `withdrawals: Object|Array` - list of withdrawals
      - `index: QUANTITY` - monotonically increasing identifier issued by consensus layer
      - `validatorIndex: QUANTITY`- index of validator associated with withdrawal
	    - `address: ADDRESS` - target address for withdrawn ether
	    - `amount: QUANTITY` - value of withdrawal in Gwei
  - `blockValue: QUANTITY` - value of the block in Gwei

- `excludedTxns: Object|Array` -
  - `hash: DATA, 32 Bytes` - hash of the transaction that was not included into the block.
  - `reason: DATA` - reason (error) why the transaction could not be included.

- `receipts: Object|Array` - list of receipts of the transactions included into the block.
- `traces: Object|Array` - list of traces of the transactions included into the block.

#### Example

```json
// Request
curl -X POST -H 'content-type: application/json' --data '{"jsonrpc":"2.0","method":"sealer_sealBlock","params":[{"gasLimit":30000000},[],false,false],"id":67}' http://localhost:8545
// Result
{
  "jsonrpc": "2.0",
  "id": 67,
  "result": {
    "executableData": {
      "parentHash": "0x5eef55399e7096dea25ec8bd2c01bef23ee3ebe60aa80000974250eeb2657458",
      "feeRecipient": "0x0000000000000000000000000000000000000000",
      "stateRoot": "0xfa24ca12c7942609e5b066ef24790ed183dd47d61d6d2a193d2dfa06da024573",
      "receiptsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
      "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      "prevRandao": "0x0000000000000000000000000000000000000000000000000000000000000000",
      "blockNumber": "0x7c35be",
      "gasLimit": "0x1c9c380",
      "gasUsed": "0x0",
      "timestamp": "0x0",
      "extraData": "0x4d616e69666f6c64",
      "baseFeePerGas": "0x5046c8351",
      "blockHash": "0x33918754c1870c72a6bd575f851abc771cf38317d3740104ff0554adb88aaeb9",
      "transactions": []
    },
    "excludedTxns": [],
    "receipts": [],
    "profit": "0x0"
  }
}

```
