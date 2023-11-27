package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/rpc"
	"eth2-exporter/services"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"net/http"
	"strings"
)

const (
	GWei = 1e9
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
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	latestFinalizedEpoch := services.LatestFinalizedEpoch()
	if err != nil {
		logger.WithError(err).Error("error getting LatestFinalizedEpoch")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.Config.DepositChainID)
	rpcClient, err := rpc.NewLighthouseClient("http://"+utils.Config.Indexer.Node.Host+":"+utils.Config.Indexer.Node.Port, chainIDBig)
	if err != nil {
		logger.WithError(err).Error("new total supply Lighthouse client in monitor error")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	// Get the total staked gwei that was active (i.e., able to vote) during the latestFinalizedEpoch epoch
	validatorParticipation, err := rpcClient.GetValidatorParticipation(latestFinalizedEpoch)
	if err != nil {
		logger.WithError(err).Error("error getting GetValidatorParticipation")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	latestBurnData := services.LatestBurnData()
	address := common.FromHex(strings.TrimPrefix(utils.Config.Chain.Config.DepositContractAddress, "0x"))
	addressMetadata, err := db.BigtableClient.GetMetadataForAddress(address)

	if err != nil {
		logger.Errorf("error retieving balances for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	depositContractBalance := uint64(utils.WeiBytesToEther(addressMetadata.EthBalance.Balance).InexactFloat64())
	depositContractBalanceAmount := (new(big.Int).Mul(new(big.Int).SetUint64(depositContractBalance), big.NewInt(params.Ether))).Uint64()

	totalSupply := (genesisTotalSupply + totalAmountWithdrawn + validatorParticipation.EligibleEther) - depositContractBalanceAmount - uint64(latestBurnData.TotalBurned)
	amount := new(big.Int).Mul(new(big.Int).SetUint64(totalSupply), big.NewInt(params.GWei))

	data := types.SupplyResponse{
		TotalSupply: amount.String(),
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