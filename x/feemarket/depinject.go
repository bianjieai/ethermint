package feemarket

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	store "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	modulev1 "github.com/evmos/ethermint/api/ethermint/feemarket/module/v1"
	"github.com/evmos/ethermint/x/feemarket/keeper"
	"github.com/evmos/ethermint/x/feemarket/types"
)

// App Wiring Setup
func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule, ProvideKeyTable),
	)
}

var _ appmodule.AppModule = AppModule{}

// ProvideKeyTable returns the KeyTable for the feemarket module.
//
// It calls the ParamKeyTable function from the types package to retrieve the KeyTable.
// The KeyTable is used to register parameter sets for the feemarket module.
//
// Returns:
// - types.KeyTable: The KeyTable for the feemarket module.
func ProvideKeyTable() paramstypes.KeyTable {
	return types.ParamKeyTable()
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Inputs define the module inputs for the depinject.
type Inputs struct {
	depinject.In

	StoreKey     *store.KVStoreKey
	Cdc          codec.Codec
	TransientKey *store.TransientStoreKey
	Config       *modulev1.Module

	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace paramstypes.Subspace `optional:"true"`
}

// Outputs define the module outputs for the depinject.
type Outputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

// ProvideModule creates and returns the feemarket module with the specified inputs.
//
// It takes Inputs as the parameter, which includes the configuration, codec, key, account keeper, and bank keeper.
// It returns Outputs containing the feemarket keeper and the app module.
func ProvideModule(in Inputs) Outputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	keeper := keeper.NewKeeper(
		in.Cdc,
		authority,
		in.StoreKey,
		in.TransientKey,
		in.LegacySubspace,
	)
	m := NewAppModule(keeper, in.LegacySubspace)
	return Outputs{Keeper: keeper, Module: m}
}
