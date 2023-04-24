package sealer

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type NonceAndBalance struct {
	Address common.Address `json:"address"`
	Nonce   hexutil.Uint64 `json:"nonce"`
	Balance hexutil.Big    `json:"balance"`
}

func (s *Sealer) GetBulkAccountNoncesAndBalances(accounts []common.Address, blockHash common.Hash) ([]NonceAndBalance, error) {
	bl := s.chain.GetBlockByHash(blockHash)
	if bl == nil {
		return nil, fmt.Errorf("could not find block %s", blockHash.Hex())
	}

	state, err := s.chain.StateAt(bl.Root())
	if err != nil {
		return nil, fmt.Errorf("could not find root state of block %s: %s", blockHash.Hex(), err)
	}

	state.StartPrefetcher("sealer")
	defer state.StopPrefetcher()

	nnb := make([]NonceAndBalance, len(accounts))
	for i, a := range accounts {
		a := a
		nonce := state.GetNonce(a)
		balance := state.GetBalance(a)
		balanceCopy := big.NewInt(0).Set(balance)

		nnb[i] = NonceAndBalance{
			Address: a,
			Nonce:   hexutil.Uint64(nonce),
			Balance: hexutil.Big(*balanceCopy),
		}
	}

	return nnb, nil

}
