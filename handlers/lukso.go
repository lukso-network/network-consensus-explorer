package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"net/http"
)

// LuksoSupply godoc
// @Summary Get circulating supply and total supply for LUKSO network
// @Tags LUKSO
// @Description Returns information about circulating supply and total supply for LUKSO network
// @Produce  json
// @Success 200 {object} types.ApiResponse{data=types.LUKSOSupplyResponse} "Success"
// @Failure 400 {object} types.ApiResponse "Failure"
// @Failure 500 {object} types.ApiResponse "Server Error"
// @Router /api/v1/lukso/supply [get]
func LuksoSupply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// From https://github.com/lukso-network/network-configs/blob/main/mainnet/shared/genesis.json
	genesisTotalSupply := uint64(42000000)
	genesisFoundationSupply := uint64(11143518)

	totalAmountWithdrawn, _, err := db.GetTotalAmountWithdrawn()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn from db")
	}

	circulatingSupply := (genesisTotalSupply - genesisFoundationSupply) + totalAmountWithdrawn
	totalSupply := genesisTotalSupply + totalAmountWithdrawn

	data := types.LUKSOSupplyResponse{
		CirculatingSupply: circulatingSupply,
		TotalSupply:       totalSupply,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}
