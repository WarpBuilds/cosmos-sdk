package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	ctx := sdk.NewContext(nil, cmtproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	assert.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	assert.Equal(t, int64(905), msgUpdateParams.Params.SignedBlocksWindow)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(7, 2), msgUpdateParams.Params.MinSignedPerWindow)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(60, 2), msgUpdateParams.Params.SlashFractionDoubleSign)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(89, 2), msgUpdateParams.Params.SlashFractionDowntime)
	assert.Equal(t, 3313479009*time.Second, msgUpdateParams.Params.DowntimeJailDuration)
}
