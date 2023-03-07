package state

import (
	"github.com/ethereum/go-ethereum/common"
)

type storageCheckList struct {
	addresses map[common.Address]int
	slots     []map[common.Hash]common.Hash
}

// newStorageCheckList creates a new storageCheckList
func newStorageCheckList() *storageCheckList {
	return &storageCheckList{
		addresses: make(map[common.Address]int),
		slots:     make([]map[common.Hash]common.Hash, 0),
	}
}

// ContainsAddress returns true if the address is in the storage check list.
func (sc *storageCheckList) ContainsAddress(address common.Address) bool {
	_, ok := sc.addresses[address]
	return ok
}

// Contains checks if a slot within an account is present in the access list, returning
// separate flags for the presence of the account and the slot respectively.
func (sc *storageCheckList) Contains(address common.Address, slot common.Hash, value common.Hash) (addressPresent bool, slotPresent bool, valueMatches bool) {
	idx, ok := sc.addresses[address]
	if !ok {
		// no such address (and hence zero slots)
		return false, false, false
	}
	if idx == -1 {
		// address yes, but no slots
		return true, false, false
	}
	if slotValue, slotPresent := sc.slots[idx][slot]; slotPresent {
		return true, true, slotValue == value
	}
	return true, false, false
}

// AddAddress adds an address to the access list, and returns 'true' if the operation
// caused a change (addr was not previously in the list).
func (sc *storageCheckList) AddAddress(address common.Address) bool {
	if _, present := sc.addresses[address]; present {
		return false
	}
	sc.addresses[address] = -1
	return true
}

// AddSlotAndValue adds the specified (addr, slot, value) combo to the check list.
// Return values are:
// - address added
// - slot added
// - value added
// For any 'true' value returned, a corresponding journal entry must be made.
func (sc *storageCheckList) AddSlotAndValue(addr common.Address, slot common.Hash, value common.Hash) (addrChange bool, slotChange bool, valueChange bool) {
	idx, addrPresent := sc.addresses[addr]
	if !addrPresent || idx == -1 {
		// Address not present, or addr present but no slots there
		sc.addresses[addr] = len(sc.slots)
		slotmap := map[common.Hash]common.Hash{}
		sc.slots = append(sc.slots, slotmap)
		return !addrPresent, true, false
	}
	// There is already an (address,slot) mapping
	slotmap := sc.slots[idx]
	if _, ok := slotmap[slot]; !ok {
		slotmap[slot] = value
		// Journal add slot change
		return false, true, true
	}
	// No changes required
	return false, false, false
}
