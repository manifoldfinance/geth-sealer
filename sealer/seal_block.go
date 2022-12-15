package sealer

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/params"
)

type ExcludedTransaction struct {
	Hash   common.Hash `json:"hash"`
	Reason string      `json:"reason"`
}

type SealedBlock struct {
	ExecutableData       *beacon.ExecutableDataV1 `json:"executableData"`
	ExcludedTransactions []ExcludedTransaction    `json:"excludedTxns"`
	Receipts             []*types.Receipt         `json:"receipts"`
	Traces               []json.RawMessage        `json:"traces,omitempty"`
	Profit               *hexutil.Big             `json:"profit"`
}

type BlockParameters struct {
	ParentHash common.Hash    `json:"parent"`
	Coinbase   common.Address `json:"coinbase"`
	Timestamp  uint64         `json:"timestamp"`
	GasLimit   uint64         `json:"gasLimit"`
	Random     common.Hash    `json:"random"`
	Extra      []byte         `json:"extraData"`
}

type Sealer struct {
	signer      types.Signer
	chainConfig *params.ChainConfig
	chain       *core.BlockChain
	engine      consensus.Engine
	txpool      *core.TxPool
}

func newSealer(
	backend *eth.Ethereum,
) *Sealer {
	chain := backend.BlockChain()
	chainConfig := chain.Config()
	signer := types.LatestSignerForChainID(chainConfig.ChainID)
	return &Sealer{
		signer:      signer,
		chainConfig: chainConfig,
		chain:       chain,
		engine:      chain.Engine(),
		txpool:      backend.TxPool(),
	}
}

func (s *Sealer) SealBlock(p *BlockParameters, txns []*types.Transaction, fillWithMempool bool, trace bool) (*SealedBlock, error) {
	var gp core.GasPool

	parent := s.chain.CurrentBlock()
	if p.ParentHash != (common.Hash{}) {
		parent = s.chain.GetBlockByHash(p.ParentHash)
		if parent == nil {
			return nil, fmt.Errorf("could not find parent block %s", p.ParentHash.Hex())
		}
	}

	num := parent.Number()

	extraData := p.Extra
	if extraData == nil {
		extraData = []byte("Manifold")
	}

	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     num.Add(num, common.Big1),
		GasLimit:   p.GasLimit,
		Time:       p.Timestamp,
		Coinbase:   p.Coinbase,
		Extra:      extraData,
		MixDigest:  p.Random,
	}

	// Set baseFee and GasLimit if we are on an EIP-1559 chain
	if s.chainConfig.IsLondon(header.Number) {
		header.BaseFee = misc.CalcBaseFee(s.chainConfig, parent.Header())
	}

	// Run the consensus preparation with the default or customized consensus engine.
	err := s.engine.Prepare(s.chain, header)
	if err != nil {
		return nil, err
	}

	receipts := []*types.Receipt{}
	includedTxns := []*types.Transaction{}

	state, err := s.chain.StateAt(parent.Root())
	if err != nil {
		return nil, fmt.Errorf("could not find root state of parent block %s: %s", p.ParentHash.Hex(), err)
	}

	state.StartPrefetcher("sealer")
	defer state.StopPrefetcher()

	blockReward := big.NewInt(0)

	sb := &SealedBlock{
		ExcludedTransactions: []ExcludedTransaction{},
	}

	gp.AddGas(header.GasLimit)

	mempool := map[common.Address]types.Transactions{}

	if fillWithMempool {
		mempool = s.txpool.Pending(true)
	}

	stream := newTxStream(txns, types.NewTransactionsByPriceAndNonce(s.signer, mempool, header.BaseFee))

	for {
		tx := stream.peek()
		if tx == nil {
			break
		}

		if gp.Gas() < params.TxGas {
			break
		}

		// Error may be ignored here. The error has already been checked
		// during transaction acceptance is the transaction pool.
		//
		// We use the eip155 signer regardless of the current hf.
		_, err := types.Sender(s.signer, tx)
		if err != nil {
			sb.ExcludedTransactions = append(sb.ExcludedTransactions, ExcludedTransaction{Hash: tx.Hash(), Reason: err.Error()})
			continue
		}

		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		if tx.Protected() && !s.chainConfig.IsEIP155(header.Number) {
			err := fmt.Errorf("ignoring reply protected transaction with hash %s eip155 %s", tx.Hash().Hex(), s.chainConfig.EIP155Block.String())
			sb.ExcludedTransactions = append(sb.ExcludedTransactions, ExcludedTransaction{Hash: tx.Hash(), Reason: err.Error()})
			continue
		}
		// Start executing the transaction
		state.Prepare(tx.Hash(), len(includedTxns))

		receipt, traceBody, err := s.commitTransaction(
			state,
			&gp,
			p.Coinbase,
			header,
			tx,
			trace,
		)
		if err != nil {
			sb.ExcludedTransactions = append(sb.ExcludedTransactions, ExcludedTransaction{Hash: tx.Hash(), Reason: err.Error()})
			stream.pop()
			continue
		}

		reward := big.NewInt(int64(receipt.GasUsed))
		reward = reward.Mul(reward, tx.EffectiveGasTipValue(header.BaseFee))

		blockReward = blockReward.Add(blockReward, reward)

		receipts = append(receipts, receipt)
		includedTxns = append(includedTxns, tx)
		if trace {
			sb.Traces = append(sb.Traces, traceBody)
		}
		stream.shift()
	}

	bl, err := s.engine.FinalizeAndAssemble(s.chain, header, state, includedTxns, nil, receipts)
	if err != nil {
		return nil, fmt.Errorf("could not assemble block: %w", err)
	}

	sb.ExecutableData = beacon.BlockToExecutableData(bl)

	// patch up receipts - null logs is not 'good' for deserializing JSON
	for _, r := range receipts {
		if r.Logs == nil {
			r.Logs = []*types.Log{}
		}
	}
	sb.Receipts = receipts
	sb.Profit = (*hexutil.Big)(blockReward)

	return sb, nil
}

func (s *Sealer) commitTransaction(
	state *state.StateDB,
	gasPool *core.GasPool,
	coinbase common.Address,
	header *types.Header,
	tx *types.Transaction,
	trace bool,
) (rcpt *types.Receipt, traceBody json.RawMessage, err error) {
	snap := state.Snapshot()
	vmConfig := *s.chain.GetVMConfig()
	if trace {
		callTracer, err := tracers.New("callTracer", &tracers.Context{}, []byte(`{}`))
		if err != nil {
			return nil, nil, fmt.Errorf("could not create a new call tracer")
		}
		vmConfig.Tracer = callTracer
		vmConfig.Debug = true
		defer func() {
			if err != nil {
				return
			}
			traceBody, err = callTracer.GetResult()
		}()
	}
	receipt, err := core.ApplyTransaction(s.chainConfig, s.chain, &coinbase, gasPool, state, header, tx, &header.GasUsed, vmConfig)
	if err != nil {
		state.RevertToSnapshot(snap)
		return nil, nil, err
	}
	return receipt, nil, nil

}
