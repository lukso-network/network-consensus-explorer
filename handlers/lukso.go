package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"net/http"
)

type luksoSupplyResponse struct {
	CirculatingSupply uint64 `json:"circulating_supply"`
	TotalSupply       uint64 `json:"total_supply"`
}

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

	data := luksoSupplyResponse{
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
