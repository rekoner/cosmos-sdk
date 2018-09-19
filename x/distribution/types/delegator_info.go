package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// distribution info for a delegation
type DelegatorDistInfo struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`
	ValOperatorAddr  sdk.ValAddress `json:"val_operator_addr"`
	WithdrawalHeight int64          `json:"withdrawal_height"` // last time this delegation withdrew rewards
}

// withdraw rewards from delegator
func (di DelegatorDistInfo) WithdrawRewards(g Global, vi ValidatorDistInfo,
	height int64, totalBonded, vdTokens, totalDelShares, delegatorShares,
	commissionRate Dec) (di DelegatorDistInfo, g Global, withdrawn DecCoins) {

	vi.UpdateTotalDelAccum(height, totalDelShares)
	g = vi.TakeGlobalRewards(g, height, totalBonded, vdTokens, commissionRate)

	blocks = height - di.WithdrawalHeight
	di.WithdrawalHeight = height
	accum := delegatorShares.Mul(sdk.NewDec(blocks))
	withdrawalTokens := vi.Pool.Mul(accum.Quo(vi.TotalDelAccum))

	vi.Pool = vi.Pool.Sub(withdrawalTokens)
	vi.TotalDelAccum = vi.TotalDelAccum.sub(accum)

	return di, g, withdrawalTokens
}

//_____________________________________________________________________

// withdraw address for the delegation rewards
type DelegatorWithdrawInfo struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	WithdrawAddr  sdk.AccAddress `json:"withdraw_addr"`
}