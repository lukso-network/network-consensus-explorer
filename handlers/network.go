package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"net/http"
)

// Supply godoc
// @Summary Get total supply of all native network tokens in existence
// @Tags Supply
// @Description Returns information about  total supply of all native network tokens in existence (in wei).
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.SupplyResponse} "Success"
// @Failure 400 {object} types.ApiResponse "Failure"
// @Failure 500 {object} types.ApiResponse "Server Error"
// @Router /api/v2/totalsupply [get]
func Supply(w http.ResponseWriter, r *http.Request, cfg *types.Config) {
	w.Header().Set("Content-Type", "application/json")

	genesisTotalSupply := cfg.Chain.GenesisTotalSupply

	totalAmountWithdrawn, _, err := db.GetTotalAmountWithdrawn()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn from db")
	}

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn from db")
	}

	//TODO: provide validators or custom GetValidatorsTotalBalance
	latestBalances, err := db.BigtableClient.GetValidatorBalanceHistory(validators, latestFinalizedEpoch, latestFinalizedEpoch)
	if err != nil {
		logger.Errorf("error getting validator balance data in GetValidatorEarnings: %v", err)
	}

	balancesMap := make(map[uint64]*types.Validator, 0)
	totalBalance := uint64(0)

	for balanceIndex, balance := range latestBalances {
		if len(balance) == 0 {
			continue
		}

		if balancesMap[balanceIndex] == nil {
			balancesMap[balanceIndex] = &types.Validator{}
		}
		balancesMap[balanceIndex].Balance = balance[0].Balance
		balancesMap[balanceIndex].EffectiveBalance = balance[0].EffectiveBalance

		totalBalance += balance[0].Balance
	}

	totalSupply := genesisTotalSupply + totalAmountWithdrawn

	data := types.SupplyResponse{
		TotalSupply: totalSupply,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}
