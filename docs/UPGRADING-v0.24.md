# v0.24 dependency upgrade

This release starts from `v0.23.0-irishub` and updates Cosmos SDK to `v0.53.8`,
IBC-Go to `v10.7.0`, and CometBFT to `v0.38.23`. Go `1.23.8` or newer is required.
The application follows the [SDK upgrade reference](https://github.com/cosmos/cosmos-sdk/blob/v0.53.8/UPGRADING.md)
and the [IBC v10 migration guide](https://docs.cosmos.network/ibc/latest/migrations/v8_1-to-v10).

The auth pre-blocker follows upgrade. Unordered transactions remain disabled;
epochs, protocolpool, and IBC v2 application routing are not enabled. IBC v1 transfer
remains registered, and Tendermint and Solo Machine light clients are explicitly
registered using the v10 light-client API. The obsolete capability module and its
keeper/store are removed.

## EVM dependency

IBC-Go declares Geth `v1.15.11`, but this application pins the effective Geth module
to `v1.10.26` through `replace` to retain the existing EVM implementation. IBC-Go
uses Geth's `accounts/abi` for ICS-20 ABI packet encoding. Its transfer types tests
must pass with this replacement. This is an application-tested dependency
combination, not an upstream IBC certification of Geth `v1.10.26`.

## Existing chain deployment

The SDK upgrade requires coordinated chain deployment. This branch does not define
a production upgrade plan or height: `app/upgrades.go` still registers only
`integration-test-upgrade`. Before deploying over existing state, the chain owner
must define and test the production upgrade, module migrations, and removal of the
`capability` store with `StoreUpgrades.Deleted`, as required by IBC. Fresh genesis
initialization and unit tests do not demonstrate compatibility with a production
state snapshot. A genesis-based hard fork also needs its chain-specific exported
state conversion and validation.

## Validation commands

```sh
go list -m github.com/cosmos/cosmos-sdk github.com/cosmos/ibc-go/v10 github.com/cometbft/cometbft github.com/ethereum/go-ethereum
go test ./app/... ./x/evm/... ./x/feemarket/...
go test github.com/cosmos/ibc-go/v10/modules/apps/transfer/types
go test ./... -run '^$'
go build ./cmd/ethermintd
```

The existing FeeMarket integration suite has five failing low-fee transaction
cases, also reproduced on `v0.23.0-irishub` (33 of 38 Ginkgo cases pass). Its
fixture sets a base fee and then commits empty blocks, which changes that fee
before the transaction runs, including applying the `MinGasPrice` floor. The
upgrade preserves the fee calculation and does not rewrite those integration
expectations. Treat these as a known baseline validation gap.
