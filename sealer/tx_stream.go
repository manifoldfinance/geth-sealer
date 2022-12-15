package sealer

import "github.com/ethereum/go-ethereum/core/types"

type txStream struct {
	topTransactions []*types.Transaction
	mempool         *types.TransactionsByPriceAndNonce
}

func newTxStream(topTransactions []*types.Transaction, mempool *types.TransactionsByPriceAndNonce) *txStream {
	return &txStream{
		topTransactions: topTransactions,
		mempool:         mempool,
	}

}

func (ts *txStream) peek() *types.Transaction {
	if len(ts.topTransactions) > 0 {
		return ts.topTransactions[0]
	}
	return ts.mempool.Peek()
}

func (ts *txStream) pop() {
	if len(ts.topTransactions) > 0 {
		ts.topTransactions = ts.topTransactions[1:]
		return
	}
	ts.mempool.Pop()
}

func (ts *txStream) shift() {
	if len(ts.topTransactions) > 0 {
		ts.topTransactions = ts.topTransactions[1:]
		return
	}
	ts.mempool.Shift()
}
