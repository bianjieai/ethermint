package app

import (
	"testing"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
)

func TestEthermintAppExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := SetupWithDB(false, nil, db)
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
	})
	require.NoError(t, err, "FinalizeBlock should not have an error")

	_, err = app.Commit()
	require.NoError(t, err, "Commit failed on setup with db")

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewEthermintApp(
		log.NewNopLogger(), db, nil, true, map[int64]bool{}, DefaultNodeHome, simtestutil.EmptyAppOptions{})
	_, err = app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}

// TestSDK53ModuleWiring exercises the required pre-block and IBC routes against
// initialized genesis, so missing light-client registration cannot go unnoticed.
func TestSDK53ModuleWiring(t *testing.T) {
	app := Setup(false, nil)
	require.Equal(t, []string{upgradetypes.ModuleName, authtypes.ModuleName}, app.mm.OrderPreBlockers)
	ctx := app.NewContextLegacy(false, tmproto.Header{Height: 1})
	for _, clientID := range []string{"07-tendermint-0", "06-solomachine-0"} {
		client, err := app.IBCKeeper.ClientKeeper.Route(ctx, clientID)
		require.NoError(t, err)
		require.NotNil(t, client)
	}
	require.True(t, app.IBCKeeper.PortKeeper.Router.HasRoute(transfertypes.ModuleName))
	require.True(t, app.IBCKeeper.PortKeeper.Router.Sealed())
	require.False(t, app.AccountKeeper.UnorderedTransactionsEnabled())
}
