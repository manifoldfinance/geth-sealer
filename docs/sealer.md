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

#### Returns

- `executableData: Object` - `beacon.ExecutableDataV1` of the sealed block.

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

- `excludedTxns: Object|Array` -

  - `hash: DATA, 32 Bytes` - hash of the transaction that was not included into the block.
  - `reason: DATA` - reason (error) why the transaction could not be included.

- `receipts: Object|Array` - list of receipts of the transactions included into the block.
- `traces: Object|Array` - list of traces of the transactions included into the block.
- `profit: QUANTITY` - total gas tip for the built block in Wei.
