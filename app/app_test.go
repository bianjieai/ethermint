package app

import (
	"testing"

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
