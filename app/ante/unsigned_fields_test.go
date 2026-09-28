package ante_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/evmos/ethermint/ethereum/eip712"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// Mutate the encoded body after signing: neither legacy signing format covers
// the SDK 0.53 fields, so these must be rejected independently of the signature.
func TestRejectUnsignedTxBodyFields(t *testing.T) {
	for _, legacyEIP712 := range []bool{false, true} {
		name := "Ethereum"
		if legacyEIP712 {
			name = "LegacyEIP712"
		}
		t.Run(name, func(t *testing.T) {
			suite := &AnteTestSuite{
				enableLondonHF:           true,
				useLegacyEIP712Extension: legacyEIP712,
				useLegacyEIP712TypedData: legacyEIP712,
			}
			suite.SetT(t)
			suite.SetupTest()
			suite.app.FeeMarketKeeper.SetBaseFee(suite.ctx, big.NewInt(100))
			from := sdk.AccAddress(suite.priv.PubKey().Address())
			var signedTx sdk.Tx
			if legacyEIP712 {
				fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20_000_000)))
				signedTx = suite.CreateTestEIP712TxBuilderMsgSend(from, suite.priv, suite.ctx.ChainID(), 200_000, fees).GetTx()
			} else {
				msg := evmtypes.NewTxContract(suite.app.EvmKeeper.ChainID(), 0, big.NewInt(10), 100_000, big.NewInt(150), big.NewInt(200), nil, nil, nil)
				msg.From = common.BytesToAddress(from).Hex()
				signedTx = suite.CreateTestTx(msg, suite.priv, 0, false)
			}
			signedBytes, err := suite.clientCtx.TxConfig.TxEncoder()(signedTx)
			require.NoError(t, err)
			future := suite.ctx.BlockTime().Add(time.Minute)
			zero := time.Time{}
			for _, mode := range []struct {
				name string
				exec sdk.ExecMode
			}{
				{"CheckTx", sdk.ExecModeCheck},
				{"DeliverTx", sdk.ExecModeFinalize},
				{"ReCheckTx", sdk.ExecModeReCheck},
				{"Simulate", sdk.ExecModeSimulate},
			} {
				t.Run(mode.name, func(t *testing.T) {
					for _, tc := range []struct {
						name      string
						unordered bool
						timestamp *time.Time
					}{
						{name: "unchanged"},
						{name: "unordered", unordered: true},
						{name: "unordered_with_timestamp", unordered: true, timestamp: &future},
						{name: "timestamp", timestamp: &future},
						{name: "explicit_zero_timestamp", timestamp: &zero},
					} {
						t.Run(tc.name, func(t *testing.T) {
							var raw txtypes.TxRaw
							require.NoError(t, raw.Unmarshal(signedBytes))
							var body txtypes.TxBody
							require.NoError(t, body.Unmarshal(raw.BodyBytes))
							body.Unordered, body.TimeoutTimestamp = tc.unordered, tc.timestamp
							raw.BodyBytes, err = body.Marshal()
							require.NoError(t, err)
							encoded, err := raw.Marshal()
							require.NoError(t, err)
							decoded, err := suite.clientCtx.TxConfig.TxDecoder()(encoded)
							require.NoError(t, err)
							ctx, _ := suite.ctx.CacheContext()
							ctx = ctx.WithTxBytes(encoded).
								WithIsCheckTx(mode.exec == sdk.ExecModeCheck).
								WithIsReCheckTx(mode.exec == sdk.ExecModeReCheck).
								WithExecMode(mode.exec)
							_, err = suite.anteHandler(ctx, decoded, mode.exec == sdk.ExecModeSimulate)
							if tc.name == "unchanged" {
								require.NoError(t, err)
							} else {
								require.ErrorIs(t, err, errortypes.ErrInvalidRequest)
								require.ErrorContains(t, err, "Unordered and TimeoutTimestamp")
							}
						})
					}
				})
			}
		})
	}
}

