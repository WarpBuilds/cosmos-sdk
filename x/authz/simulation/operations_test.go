package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz/testutil"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app               *runtime.App
	legacyAmino       *codec.LegacyAmino
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	authzKeeper       authzkeeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	app, err := simtestutil.Setup(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&suite.legacyAmino,
		&suite.codec,
		&suite.interfaceRegistry,
		&suite.txConfig,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.authzKeeper,
	)
	suite.Require().NoError(err)
	suite.app = app
	suite.ctx = app.NewContext(false)
}

func (suite *SimTestSuite) TestWeightedOperations() {
	cdc := suite.codec
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(suite.interfaceRegistry, appParams, cdc, suite.txConfig, suite.accountKeeper,
		suite.bankKeeper, suite.authzKeeper)

	s := rand.NewSource(3)
	r := rand.New(s)
	// setup 2 accounts
	accs := suite.getTestingAccounts(r, 2)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.WeightGrant, authz.ModuleName, simulation.TypeMsgGrant},
		{simulation.WeightExec, authz.ModuleName, simulation.TypeMsgExec},
		{simulation.WeightRevoke, authz.ModuleName, simulation.TypeMsgRevoke},
	}

	require := suite.Require()
	for i, w := range weightedOps {
		op, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, "")
		require.NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(expected[i].weight, w.Weight(), "weight should be the same. %v", op.Comment)
		require.Equal(expected[i].opMsgRoute, op.Route, "route should be the same. %v", op.Comment)
		require.Equal(expected[i].opMsgName, op.Name, "operation Msg name should be the same %v", op.Comment)
	}
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) TestSimulateGrant() {
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithBlockTime(blockTime)

	_, err := suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	granter := accounts[0]
	grantee := accounts[1]

	// execute operation
	op := simulation.SimulateMsgGrant(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.authzKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgGrant
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(granter.Address.String(), msg.Granter)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateRevoke() {
	// setup 3 accounts
	s := rand.NewSource(2)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	_, err := suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	require.NoError(suite.T(), err)

	initAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	granter := accounts[0]
	grantee := accounts[1]
	a := banktypes.NewSendAuthorization(initCoins, nil)
	expire := time.Now().Add(30 * time.Hour)

	err = suite.authzKeeper.SaveGrant(suite.ctx, grantee.Address, granter.Address, a, &expire)
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgRevoke(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.authzKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgRevoke
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(granter.Address.String(), msg.Granter)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Equal(banktypes.SendAuthorization{}.MsgTypeURL(), msg.MsgTypeUrl)
	suite.Require().Len(futureOperations, 0)
}

func (suite *SimTestSuite) TestSimulateExec() {
	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 3)

	_, err := suite.app.FinalizeBlock(&abci.FinalizeBlockRequest{Height: suite.app.LastBlockHeight() + 1, Hash: suite.app.LastCommitID().Hash})
	suite.Require().NoError(err)

	initAmt := sdk.TokensFromConsensusPower(200000, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	granter := accounts[0]
	grantee := accounts[1]
	a := banktypes.NewSendAuthorization(initCoins, nil)
	expire := suite.ctx.BlockTime().Add(1 * time.Hour)

	err = suite.authzKeeper.SaveGrant(suite.ctx, grantee.Address, granter.Address, a, &expire)
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgExec(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.authzKeeper, suite.codec)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, suite.ctx, accounts, "")
	suite.Require().NoError(err)

	var msg authz.MsgExec
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(grantee.Address.String(), msg.Grantee)
	suite.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
