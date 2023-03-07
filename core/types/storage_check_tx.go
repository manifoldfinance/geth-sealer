package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate go run github.com/fjl/gencodec -type StorageCheckTuple -out gen_storage_check_tuple.go

// StorageCheck is an EIP-XXXX storage
type StorageCheckList []StorageCheckTuple

type StorageCheckTuple struct {
	Address               common.Address         `json:"address"               gencodec:"required"`
	StorageKeyValueChecks []StorageKeyValueCheck `json:"storageKeyValueChecks" gencodec:"required"`
}

type StorageKeyValueCheck struct {
	Index     common.Hash `gencodec:"required"`
	Value     common.Hash `gencodec:"required"`
	Operation *byte
}

// StorageKeyValueChecks returns the total number of storage checks entries in the storage checks list.
func (sc StorageCheckList) StorageKeyValueChecks() int {
	sum := 0
	for _, tuple := range sc {
		sum += len(tuple.StorageKeyValueChecks)
	}
	return sum
}

// StorageCheckListTX is the data of EIP-XXXX storage check list transactions.
type StorageCheckListTx struct {
	ChainID          *big.Int         // destination chain ID
	Nonce            uint64           // nonce of sender account
	GasPrice         *big.Int         // wei per gas
	Gas              uint64           // gas limit
	To               *common.Address  `rlp:"nil"` // nil means contract creation
	Value            *big.Int         // wei amount
	Data             []byte           // contract invocation input data
	StorageCheckList StorageCheckList // EIP-XXXX storage check list
	V, R, S          *big.Int         // signature values
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *StorageCheckListTx) copy() TxData {
	cpy := &StorageCheckListTx{
		Nonce: tx.Nonce,
		To:    copyAddressPtr(tx.To),
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		StorageCheckList: make(StorageCheckList, len(tx.StorageCheckList)),
		Value:            new(big.Int),
		ChainID:          new(big.Int),
		GasPrice:         new(big.Int),
		V:                new(big.Int),
		R:                new(big.Int),
		S:                new(big.Int),
	}
	copy(cpy.StorageCheckList, tx.StorageCheckList)
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasPrice != nil {
		cpy.GasPrice.Set(tx.GasPrice)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// accessors for innerTx.
func (tx *StorageCheckListTx) txType() byte                       { return StorageCheckListTxType }
func (tx *StorageCheckListTx) chainID() *big.Int                  { return tx.ChainID }
func (tx *StorageCheckListTx) accessList() AccessList             { return nil }
func (tx *StorageCheckListTx) storageCheckList() StorageCheckList { return tx.StorageCheckList }
func (tx *StorageCheckListTx) data() []byte                       { return tx.Data }
func (tx *StorageCheckListTx) gas() uint64                        { return tx.Gas }
func (tx *StorageCheckListTx) gasPrice() *big.Int                 { return tx.GasPrice }
func (tx *StorageCheckListTx) gasTipCap() *big.Int                { return tx.GasPrice }
func (tx *StorageCheckListTx) gasFeeCap() *big.Int                { return tx.GasPrice }
func (tx *StorageCheckListTx) value() *big.Int                    { return tx.Value }
func (tx *StorageCheckListTx) nonce() uint64                      { return tx.Nonce }
func (tx *StorageCheckListTx) to() *common.Address                { return tx.To }

func (tx *StorageCheckListTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int {
	return dst.Set(tx.GasPrice)
}

func (tx *StorageCheckListTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

func (tx *StorageCheckListTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}
