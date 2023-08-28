package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"math/big"
	"net/http"
)

// Supply godoc
// @Summary Get total supply of all native network tokens in existence
// @Tags Misc
// @Description Returns information about total supply of all native network tokens in existence (in wei).
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.SupplyResponse} "Success"
// @Failure 400 {object} types.ApiResponse "Failure"
// @Failure 500 {object} types.ApiResponse "Server Error"
// @Router /api/v2/totalsupply [get]
func Supply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	genesisTotalSupply := utils.Config.Chain.GenesisTotalSupply

	totalAmountWithdrawn, _, err := db.GetTotalAmountWithdrawn()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn from db")
	}

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if err != nil {
		logger.WithError(err).Error("error getting LatestFinalizedEpoch")
	}

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient("http://"+utils.Config.Indexer.Node.Host+":"+utils.Config.Indexer.Node.Port, chainIDBig)
	if err != nil {
		logger.WithError(err).Error("new bigtable lighthouse client in monitor error", 0)
	}

	validatorParticipation, err := rpcClient.GetValidatorParticipation(latestFinalizedEpoch)

	balanceStatistics, err := db.BigtableClient.GetValidatorBalanceStatistics(latestFinalizedEpoch, latestFinalizedEpoch)

	totalValidatorsBalance := uint64(0)
	for _, statistic := range balanceStatistics {
		totalValidatorsBalance += statistic.EndBalance
	}

	logger.Infof("exported validators statistics for finalized epoch %v (balanceStatistics len: %v, total balance: %v)", latestFinalizedEpoch, len(balanceStatistics), totalValidatorsBalance)

	totalSupply := genesisTotalSupply + totalAmountWithdrawn + validatorParticipation.EligibleEther

	data := types.SupplyResponse{
		TotalSupply: totalSupply,
	}

	response := &types.ApiResponse{
		Status: "OK",
		Data:   data,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}
