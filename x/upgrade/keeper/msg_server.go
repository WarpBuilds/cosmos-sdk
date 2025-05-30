package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the upgrade MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

var (
	_    types.MsgServer = msgServer{}
	_, _ sdk.Msg         = &types.MsgSoftwareUpgrade{}, &types.MsgCancelUpgrade{}
)

// SoftwareUpgrade implements the Msg/SoftwareUpgrade Msg service.
func (k msgServer) SoftwareUpgrade(goCtx context.Context, msg *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "expected %s got %s", k.authority, msg.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	err := k.ScheduleUpgrade(ctx, msg.Plan)
	if err != nil {
		return nil, err
	}

	return &types.MsgSoftwareUpgradeResponse{}, nil
}

// CancelUpgrade implements the Msg/CancelUpgrade Msg service.
func (k msgServer) CancelUpgrade(ctx context.Context, msg *types.MsgCancelUpgrade) (*types.MsgCancelUpgradeResponse, error) {
	if k.authority != msg.Authority {
		return nil, errors.Wrapf(types.ErrInvalidSigner, "expected %s got %s", k.authority, msg.Authority)
	}

	err := k.ClearUpgradePlan(ctx)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelUpgradeResponse{}, nil
}
