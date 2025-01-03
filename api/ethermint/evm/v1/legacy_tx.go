package evmv1

import (
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/types"
)

// GetChainID returns the chain id field from the derived signature values
func (tx *LegacyTx) GetChainID() *big.Int {
	v, _, _ := tx.GetRawSignatureValues()
	return types.DeriveChainID(v)
}

// AsEthereumData returns an LegacyTx transaction tx from the proto-formatted
// TxData defined on the Cosmos EVM.
func (tx *LegacyTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.LegacyTx{
		Nonce:    tx.GetNonce(),
		GasPrice: stringToBigInt(tx.GetGasPrice()),
		Gas:      tx.GetGas(),
		To:       stringToAddress(tx.GetTo()),
		Value:    stringToBigInt(tx.GetValue()),
		Data:     tx.GetData(),
		V:        v,
		R:        r,
		S:        s,
	}
}

// GetRawSignatureValues returns the V, R, S signature values of the transaction.
// The return values should not be modified by the caller.
func (tx *LegacyTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return types.RawSignatureValues(tx.V, tx.R, tx.S)
}

// GetAccessList returns nil
func (tx *LegacyTx) GetAccessList() ethtypes.AccessList {
	return nil
}