// Extensionless EIP-712 transactions use the SDK's DIRECT sign mode and reach
// EIP-712 verification through the public key's signature fallback.
func TestExtensionlessEIP712RejectsUnsignedTxBodyFields(t *testing.T) {
	for _, legacyTypedData := range []bool{false, true} {
		name := "current_typed_data"
		if legacyTypedData {
			name = "legacy_typed_data"
		}
		t.Run(name, func(t *testing.T) {
			suite := &AnteTestSuite{enableLondonHF: true, useLegacyEIP712TypedData: legacyTypedData}
			suite.SetT(t)
			suite.SetupTest()
			from := sdk.AccAddress(suite.priv.PubKey().Address())
			fees := sdk.NewCoins(sdk.NewCoin(evmtypes.DefaultEVMDenom, sdkmath.NewInt(20_000_000)))
			builder := suite.CreateTestEIP712TxBuilderMsgSend(from, suite.priv, suite.ctx.ChainID(), 200_000, fees)
			sigs, err := builder.GetTx().GetSignaturesV2()
			require.NoError(t, err)
			require.Equal(t, signing.SignMode_SIGN_MODE_DIRECT, sigs[0].Data.(*signing.SingleSignatureData).SignMode)
			signedBytes, err := suite.clientCtx.TxConfig.TxEncoder()(builder.GetTx())
			require.NoError(t, err)
			future := suite.ctx.BlockTime().Add(time.Minute)
			zero := time.Time{}
			for _, tc := range []struct {
				name      string
				unordered bool
				timestamp *time.Time
			}{
				{name: "unchanged"},
				{name: "timestamp", timestamp: &future},
				{name: "explicit_zero_timestamp", timestamp: &zero},
				{name: "unordered", unordered: true},
				{name: "unordered_with_timestamp", unordered: true, timestamp: &future},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var raw txtypes.TxRaw
					require.NoError(t, raw.Unmarshal(signedBytes))
					var body txtypes.TxBody
					require.NoError(t, body.Unmarshal(raw.BodyBytes))
					require.Empty(t, body.ExtensionOptions)
					body.Unordered, body.TimeoutTimestamp = tc.unordered, tc.timestamp
					raw.BodyBytes, err = body.Marshal()
					require.NoError(t, err)
					encoded, err := raw.Marshal()
					require.NoError(t, err)
					// Exercise actual signature verification in both admission and execution.
					for _, checkTx := range []bool{true, false} {
						decoded, err := suite.clientCtx.TxConfig.TxDecoder()(encoded)
						require.NoError(t, err)
						ctx, _ := suite.ctx.CacheContext()
						mode := sdk.ExecModeFinalize
						if checkTx {
							mode = sdk.ExecModeCheck
						}
						ctx = ctx.WithTxBytes(encoded).WithIsCheckTx(checkTx).WithExecMode(mode).WithIsSigverifyTx(true)
						_, err = suite.anteHandler(ctx, decoded, false)
						switch {
						case tc.unordered:
							require.ErrorIs(t, err, errortypes.ErrNotSupported)
						case tc.timestamp != nil:
							require.ErrorIs(t, err, errortypes.ErrUnauthorized)
						default:
							require.NoError(t, err)
						}
					}

					signDoc := txtypes.SignDoc{
						BodyBytes: raw.BodyBytes, AuthInfoBytes: raw.AuthInfoBytes,
						ChainId: suite.ctx.ChainID(), AccountNumber: suite.app.AccountKeeper.GetAccount(suite.ctx, from).GetAccountNumber(),
					}
					signBytes, err := signDoc.Marshal()
					require.NoError(t, err)
					// Both conversion entry points must reject, including the legacy
					// entry point when called independently of the public key fallback.
					for _, convert := range []func([]byte) ([]byte, error){eip712.GetEIP712BytesForMsg, eip712.LegacyGetEIP712BytesForMsg} {
						_, err := convert(signBytes)
						if tc.unordered || tc.timestamp != nil {
							require.ErrorContains(t, err, "body contains unsupported fields")
						} else {
							require.NoError(t, err)
						}
					}
					if tc.unordered {
						return
					}
					// A native DIRECT signature does cover the timestamp and remains
					// valid when the same body is signed instead of converted to EIP-712.
					signature, err := suite.priv.Sign(signBytes)
					require.NoError(t, err)
					raw.Signatures = [][]byte{signature}
					encoded, err = raw.Marshal()
					require.NoError(t, err)
					decoded, err := suite.clientCtx.TxConfig.TxDecoder()(encoded)
					require.NoError(t, err)
					ctx, _ := suite.ctx.CacheContext()
					ctx = ctx.WithTxBytes(encoded).WithIsCheckTx(true).WithIsSigverifyTx(true)
					_, err = suite.anteHandler(ctx, decoded, false)
					require.NoError(t, err)
				})
			}
		})
	}
}
