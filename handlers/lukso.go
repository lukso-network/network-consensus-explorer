package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"net/http"
)

type luksoTotalSupplyResponse struct {
	TotalSupply uint64 `json:"total_supply"`
}

func LuksoTotalSupply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	genesisTotalSupply := uint64(42000000)

	totalAmountWithdrawn, _, err := db.GetTotalAmountWithdrawn()
	if err != nil {
		logger.WithError(err).Error("error getting total amount withdrawn from db")
	}

	totalSupply := genesisTotalSupply + totalAmountWithdrawn

	data := luksoTotalSupplyResponse{
		TotalSupply: totalSupply,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.WithError(err).WithField("route", r.URL.String()).Error("error encoding json response")
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

}
