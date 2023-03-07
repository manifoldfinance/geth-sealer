package logger

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

// storageCheckList is an accumulator for the set of accounts and storage slots with with existing value an EVM
// contract execution touches.
type storageCheckList map[common.Address]storageCheckListSlots

// storageCheckListSlots is an accumulator for the set of storage slots within a single
// contract that an EVM contract execution touches.
type storageCheckListSlots map[common.Hash]common.Hash

// newStorageCheckList creates a new storageCheckList.
func newStorageCheckList() storageCheckList {
	return make(storageCheckList)
}

// addAddress adds an address to the storageCheckList.
func (scl storageCheckList) addAddress(address common.Address) {
	// Set address if not previously present
	if _, present := scl[address]; !present {
		scl[address] = make(storageCheckListSlots)
	}
}

// addSlot adds a storage slot to the accesslist.
func (scl storageCheckList) addStorageCheckSlot(address common.Address, slotKey common.Hash, slotValue common.Hash) {
	// Set address if not previously present
	scl.addAddress(address)

	// Set the slot on the surely existent storage set
	scl[address][slotKey] = slotValue
}

// equal checks if the content of the current storage check list is the same as the
// content of the other one.
func (scl storageCheckList) equal(other storageCheckList) bool {
	// Cross reference the accounts first
	if len(scl) != len(other) {
		return false
	}
	// Given that len(al) == len(other), we only need to check that
	// all the items from al are in other.
	for addr := range scl {
		if _, ok := other[addr]; !ok {
			return false
		}
	}

	// Accounts match, cross reference the storage check slots too
	for addr, slots := range scl {
		otherslots := other[addr]

		if len(slots) != len(otherslots) {
			return false
		}
		// Given that len(slots) == len(otherslots), we need to check that
		// all the items from slots are in otherslots and also that are equal.
		for hash := range slots {
			if otherHash, ok := otherslots[hash]; !ok || hash != otherHash {
				return false
			}
		}
	}
	return true
}

// storageCheckList converts the storageCheckList to a types.StorageCheckList.
func (scl storageCheckList) storageCheckList() types.StorageCheckList {
	sc := make(types.StorageCheckList, 0, len(scl))
	for addr, slots := range scl {
		tuple := types.StorageCheckTuple{Address: addr, StorageKeyValueChecks: []types.StorageKeyValueCheck{}}
		for storageIndex, storageValue := range slots {
			skv := types.StorageKeyValueCheck{Index: storageIndex, Value: storageValue}
			tuple.StorageKeyValueChecks = append(tuple.StorageKeyValueChecks, skv)
		}
	}
	return sc
}

// StorageListTracer is a tracer that accumulates touched accounts, storage key slots and storage values
// into an internal set.
type StorageCheckListTracer struct {
	env  *vm.EVM
	excl map[common.Address]struct{} // Set of accounts to exclude from the list
	list storageCheckList            // Set of accounts, storage keys and storage values accessed
}

func NewStorageCheckListTracer(scl types.StorageCheckList, from, to common.Address, precompiles []common.Address) *StorageCheckListTracer {
	excl := map[common.Address]struct{}{
		from: {}, to: {},
	}
	for _, addr := range precompiles {
		excl[addr] = struct{}{}
	}
	list := newStorageCheckList()
	for _, al := range scl {
		if _, ok := excl[al.Address]; !ok {
			list.addAddress(al.Address)
		}
		for _, slot := range al.StorageKeyValueChecks {
			list.addStorageCheckSlot(al.Address, slot.Index, slot.Value)
		}
	}
	return &StorageCheckListTracer{
		excl: excl,
		list: list,
	}
}

// CaptureStart implements the EVMLogger interface to initialize the tracing operation.
func (t *StorageCheckListTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	t.env = env
}

// CaptureEnd is called after the call finishes to finalize the tracing.
func (t *StorageCheckListTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {
}

// CaptureState captures all opcodes that touch storage or addresses and adds them to the accesslist.
func (t *StorageCheckListTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	stack := scope.Stack
	stackData := stack.Data()
	stackLen := len(stackData)

	if stackLen >= 1 && op == vm.SLOAD {
		address := scope.Contract.Address()
		slot := common.Hash(stackData[stackLen-1].Bytes32())
		value := t.env.StateDB.GetState(address, slot)
		t.list.addStorageCheckSlot(address, slot, value)
	}
}

// CaptureFault implements the EVMLogger interface to trace an execution fault.
func (t *StorageCheckListTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
}

// CaptureEnter is called when EVM enters a new scope (via call, create or selfdestruct).
func (t *StorageCheckListTracer) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

// CaptureExit is called when EVM exits a scope, even if the scope didn't
// execute any code.
func (t *StorageCheckListTracer) CaptureExit(output []byte, gasUsed uint64, err error) {}

func (t *StorageCheckListTracer) CaptureTxStart(gasLimit uint64) {}

func (t *StorageCheckListTracer) CaptureTxEnd(restGas uint64) {}

// StorageCheckList returns the current storageCheckList maintained by the tracer.
func (t *StorageCheckListTracer) StorageCheckList() types.StorageCheckList {
	return t.list.storageCheckList()
}

// Equal returns if the content of two storage check list traces are equal.
func (t *StorageCheckListTracer) Equal(other *StorageCheckListTracer) bool {
	return t.list.equal(other.list)
}
