package handlers

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/rpc"
	"github.com/gobitfly/eth2-beaconchain-explorer/services"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/gobitfly/eth2-beaconchain-explorer/utils"
	"github.com/shopspring/decimal"
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

	chainIDBig := new(big.Int).SetUint64(utils.Config.Chain.ClConfig.DepositChainID)
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
	address := common.FromHex(strings.TrimPrefix(utils.Config.Chain.ClConfig.DepositContractAddress, "0x"))

	addressMetadata, err := db.BigtableClient.GetMetadataForAddress(address, 0, 1)
	if err != nil {
		logger.Errorf("error retieving balances for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)

		return
	}

	depositContractBalanceGWei := decimal.NewFromBigInt(new(big.Int).SetBytes(addressMetadata.EthBalance.Balance), 0).DivRound(decimal.NewFromInt(params.GWei), 18)

	totalSupply := (genesisTotalSupply + totalAmountWithdrawn + validatorParticipation.EligibleEther) - (uint64(depositContractBalanceGWei.InexactFloat64()) + uint64(latestBurnData.TotalBurned))
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
