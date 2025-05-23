package module

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
)

func EndBlocker(ctx context.Context, k keeper.Keeper) error {
	// 200 is an arbitrary value, we can change it later if needed
	return k.RemoveExpiredAllowances(ctx, 200)
}
