package evm

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/codec"
	store "github.com/cosmos/cosmos-sdk/store/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	modulev1 "github.com/evmos/ethermint/api/ethermint/evm/module/v1"
	"github.com/evmos/ethermint/x/evm/keeper"
	"github.com/evmos/ethermint/x/evm/types"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

// App Wiring Setup
func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule,ProvideKeyTable),
		appmodule.Invoke(InvokeHooks),
	)
}

var _ appmodule.AppModule = AppModule{}

// ProvideKeyTable returns the KeyTable for the evm module.
//
// It calls the ParamKeyTable function from the types package to retrieve the KeyTable.
// The KeyTable is used to register parameter sets for the evm module.
//
// Returns:
// - types.KeyTable: The KeyTable for the evm module.
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

	StoreKey        *store.KVStoreKey
	TransientKey    *store.TransientStoreKey
	Config          *modulev1.Module
	Cdc             codec.BinaryCodec
	AccountKeeper   types.AccountKeeper
	BankKeeper      types.BankKeeper
	StakingKeeper   types.StakingKeeper
	FeeMarketKeeper types.FeeMarketKeeper

	EvmConstructor evm.Constructor
	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace    paramstypes.Subspace     `optional:"true"`
	CustomPrecompiles evm.PrecompiledContracts `optional:"true"`
}

// HookInputs define the evm module hooks inputs.
type HookInputs struct {
	depinject.In

	Hooks []types.EvmHooks
	Keeper *keeper.Keeper
}

// Outputs define the module outputs for the depinject.
type Outputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

// ProvideModule creates and returns the evm module with the specified inputs.
//
// It takes Inputs as the parameter, which includes the configuration, codec, key, account keeper, and bank keeper.
// It returns Outputs containing the evm keeper and the app module.
func ProvideModule(in Inputs) Outputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	keeper := keeper.NewKeeper(
		in.Cdc,
		in.StoreKey,
		in.TransientKey,
		authority,
		in.AccountKeeper,
		in.BankKeeper,
		in.StakingKeeper,
		in.FeeMarketKeeper,
		in.CustomPrecompiles,
		in.EvmConstructor,
		in.Config.Tracer,
		in.LegacySubspace,
	)
	m := NewAppModule(keeper, in.AccountKeeper, in.LegacySubspace)
	return Outputs{Keeper: keeper, Module: m}
}

// InvokeHooks sets the EVM hooks for the provided HookInputs.
//
// Parameters:
// - in: the input HookInputs containing the hooks to set.
func InvokeHooks(in HookInputs) {
	mutiHook := keeper.MultiEvmHooks(in.Hooks)
	in.Keeper.SetHooks(mutiHook)
}
